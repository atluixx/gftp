package main

import (
	"net"
	"time"
)

// udpConn presents datagrams as a buffered byte stream. Packets are capped
// below the UDP payload limit by the protocol, so one gftp packet remains one
// datagram while ReadFull can still consume its header and payload separately.
type udpConn struct {
	conn    *net.UDPConn
	peer    *net.UDPAddr
	pending []byte
}

func (c *udpConn) Read(b []byte) (int, error) {
	if len(c.pending) == 0 {
		buf := make([]byte, 65535)
		for {
			n, peer, err := c.conn.ReadFromUDP(buf)
			if err != nil {
				return 0, err
			}
			if c.peer == nil {
				c.peer = peer
			}
			if peer.IP.Equal(c.peer.IP) && peer.Port == c.peer.Port {
				c.pending = append(c.pending, buf[:n]...)
				break
			}
		}
	}
	n := copy(b, c.pending)
	c.pending = c.pending[n:]
	return n, nil
}
func (c *udpConn) Write(b []byte) (int, error)        { return c.conn.WriteToUDP(b, c.peer) }
func (c *udpConn) Close() error                       { return c.conn.Close() }
func (c *udpConn) LocalAddr() net.Addr                { return c.conn.LocalAddr() }
func (c *udpConn) RemoteAddr() net.Addr               { return c.peer }
func (c *udpConn) SetDeadline(t time.Time) error      { return c.conn.SetDeadline(t) }
func (c *udpConn) SetReadDeadline(t time.Time) error  { return c.conn.SetReadDeadline(t) }
func (c *udpConn) SetWriteDeadline(t time.Time) error { return c.conn.SetWriteDeadline(t) }
