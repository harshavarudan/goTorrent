package torrent

import (
	"github.com/harshavarudan/goTorrent/internal/torrent/parser"
	"github.com/harshavarudan/goTorrent/internal/torrent/tracker"
)

type Torrent struct {
	tracker *tracker.Tracker
	//downloader *downloader
	//uploader *uploader
	metaInfo *parser.MetaInfo
	peerSet  map[string]bool //might change type
}
