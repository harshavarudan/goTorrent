package parser

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/zeebo/bencode"

	"github.com/harshavarudan/goTorrent/internal/torrent"
)

func ParseTorrentFile(filePath string, metadata *torrent.MetaDataInfo) error {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	if err := bencode.DecodeBytes(fileContent, metadata); err != nil {
		fmt.Println("Error decoding torrent file:", err)
		return err
	}

	fmt.Printf("Announce: %s\n", metadata.Announce)
	fmt.Printf("Announce List: %v\n", metadata.AnnounceList)

	fmt.Printf("Name: %s\n", metadata.Info.Name)
	fmt.Printf("Length: %d\n", metadata.Info.Length)
	fmt.Printf("Piece Length: %d\n", metadata.Info.PieceLength)
	fmt.Printf("Pieces: %s\n", hex.EncodeToString(metadata.Info.Pieces))
	return err
}
