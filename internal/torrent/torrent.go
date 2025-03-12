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
		tracker:         NewTrackerSet(10),
		metaInfo:        &MetaDataInfo{},
		PeerSet:         NewPeerSet(),
	}
}

// ParseFile parse the file
func (t *Torrent) ParseFile() error {
	err := ParseTorrentFile(t.torrentFilePath, t.metaInfo)
	if err != nil {
		return err
	}
	return nil
}
func (t *Torrent) GetFileInfoStatus() {
	fm := t.GetCurrentFileStatusFromFile()
	if fm != nil {
		t.currentFileStatus = fm
		return
	} else { //download from scratch
		t.currentFileStatus = &FileStatusMetadata{
			downloaded: 0,
			uploaded:   0,
			left:       t.metaInfo.Info.Length,
			isComplete: false,
		}
	}
	//TODO  optional parse the file which gives file downloaded so far
}
func (t *Torrent) GetCurrentFileStatusFromFile() *FileStatusMetadata {
	return nil
}

func (t *Torrent) CreateTracker() error {
	return t.tracker.Init(t.metaInfo)
}
func (t *Torrent) StartTracker() {
	t.tracker.Start(t.PeerSet, t.metaInfo, t.currentFileStatus)
}
