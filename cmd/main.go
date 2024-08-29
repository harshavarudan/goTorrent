package main

import (
	"fmt"

	"github.com/harshavarudan/goTorrent/internal/parser"
	"github.com/harshavarudan/goTorrent/internal/tracker"
)

func main() {
	// code
	fmt.Println("Hello, World!")
	file, err := parser.ParseTorrentFile("internal/parser/Hotshots.torrent")
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
	t.Init(trackerList...)

}
