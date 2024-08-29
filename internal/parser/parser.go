package parser

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/zeebo/bencode"
)

type TorrentFile struct {
	Announce     string      `bencode:"announce"`
	AnnounceList [][]string  `bencode:"announce-list"`
	Info         TorrentInfo `bencode:"info"`
	Comment      string      `bencode:"comment"`
	CreatedBy    string      `bencode:"created by"`
	CreationDate int64       `bencode:"creation date"`
	Encoding     string      `bencode:"encoding"`
}

// TorrentInfo represents the "info" dictionary within the torrent file
type TorrentInfo struct {
	Name        string `bencode:"name"`
	Length      int64  `bencode:"length"`
	Pieces      []byte `bencode:"pieces"`
	PieceLength int64  `bencode:"piece length"`
}

func ParseTorrentFile(filePath string) (TorrentFile, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return TorrentFile{}, err
	}
	torrent := TorrentFile{}
	if err := bencode.DecodeBytes(fileContent, &torrent); err != nil {
		fmt.Println("Error decoding torrent file:", err)
		return TorrentFile{}, err
	}

	fmt.Printf("Announce: %s\n", torrent.Announce)
	fmt.Printf("Announce List: %v\n", torrent.AnnounceList)

	fmt.Printf("Name: %s\n", torrent.Info.Name)
	fmt.Printf("Length: %d\n", torrent.Info.Length)
	fmt.Printf("Piece Length: %d\n", torrent.Info.PieceLength)
	fmt.Printf("Pieces: %s\n", hex.EncodeToString(torrent.Info.Pieces))
	return torrent, err
}
