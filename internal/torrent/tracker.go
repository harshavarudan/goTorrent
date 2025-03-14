package torrent

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/harshavarudan/goTorrent/internal/worker"
)

// Single instance of tracker (only udp trackers)
// single socket for udp is enough

type AnnounceRequest struct {
	connectionID uint64
	infoHash     [20]byte
	peerID       [20]byte
	downloaded   int64
	left         int64
	uploaded     int64
	event        uint32
	ip           uint32
	key          uint32
	numWant      int32
	port         uint16
}
type TrackerSet struct {
	conn  *net.UDPConn
	state int
	//0 for tracker to stop tracking,1 to periodically check
	//Peer set for tracker, different from torrent peer set but essentially does the same
	//Torrent peer set has to be put in sync with tracker peer set
	//design decision for no reason
	trackerSet map[string]*Tracker
	dispatcher *worker.Dispatcher
	quit       chan struct{}

	//TODO achieve unique peer id
	//id is only generated once. Normally an id is set every time the client loads and should be the same until it’s closed.
	//Add number of workers needed time for periodic check ...
	workerCount int
}

var periodicCheck = 60 * time.Second
var defaultPort uint16 = 6881

type Tracker struct {
	address        *net.UDPAddr
	isAlive        bool
	retries        int
	lastConnection time.Time
	state          int //1 to represent it is available
	mu             *sync.RWMutex
	id             int
}

func (t *Tracker) makeCallAndUpdatePeerSet(ps *PeerSet, conn *net.UDPConn, infoHash [20]byte, fm *FileStatusMetadata) {
	t.lastConnection = time.Now()
	t.mu.Lock()
	defer t.mu.Unlock()
	id, err := establishConnection(conn, t.address)
	if err != nil {
		fmt.Println("Error establishing connection:", err)
		t.ConnFailed()
		return
	}
	//TODO change peer id
	var peerID [20]byte
	_, err = rand.Read(peerID[:])
	if err != nil {
		fmt.Println("Error generating peer ID:", err)
		return
	}
	//change default values
	fm.mu.RLock()
	receivedPacket, err := announce(AnnounceRequest{
		connectionID: id,
		infoHash:     infoHash,
		peerID:       peerID,
		downloaded:   fm.downloaded,
		left:         fm.left, // Default value, replace with actual remaining size
		uploaded:     fm.uploaded,
		event:        0,
		ip:           0,           // Default IP, address usually infers it
		key:          0,           // Default key, replace with actual key
		numWant:      -1,          // Default to requesting all peers
		port:         defaultPort, // Default port, replace with actual port if needed
	}, conn, t.address)

	fm.mu.RUnlock()
	if err != nil {
		t.ConnFailed()
		return
	}
	t.ConnSuccess()
	peerList := getPeersFromAnnounceResponse(receivedPacket)
	for _, p := range peerList {
		println("IP Address:"+p.IPAddress+" with port:", p.tcpPort)
	}
	//Add to peer set
	for _, p := range peerList {
		ps.AddPeer(p)
	}

}

func NewTrackerSet(workerCount int) *TrackerSet {
	dispatcher := worker.NewDispatcher(workerCount)
	dispatcher.Run()
	return &TrackerSet{
		trackerSet:  make(map[string]*Tracker),
		dispatcher:  dispatcher,
		workerCount: workerCount,
		quit:        make(chan struct{}),
	}
}
func getPeersFromAnnounceResponse(receivedPacket []byte) []peer {
	if len(receivedPacket) < 20 {
		return nil // Not enough data to parse peers
	}

	var peers []peer
	offset := 20 // Start reading peer data after tracker response header

	for offset+6 <= len(receivedPacket) {
		ip := net.IP(receivedPacket[offset : offset+4]).String()
		port := int(receivedPacket[offset+4])<<8 | int(receivedPacket[offset+5]) // Convert bytes to int

		p := peer{
			IPAddress: ip,
			tcpPort:   port,
		}

		peers = append(peers, p)
		offset += 6 // Move to the next peer
	}

	return peers
}

type JobFunc func()

func (jf JobFunc) Job() {
	jf()
}

func (ts *TrackerSet) Start(ps *PeerSet, mdi *MetaDataInfo, fm *FileStatusMetadata) {
	ts.dispatcher.Run()
	go ts.StartLoop(ps, mdi, fm)
}
func (t *Tracker) ConnFailed() {
	t.retries += 1
	t.isAlive = false
	t.lastConnection = time.Now()
}
func (t *Tracker) ConnSuccess() {
	t.retries = 0
	t.isAlive = true
	t.lastConnection = time.Now()
}

// StartLoop processes one tracker per tick (1 second) and stops when quit is signaled.
func (ts *TrackerSet) StartLoop(ps *PeerSet, mdi *MetaDataInfo, fm *FileStatusMetadata) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ts.quit:
			fmt.Println("TrackerSet loop: received quit signal, stopping loop")
			return
		case <-ticker.C:
			//we will do a probabilistic fetch
			var tracker *Tracker
			for _, tracker = range ts.trackerSet {
				break
			}

			if tracker.validToCall() {
				println("Tracker is valid to be called ", tracker.id)
				ts.dispatcher.SendJob(JobFunc(func() {
					println("Calling job tracker with add:", tracker.address)
					tracker.lastConnection = time.Now()
					tracker.makeCallAndUpdatePeerSet(ps, ts.conn, mdi.InfoHash, fm)
				}))
			} else {
				println("Tracker is not valid to be called ", tracker.id)
			}
		}
	}
}
func (ts *TrackerSet) StopLoop() {
	close(ts.quit)
	fmt.Println("TrackerSet loop has been killed.")
}
func (t *Tracker) validToCall() bool {
	if t.state == 0 {
		return false
	}
	if t.isAlive == false {
		if time.Since(t.lastConnection) < time.Duration(math.Min(10*math.Pow(2, float64(t.retries)), 100))*time.Second {
			return false
		}

	}
	if t.isAlive == true && time.Since(t.lastConnection) < periodicCheck {
		return false
	}

	return true
}
func (ts *TrackerSet) Init(mdi *MetaDataInfo) error {
	//Add list of trackers
	trackerList := []string{mdi.Announce}
	for _, stringArr := range mdi.AnnounceList {
		if len(stringArr) != 0 {
			trackerList = append(trackerList, stringArr[0])
		}
	}
	//create new socket
	if ts.conn == nil {
		var err error
		ts.conn, err = net.ListenUDP("udp", nil)

		if err != nil {
			fmt.Println("Error creating UDP socket:", err)
			return err
		}
	}

	//set state
	ts.state = 1

	ts.quit = make(chan struct{})

	//set dispatcher
	ts.dispatcher = worker.NewDispatcher(ts.workerCount)

	//Create Set
	if ts.trackerSet == nil {
		ts.trackerSet = make(map[string]*Tracker)
	}

	//adding to tracker set
	for id, url := range trackerList {
		udpAddr, ok := parseUDPTrackerList(url)
		if ok {

			fmt.Println(udpAddr)
			ts.trackerSet[url] = &Tracker{
				address:        udpAddr,
				isAlive:        false,
				retries:        0,
				lastConnection: time.Time{},
				state:          1,
				mu:             &sync.RWMutex{},
				id:             id,
			}
		}
	}
	println("Number of trackers : ", len(ts.trackerSet))
	return nil
}

// TODO parse only udp ones and not others and return bool
func parseUDPTrackerList(url string) (*net.UDPAddr, bool) {
	//TODO parse properly
	url, _ = strings.CutPrefix(url, "udp://")
	url, _ = strings.CutSuffix(url, "/announce")

	fmt.Println("TrackerSet URL:", url)
	udpAddr, err := net.ResolveUDPAddr("udp", url)
	if err != nil {
		fmt.Println("Error resolving UDP address:", err)
		return nil, false
	}
	return udpAddr, true
}

func establishConnection(conn *net.UDPConn, addr *net.UDPAddr) (uint64, error) {
	//send a connection request

	// Construct buffer
	buffer := make([]byte, 16)
	// Connect request fixed magic number
	magic := uint64(0x41727101980)
	// Action for connect request is 0
	action := uint32(0)
	// Generate a random transaction ID
	transactionID := uint32(time.Now().UnixNano())
	// Construct the connect request packet

	binary.BigEndian.PutUint64(buffer[0:8], magic)
	binary.BigEndian.PutUint32(buffer[8:12], action)
	binary.BigEndian.PutUint32(buffer[12:16], transactionID)

	_, err := conn.WriteToUDP(buffer, addr)
	if err != nil {
		return 0, fmt.Errorf("failed to send connect request: %v for addr: %v", err, addr)
	}

	fmt.Println("Connect request sent to", addr.String())
	//receive response
	received := make([]byte, 16)
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second)) //set time before to read tracker
	_, err = conn.Read(received)
	if err != nil {
		return 0, fmt.Errorf("failed to recieve response: %v for addr: %v", err, addr)
	}
	r_action := binary.BigEndian.Uint32(received[0:4])
	r_transactionID := binary.BigEndian.Uint32(received[4:8])

	if r_action != 0 || r_transactionID != transactionID {
		return 0, fmt.Errorf("action or transactionID mismatch ")
	}
	connectionID := binary.BigEndian.Uint64(received[8:16])
	fmt.Println("Connection ID: ", connectionID)
	return connectionID, nil

}

func announce(request AnnounceRequest, conn *net.UDPConn, addr *net.UDPAddr) ([]byte, error) {
	//Construct buffer
	buffer := make([]byte, 98)
	// Action for announce request is 1
	action := uint32(1)

	// Generate a random transaction ID
	transactionID := uint32(time.Now().UnixNano())

	// Construct the announcement request packet
	/*
		Offset  Size    Name    Value
		0       64-bit integer  connection_id
		8       32-bit integer  action          1 // announce
		12      32-bit integer  transaction_id
		16      20-byte string  info_hash
		36      20-byte string  peer_id
		56      64-bit integer  downloaded
		64      64-bit integer  left
		72      64-bit integer  uploaded
		80      32-bit integer  event           0 // 0: none; 1: completed; 2: started; 3: stopped
		84      32-bit integer  IP address      0 // default
		88      32-bit integer  key             ? // random
		92      32-bit integer  num_want        -1 // default
		96      16-bit integer  port            ? // should be between
	*/
	binary.BigEndian.PutUint64(buffer[0:8], request.connectionID)
	binary.BigEndian.PutUint32(buffer[8:12], action)
	binary.BigEndian.PutUint32(buffer[12:16], transactionID)
	copy(buffer[16:36], request.infoHash[:])
	copy(buffer[36:56], request.peerID[:])
	binary.BigEndian.PutUint64(buffer[56:64], uint64(request.downloaded))
	binary.BigEndian.PutUint64(buffer[64:72], uint64(request.left))
	binary.BigEndian.PutUint64(buffer[72:80], uint64(request.uploaded))
	binary.BigEndian.PutUint32(buffer[80:84], request.event)
	binary.BigEndian.PutUint32(buffer[84:88], request.ip)
	binary.BigEndian.PutUint32(buffer[88:92], request.key)
	binary.BigEndian.PutUint32(buffer[92:96], uint32(request.numWant))
	binary.BigEndian.PutUint16(buffer[96:98], request.port)
	_, err := conn.WriteToUDP(buffer, addr)
	if err != nil {
		fmt.Println("Error sending announce request:", err)
		return nil, err
	}
	fmt.Println("Announce request sent to", addr.String())
	//receive response
	received := make([]byte, 98)
	_ = conn.SetReadDeadline(time.Now().Add(30 * time.Second)) //set time before to read tracker
	_, err = conn.Read(received)
	if err != nil {
		fmt.Println("Error receiving announce response:", err)
		return nil, err
	}
	r_action := binary.BigEndian.Uint32(received[0:4])
	r_transactionID := binary.BigEndian.Uint32(received[4:8])
	if r_action != 1 || r_transactionID != transactionID {
		fmt.Println("Action or transactionID mismatch")
		err = errors.New("action or transactionID mismatch")
		return nil, err
	}
	//parse response
	fmt.Println("Announce response received from", addr.String())
	interval := binary.BigEndian.Uint32(received[8:12])
	leechers := binary.BigEndian.Uint32(received[12:16])
	seeders := binary.BigEndian.Uint32(received[16:20])
	fmt.Println("Interval:", interval)
	fmt.Println("Leechers:", leechers)
	fmt.Println("Seeders:", seeders)
	return received, nil
}
