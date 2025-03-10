package torrent

type Torrent struct {
	tracker *TrackerSet
	//downloader *downloader
	//uploader *uploader
	metaInfo *MetaDataInfo
	PeerSet  *PeerSet
}
