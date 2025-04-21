package torrent

import (
	"fmt"
	"math"
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

	//downloaded details
	uploaded        int
	downloaded      int
	am_choking      bool
	am_interested   bool
	peer_choking    bool
	peer_interested bool

	mu sync.RWMutex
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
	p.state = 1
	p.isConnected = false
	ps.peerSet[p.IPAddress] = &p
	println("Peer successfully added total length ", len(ps.peerSet))
}
func (ps *PeerSet) Connect() {
	// Make a copy of the peers for safe iteration.
	ps.mu.RLock()
	peers := make([]*peer, 0, len(ps.peerSet))
	for _, p := range ps.peerSet {
		peers = append(peers, p)
		fmt.Println("Peers for tcp connection", p)
	}
	ps.mu.RUnlock()

	for _, p := range peers {
		if !p.ValidToCall() {
			continue
		}

		// Initiate connection in its own goroutine. //  ** * * ** *  (but y :(  )
		go func(peer *peer) {
			fmt.Println("Calling peer for tcp connection", p)
			address := net.JoinHostPort(peer.IPAddress, strconv.Itoa(peer.tcpPort))
			tcpAddr, err := net.ResolveTCPAddr("tcp", address)
			if err != nil {
				p.state = 0
				fmt.Printf("Error resolving address for %s: %v\n", address, err)
				return
			}

			conn, err := net.DialTCP("tcp", nil, tcpAddr)
			if err != nil {
				fmt.Printf("Error connecting to peer %s: %v\n", address, err)
				peer.retries++
				p.lastConnected = time.Now()
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

func (p *peer) ConnectionError() {
	p.isConnected = false
	p.retries++
	p.lastConnected = time.Now()
}
func (p *peer) ValidToCall() bool {
	if p.isConnected || p.state == 0 ||
		time.Since(p.lastConnected) < time.Duration(math.Min(5*math.Pow(2, float64(p.retries)), 250))*time.Second {
		return false
	}
	return true
}

//add
