package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/atluixx/gftp/internal/protocol"
	"github.com/atluixx/gftp/internal/transfer"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "send":
		err = runSend(os.Args[2:])
	case "receive":
		err = runReceive(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "gftp:", err)
		os.Exit(1)
	}
}
func usage() {
	fmt.Fprintln(os.Stderr, "Usage:\n  gftp send -file <path> [-file <path> ...] [-host localhost] [-port 8080] [-transport tcp|udp] [-token TOKEN] [-tls]\n  gftp receive [-output <directory/[filename]>] [-host 0.0.0.0] [-port 8080] [-transport tcp|udp] [-token TOKEN] [-tls-cert CERT -tls-key KEY]")
}

type files []string

func (f *files) String() string             { return strings.Join(*f, ",") }
func (f *files) Set(v string) error         { *f = append(*f, v); return nil }
func endpoint(host string, port int) string { return net.JoinHostPort(host, fmt.Sprint(port)) }
func runSend(args []string) error {
	fs := flag.NewFlagSet("send", flag.ContinueOnError)
	var paths files
	fs.Var(&paths, "file", "file or directory (repeatable)")
	host := fs.String("host", "localhost", "receiver host")
	port := fs.Int("port", 8080, "receiver port")
	token := fs.String("token", "", "shared authentication token")
	transportName := fs.String("transport", "tcp", "transport: tcp or udp")
	useTLS := fs.Bool("tls", false, "encrypt the connection")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(paths) == 0 {
		return errors.New("at least one -file is required")
	}
	var c net.Conn
	var err error
	if *transportName != "tcp" && *transportName != "udp" {
		return errors.New("-transport must be tcp or udp")
	}
	if *transportName == "udp" && *useTLS {
		return errors.New("TLS is only available with TCP")
	}
	if *useTLS {
		c, err = tls.Dial("tcp", endpoint(*host, *port), &tls.Config{ServerName: *host, InsecureSkipVerify: true})
	} else {
		c, err = net.Dial(*transportName, endpoint(*host, *port))
	}
	if err != nil {
		return err
	}
	defer c.Close()
	if err = transfer.Authenticate(c, *token); err != nil {
		return err
	}
	return transfer.EncodePaths(paths, c, func(n string, s, t int64) {
		fmt.Fprintf(os.Stderr, "\rSending %s: %d/%d bytes", n, s, t)
		if s == t {
			fmt.Fprintln(os.Stderr)
		}
	})
}
func runReceive(args []string) error {
	fs := flag.NewFlagSet("receive", flag.ContinueOnError)
	out := fs.String("output", ".", "output directory, optionally followed by a filename")
	host := fs.String("host", "0.0.0.0", "listen host")
	port := fs.Int("port", 8080, "listen port")
	token := fs.String("token", "", "shared authentication token")
	transportName := fs.String("transport", "tcp", "transport: tcp or udp")
	certPath := fs.String("tls-cert", "", "TLS certificate PEM (requires -tls-key)")
	keyPath := fs.String("tls-key", "", "TLS private key PEM (requires -tls-cert)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *transportName != "tcp" && *transportName != "udp" {
		return errors.New("-transport must be tcp or udp")
	}
	if *transportName == "udp" && (*certPath != "" || *keyPath != "") {
		return errors.New("TLS is only available with TCP")
	}
	if (*certPath == "") != (*keyPath == "") {
		return errors.New("-tls-cert and -tls-key must be supplied together")
	}
	if *transportName == "udp" {
		return receiveUDP(endpoint(*host, *port), *out, *token)
	}
	l, err := net.Listen("tcp", endpoint(*host, *port))
	if err != nil {
		return err
	}
	if *certPath != "" {
		cert, err := tls.LoadX509KeyPair(*certPath, *keyPath)
		if err != nil {
			_ = l.Close()
			return err
		}
		l = tls.NewListener(l, &tls.Config{Certificates: []tls.Certificate{cert}})
	}
	defer l.Close()
	fmt.Println("Waiting for connection...")
	c, err := l.Accept()
	if err != nil {
		return err
	}
	defer c.Close()
	fmt.Println("Connected!")
	return receive(c, *out, *token)
}
func receiveUDP(address, output, token string) error {
	a, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return err
	}
	c, err := net.ListenUDP("udp", a)
	if err != nil {
		return err
	}
	defer c.Close()
	fmt.Println("Waiting for UDP datagrams...")
	return receive(&udpConn{conn: c}, output, token)
}
func packetError(c net.Conn, msg string) error {
	_ = transfer.WritePacket(c, protocol.Packet{Type: protocol.PacketError, PayloadSize: uint32(len(msg)), Payload: []byte(msg)})
	return errors.New(msg)
}
func receive(c net.Conn, requested, token string) error {
	if token != "" {
		p, err := transfer.ReadPacket(c)
		if err != nil {
			return err
		}
		if p.Type != protocol.PacketAuth || subtle.ConstantTimeCompare(p.Payload, []byte(token)) != 1 {
			return packetError(c, "authentication failed")
		}
		if err := transfer.WritePacket(c, protocol.Packet{Type: protocol.PacketAck}); err != nil {
			return err
		}
	}
	var f *os.File
	var part, final string
	var hash hashWriter
	var received, expected int64
	var next uint32
	defer func() {
		if f != nil {
			_ = f.Close()
		}
	}()
	for {
		p, err := transfer.ReadPacket(c)
		if err == io.EOF {
			return nil
		}
		if ne, ok := err.(net.Error); ok && ne.Timeout() {
			return nil
		}
		if err != nil {
			return err
		}
		switch p.Type {
		case protocol.PacketFileInfo:
			if f != nil {
				return packetError(c, "file metadata received during active transfer")
			}
			fi, err := protocol.DecodeFileInfo(p.Payload)
			if err != nil {
				return packetError(c, err.Error())
			}
			final, err = resolveOutputPath(requested, fi.Name)
			if err != nil {
				return packetError(c, err.Error())
			}
			if err = os.MkdirAll(filepath.Dir(final), 0755); err != nil {
				return err
			}
			part = final + ".part"
			offset := int64(0)
			if st, e := os.Stat(part); e == nil {
				offset = st.Size()
			}
			if offset > fi.FileSize {
				if err := os.Remove(part); err != nil {
					return err
				}
				offset = 0
			}
			f, err = os.OpenFile(part, os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			if _, err = f.Seek(offset, io.SeekStart); err != nil {
				return err
			}
			hash = sha256.New()
			if old, e := os.Open(part); e == nil {
				_, _ = io.CopyN(hash, old, offset)
				_ = old.Close()
			}
			received, expected, next = offset, fi.FileSize, uint32(offset/transfer.ChunkSize)
			b := make([]byte, 8)
			binary.BigEndian.PutUint64(b, uint64(offset))
			if err = transfer.WritePacket(c, protocol.Packet{Type: protocol.PacketAck, PayloadSize: 8, Payload: b}); err != nil {
				return err
			}
			fmt.Println("Receiving:", fi.Name)
		case protocol.PacketChunk:
			if f == nil {
				return packetError(c, "chunk received before file info")
			}
			if p.Index != next || p.FileSize != expected || received+int64(len(p.Payload)) > expected {
				return packetError(c, "invalid chunk sequence")
			}
			n, e := f.Write(p.Payload)
			if e != nil {
				return e
			}
			if n != len(p.Payload) {
				return io.ErrShortWrite
			}
			_, _ = hash.Write(p.Payload)
			received += int64(n)
			next++
			fmt.Fprintf(os.Stderr, "\rReceiving: %d/%d bytes", received, expected)
		case protocol.PacketFileEnd:
			if f == nil {
				return packetError(c, "end received before file info")
			}
			if received != expected {
				return packetError(c, "incomplete transfer")
			}
			sum := hash.Sum(nil)
			if len(p.Payload) != sha256.Size || subtle.ConstantTimeCompare(sum, p.Payload) != 1 {
				return packetError(c, "integrity check failed")
			}
			if err := f.Close(); err != nil {
				return err
			}
			f = nil
			if err := os.Rename(part, final); err != nil {
				return err
			}
			fmt.Fprintln(os.Stderr, "\nTransfer finished:", final)
			if err := transfer.WritePacket(c, protocol.Packet{Type: protocol.PacketAck}); err != nil {
				return err
			}
			// UDP has no EOF; an idle period ends its otherwise connectionless session.
			if _, ok := c.(*udpConn); ok {
				_ = c.SetReadDeadline(time.Now().Add(time.Second))
			}
		default:
			return packetError(c, "unexpected packet type")
		}
	}
}

type hashWriter interface {
	Write([]byte) (int, error)
	Sum([]byte) []byte
}

func resolveOutputPath(requested, name string) (string, error) {
	if requested == "" {
		return "", errors.New("output path cannot be empty")
	}
	clean := filepath.Clean(filepath.FromSlash(name))
	if filepath.IsAbs(clean) || clean == "." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) || clean == ".." {
		return "", errors.New("unsafe received filename")
	}
	dir := strings.HasSuffix(requested, string(os.PathSeparator))
	if st, e := os.Stat(requested); e == nil {
		dir = st.IsDir()
	} else if !os.IsNotExist(e) {
		return "", e
	}
	if dir {
		return filepath.Join(requested, clean), nil
	}
	return requested, nil
}
