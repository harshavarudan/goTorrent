package torrent

import (
	"fmt"
	"os"

	"github.com/zeebo/bencode"
)

func ParseTorrentFile(filePath string, metadata *MetaDataInfo) error {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	if err := bencode.DecodeBytes(fileContent, metadata); err != nil {
		fmt.Println("Error decoding torrent file:", err)
		return err
	}
	err = metadata.ComputeInfoHash()
	if err != nil {
		return err
	}
	metadata.PrintContents()
	return err
}
