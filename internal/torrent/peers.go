package torrent

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
)

type PeerSet struct {
	state             int
	peerSet           map[string]*peer
	downloadRateLimit int //in kilo bytes
	mu                sync.RWMutex
	//others as required
}
type peer struct {
	IPAddress string
	tcpPort   int
	state     int

	isConnected      bool
	lastConnected    time.Time
	bytesTransferred int
	retries          int
	conn             *net.TCPConn
}

func (ps *PeerSet) AddPeer(p peer) {
	ps.mu.RLock()
	val, ok := ps.peerSet[p.IPAddress]
	if ok && val.IPAddress == p.IPAddress && val.tcpPort == p.tcpPort { //add a peer only if something has changed
		ps.mu.RUnlock()
		return
	}
	ps.mu.RUnlock()

	ps.mu.Lock()
	defer ps.mu.Unlock()
	if p.IPAddress == "0.0.0.0" || p.tcpPort == 0 {
		return
	}
	ps.peerSet[p.IPAddress] = &p
	println("Peer successfully added total length ", len(ps.peerSet))
}
func (ps *PeerSet) Connect() {
	// Make a copy of the peers for safe iteration.
	ps.mu.RLock()
	peers := make([]*peer, 0, len(ps.peerSet))
	for _, p := range ps.peerSet {
		peers = append(peers, p)
	}
	ps.mu.RUnlock()

	for _, p := range peers {
		// Skip if already connected.
		if p.isConnected {
			continue
		}
		// Throttle connection attempts (e.g., wait at least 30 seconds between attempts).
		if time.Since(p.lastConnected) < 30*time.Second {
			continue
		}

		// Update lastAttempted timestamp.
		p.lastConnected = time.Now()

		// Initiate connection in its own goroutine.
		go func(peer *peer) {
			address := net.JoinHostPort(peer.IPAddress, strconv.Itoa(peer.tcpPort))
			tcpAddr, err := net.ResolveTCPAddr("tcp", address)
			if err != nil {
				fmt.Printf("Error resolving address for %s: %v\n", address, err)
				return
			}

			conn, err := net.DialTCP("tcp", nil, tcpAddr)
			if err != nil {
				fmt.Printf("Error connecting to peer %s: %v\n", address, err)
				peer.retries++
				return
			}

			// Update the peer's connection details safely.
			ps.mu.Lock()
			peer.conn = conn
			peer.isConnected = true
			peer.lastConnected = time.Now()
			// Reset retries on successful connection.
			peer.retries = 0
			fmt.Printf("Successfully connected to peer %s\n", address)
			ps.mu.Unlock()
		}(p)
	}
}
func NewPeerSet() *PeerSet {
	return &PeerSet{
		peerSet:           make(map[string]*peer),
		state:             1,
		downloadRateLimit: 512,
		mu:                sync.RWMutex{},
	}
}
