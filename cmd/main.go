package main

import (
	"fmt"

	"github.com/harshavarudan/goTorrent/internal/torrent"
)

func main() {
	s := "internal/torrent/parser/Hotshots.torrent"
	t := torrent.NewTorrent(s)
	err := t.ParseFile()
	err = t.CreateTracker()
	if err != nil {
		return
	}

}
func test() {
	fmt.Println("testing")
}
