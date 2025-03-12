package torrent

type Torrent struct {
	torrentFilePath string
	tracker         *TrackerSet
	//downloader *downloader
	//uploader *uploader
	metaInfo          *MetaDataInfo
	PeerSet           *PeerSet
	currentFileStatus *FileStatusMetadata
	//get default settings or store settings
	//optional: pick where its left off in case server crashes
}

func NewTorrent(filePath string) Torrent {

	return Torrent{
		torrentFilePath: filePath,
		tracker: &TrackerSet{
			conn:        nil,
			state:       0,
			trackerSet:  map[string]Tracker{},
			workerCount: 10, //default
		},
		metaInfo:          &MetaDataInfo{},
		PeerSet:           &PeerSet{},
		currentFileStatus: &FileStatusMetadata{},
	}
}

// parse the file
func (t Torrent) ParseFile() error {
	return ParseTorrentFile(t.torrentFilePath, t.metaInfo)

}

func (t Torrent) CreateTrackerSet() error {
	return t.tracker.Init(t.metaInfo)
}
