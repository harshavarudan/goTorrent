package torrent

type FileStatusMetadata struct { //store this as a file
	downloaded         uint64
	left               uint64
	downloadedFilePath string
}
