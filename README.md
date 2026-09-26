# gftp

A file transfer protocol built from scratch in Go.

## Install

```sh
git clone https://github.com/atluixx/gftp.git
cd gftp
go build ./cmd/sender
go build ./cmd/receiver
```

## Usage

Start the receiver:

```sh
go run ./cmd/receiver -output ./output
```

Then send a file:

```sh
go run ./cmd/sender -file ./path/to/file
```

The receiver saves the file in the directory specified by `-output`.

### Example

```sh
go run ./cmd/receiver -output ./output
```

```sh
go run ./cmd/sender -file ./tests/test.pdf
```

## Roadmap

* [x] Basic TCP file transfer
* [x] File chunking
* [x] Custom packet protocol
* [x] File metadata
* [x] Transfer completion packet
* [x] Sender file flag
* [x] Receiver output flag
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
