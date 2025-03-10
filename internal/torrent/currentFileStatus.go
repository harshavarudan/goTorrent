package torrent

type FileStatusMetadata struct {
	infoHash   [20]byte
	downloaded uint64
	left       uint64
}
