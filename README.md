# gftp

A file transfer protocol built from scratch in Go.

## Install

```sh
git clone https://github.com/atluixx/gftp.git
cd gftp
go build ./cmd/gftp
```

## Usage

Start the receiver:

```sh
./gftp receive -output ./output/
```

Then send a file:

```sh
./gftp send -file ./path/to/file
```

End `-output` with a slash to keep the transmitted filename, or include a
filename to rename it:

```sh
./gftp receive -output ./output/renamed-file.pdf
```

Both commands accept `-host` and `-port`. `send` may repeat `-file`; a file
or directory is sent recursively, and the receiver preserves directory
structure. Transfers show byte progress, are acknowledged per file, and use
SHA-256 verification. Interrupted downloads are retained as `.part` files and
resume automatically on the next transfer.

Use a shared `-token` on both ends to require authentication. TLS encryption
is available by passing `-tls-cert cert.pem -tls-key key.pem` to the receiver
and `-tls` to the sender. The sender deliberately accepts the server
certificate without verification, so use this only with a certificate you
control on a trusted network (or add certificate verification before exposing
the service publicly).

TCP is the default transport. Use `-transport udp` on both commands for UDP;
the same protocol checks reject missing, reordered, or corrupted packets.

### Example

```sh
./gftp receive -output ./output/
```

```sh
./gftp send -file ./tests/test.pdf
```

## Roadmap

* [x] Basic TCP file transfer
* [x] File chunking
* [x] Custom packet protocol
* [x] File metadata
* [x] Transfer completion packet
* [x] Sender file flag
* [x] Receiver output flag
* [x] Single binary for sending and receiving
* [x] Configurable host and port
* [x] Transfer progress
* [x] File integrity verification
* [x] Better error handling
* [x] Transfer acknowledgements
* [x] Resumable transfers
* [x] Multiple file transfers
* [x] Directory transfers
* [x] Authentication
* [x] Encryption
* [x] UDP transport

## Contributing

PRs are welcome.

## License

MIT © atluixx
