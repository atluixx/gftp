package protocol

type Chunk struct {
	Index    uint32
	FileSize int64
	Data     []byte
}
