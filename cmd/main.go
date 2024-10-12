package main

import (
	"crypto/sha1"
	"fmt"
	"log"

	"github.com/zeebo/bencode"

	"github.com/harshavarudan/goTorrent/internal/torrent/parser"
	"github.com/harshavarudan/goTorrent/internal/torrent/tracker"
)

func main() {
	// code
	fmt.Println("Hello, World!")
	file, err := parser.ParseTorrentFile("internal/torrent/parser/Hotshots.torrent")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(file.Announce)
	t := tracker.Tracker{}
	trackerList := make([]string, 0)
	for _, val := range file.AnnounceList {
		trackerList = append(trackerList, val...)
	}
	infoBytes, err := bencode.EncodeBytes(file.Info)
	if err != nil {
		log.Fatalf("failed to re-encode info dictionary: %v", err)
	}

	// Compute the SHA-1 hash of the bencoded Info dictionary
	infoHash := sha1.Sum(infoBytes)

	fmt.Printf("Info Hash: %x\n", infoHash)

	t.Init(infoHash, trackerList...)

	//int

}
