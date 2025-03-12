package torrent

import (
	"net"
	"sync"
	"time"
)

type PeerSet struct {
	state             int
	peerSet           map[string]peer
	downloadRateLimit int //in kilo bytes
	mu                sync.RWMutex
	//others as required
}
type peer struct {
	IPAddress  string
	connection net.TCPConn
	tcpPort    int
	state      int

	isConnected      bool
	lastConnected    time.Time
	bytesTransferred int
	retries          int
}

func (ps *PeerSet) AddPeer(p peer) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.peerSet[p.IPAddress] = p
}
func NewPeerSet() *PeerSet {
	return &PeerSet{
		peerSet:           make(map[string]peer),
		state:             1,
		downloadRateLimit: 512,
		mu:                sync.RWMutex{},
	}
}
