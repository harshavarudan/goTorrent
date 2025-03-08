package main

import (
	"crypto/sha1"
	"fmt"
	"log"
	"time"

	"github.com/zeebo/bencode"

	"github.com/harshavarudan/goTorrent/internal/torrent/parser"
	"github.com/harshavarudan/goTorrent/internal/torrent/tracker"
	"github.com/harshavarudan/goTorrent/internal/worker"
)

func main() {
	// code
	d := worker.NewDispatcher(20, test)
	d.Run()
	d.AddWorker()
	go func() {
		for i := 0; i < 25; i++ {
			time.Sleep(1 * time.Second)
			go d.SendSignal(i)
		}
	}()

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
func test() {
	fmt.Println("testing")
}
