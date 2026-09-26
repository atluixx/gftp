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
* [ ] Configurable host and port
* [ ] Transfer progress
* [ ] File integrity verification
* [ ] Better error handling
* [ ] Transfer acknowledgements
* [ ] Resumable transfers
* [ ] Multiple file transfers
* [ ] Directory transfers
* [ ] Authentication
* [ ] Encryption
* [ ] UDP transport

## Contributing

PRs are welcome.

## License

MIT © atluixx
