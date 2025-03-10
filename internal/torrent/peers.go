package torrent

import "net"

type PeerSet struct {
	state             int
	peerSet           map[peer]bool
	downloadRateLimit int //in bytes
	//others as required
}
type peer struct {
	address          string
	lastConnected    string
	isConnected      string
	connection       net.TCPConn
	bytesTransferred int
	retries          int
}
