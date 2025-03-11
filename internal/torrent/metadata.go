package torrent

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"

	"github.com/zeebo/bencode"
)

// MetaDataInfo represents the overall metadata of a torrent file.
// MetaDataInfo represents the overall metadata of a torrent file.
type MetaDataInfo struct {
	Announce     string      `bencode:"announce"`
	AnnounceList [][]string  `bencode:"announce-list,omitempty"`
	Info         TorrentInfo `bencode:"info"`
	Comment      string      `bencode:"comment,omitempty"`
	CreatedBy    string      `bencode:"created by,omitempty"`
	CreationDate int64       `bencode:"creation date,omitempty"`
	Encoding     string      `bencode:"encoding,omitempty"`
	InfoHash     [20]byte    // Computed from the bencoded "info" dictionary
}

// TorrentInfo represents the "info" dictionary within the torrent file.
type TorrentInfo struct {
	Name        string     `bencode:"name"`
	Length      int64      `bencode:"length,omitempty"` // Single-file torrents
	Files       []FileInfo `bencode:"files,omitempty"`  // Multi-file torrents
	Pieces      []byte     `bencode:"pieces"`
	PieceLength int64      `bencode:"piece length"`
}

// FileInfo represents individual files within a multi-file torrent.
type FileInfo struct {
	Length int64    `bencode:"length"`
	Path   []string `bencode:"path"`
}

// ComputeInfoHash generates the SHA-1 hash of the bencoded "info" dictionary.
func (m *MetaDataInfo) ComputeInfoHash() error {
	encodedInfo, err := bencode.EncodeBytes(m.Info)
	if err != nil {
		return err
	}
	hash := sha1.Sum(encodedInfo)
	m.InfoHash = hash
	return nil
}

// GetTotalSize calculates the total size of the torrent in bytes.
func (m *MetaDataInfo) GetTotalSize() int64 {
	if len(m.Info.Files) > 0 {
		var totalSize int64
		for _, file := range m.Info.Files {
			totalSize += file.Length
		}
		return totalSize
	}
	return m.Info.Length
}

// PrintContents prints the decoded contents of a torrent file.
func (m *MetaDataInfo) PrintContents() {
	fmt.Println("Announce URL:", m.Announce)
	if len(m.AnnounceList) > 0 {
		fmt.Println("Announce List:", m.AnnounceList)
	}
	fmt.Println("Comment:", m.Comment)
	fmt.Println("Created By:", m.CreatedBy)
	fmt.Println("Creation Date:", m.CreationDate)
	fmt.Println("Encoding:", m.Encoding)
	fmt.Println("Info Hash:", hex.EncodeToString(m.InfoHash[:]))

	fmt.Println("\nTorrent Info:")
	fmt.Println("Name:", m.Info.Name)
	fmt.Println("Piece Length:", m.Info.PieceLength/(1024*1024), "MB")
	fmt.Println("Number of Pieces:", len(m.Info.Pieces)/20) // Divide by 20 (each SHA-1 hash is 20 bytes)

	totalSize := m.GetTotalSize()
	fmt.Println("Total Size:", totalSize/1024, "KB")

	if len(m.Info.Files) > 0 {
		fmt.Println("\nMulti-File Torrent:")
		for _, file := range m.Info.Files {
			fmt.Printf(" - Path: %v | Size: %d KB\n", file.Path, file.Length/1024)
		}
	} else {
		fmt.Println("\nSingle-File Torrent:")
		fmt.Println("File Size:", m.Info.Length/1024, "KB")
	}
}
