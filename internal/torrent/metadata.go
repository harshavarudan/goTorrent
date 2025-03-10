package torrent

type MetaDataInfo struct {
	Announce     string      `bencode:"announce"`
	AnnounceList [][]string  `bencode:"announce-list"`
	Info         TorrentInfo `bencode:"info"`
	Comment      string      `bencode:"comment"`
	CreatedBy    string      `bencode:"created by"`
	CreationDate int64       `bencode:"creation date"`
	Encoding     string      `bencode:"encoding"`
	InfoHash     [20]byte
}

// TorrentInfo represents the "info" dictionary within the torrent file
type TorrentInfo struct {
	Name        string `bencode:"name"`
	Length      int64  `bencode:"length"`
	Pieces      []byte `bencode:"pieces"`
	PieceLength int64  `bencode:"piece length"`
}
