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
	go func() {
		for {
			time.Sleep(30 * time.Second)
			t.ConnectPeers()
		}
	}()
	t.ConnectPeers()
	time.Sleep(1000 * time.Second)

}
func test() {
	fmt.Println("testing")
}
