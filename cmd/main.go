package main

import (
	"fmt"
	"time"

	"github.com/harshavarudan/goTorrent/internal/torrent"
)

func main() {
	s := "internal/torrent/parser/Hotshots.torrent"
	t := torrent.NewTorrent(s)
	err := t.ParseFile()
	t.GetFileInfoStatus()
	err = t.CreateTracker()
	if err != nil {
		print(err)
	}
	t.StartTracker()
	time.Sleep(100 * time.Second)

}
func test() {
	fmt.Println("testing")
}
