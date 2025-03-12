package torrent

import "sync"

type FileStatusMetadata struct { //store this as a file
	downloaded         int64
	left               int64
	uploaded           int64
	isComplete         bool
	mu                 sync.RWMutex
	downloadedFilePath string
}
