package torrent

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"time"
)

// Single instance of tracker (only udp trackers)
// single socket for udp is enough

type AnnounceRequest struct {
	connectionID uint64
	infoHash     [20]byte
	peerID       [20]byte
	downloaded   uint64
	left         uint64
	uploaded     uint64
	event        uint32
	ip           uint32
	key          uint32
	numWant      int32
	port         uint16
}
type TrackerSet struct {
	conn  *net.UDPConn
	state int //0 for tracker to stop tracking,1 to periodically check
	//Peer set for tracker, different from torrent peer set but essentially does the same
	//Torrent peer set has to be put in sync with tracker peer set
	//design decision for no reason
	trackerSet map[Tracker]bool
	//TODO achieve unique peer id
	//id is only generated once. Normally an id is set every time the client loads and should be the same until it’s closed.

	//Add number of workers needed time for periodic check ...
	workerCount                int
	periodicCheckTimeInSeconds time.Duration
}
type Tracker struct {
	address        *net.UDPAddr
	isAlive        bool
	retries        int
	lastConnection time.Time
}
type TrackerJob struct {
	trackerAddr string
	// additional fields as needed
}

func (t TrackerJob) job() {
	//TODO implement me
	println("implement me")
}

func (t TrackerSet) Init(fm *MetaDataInfo, trackerList ...string) {
	//create new socket

	if t.conn == nil {
		var err error
		t.conn, err = net.ListenUDP("udp", nil)

		if err != nil {
			fmt.Println("Error creating UDP socket:", err)
			return
		}
	}
	if t.trackerSet == nil {
		t.trackerSet = make(map[Tracker]bool)
	}

	t.state = 1

	//adding to tracker set
	for _, val := range trackerList {
		udpAddr, ok := parseUDPTrackerList(val)
		if ok {
			//TODO
			fmt.Println(udpAddr)
			//t.trackerSet[udpAddr] = true
		}
	}
	t.establishConnection(fm)

}

// TODO establish via go routines
// implement retry also
func (t TrackerSet) establishConnection(fm *MetaDataInfo) {
	//New Dispatcher

	for tracker, ok := range t.trackerSet {
		addresses := tracker.address
		if ok {
			id, err := establishConnection(t.conn, addresses)
			if err != nil {
				fmt.Println("Error establishing connection:", err)
				continue
			}
			//TODO change peer id
			var peerID [20]byte
			_, err = rand.Read(peerID[:])
			if err != nil {
				fmt.Println("Error generating peer ID:", err)
				return
			}
			//change default values
			//TODO get details of files from downloader struct?
			announce(AnnounceRequest{
				connectionID: id,
				infoHash:     fm.InfoHash,
				peerID:       peerID,
				downloaded:   0,
				left:         2097152, // Default value, replace with actual remaining size
				uploaded:     0,
				event:        0,
				ip:           0,    // Default IP, tracker usually infers it
				key:          0,    // Default key, replace with actual key
				numWant:      -1,   // Default to requesting all peers
				port:         6881, // Default port, replace with actual port if needed
			}, t.conn, addresses)

		}
	}

}

// TODO parse only udp ones and not others and return bool
func parseUDPTrackerList(url string) (*net.UDPAddr, bool) {

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
	connectionID := binary.BigEndian.Uint64(received[8:16])
	if r_action != 0 || r_transactionID != transactionID {
		return 0, fmt.Errorf("action or transactionID mismatch ")
	}
	fmt.Println("Connection ID: ", connectionID)
	return connectionID, nil

}

// TODO add to peer set

func announce(request AnnounceRequest, conn *net.UDPConn, addr *net.UDPAddr) {
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
	binary.BigEndian.PutUint64(buffer[56:64], request.downloaded)
	binary.BigEndian.PutUint64(buffer[64:72], request.left)
	binary.BigEndian.PutUint64(buffer[72:80], request.uploaded)
	binary.BigEndian.PutUint32(buffer[80:84], request.event)
	binary.BigEndian.PutUint32(buffer[84:88], request.ip)
	binary.BigEndian.PutUint32(buffer[88:92], request.key)
	binary.BigEndian.PutUint32(buffer[92:96], uint32(request.numWant))
	binary.BigEndian.PutUint16(buffer[96:98], request.port)
	_, err := conn.WriteToUDP(buffer, addr)
	if err != nil {
		fmt.Println("Error sending announce request:", err)
	}
	fmt.Println("Announce request sent to", addr.String())
	//receive response
	received := make([]byte, 98)
	_ = conn.SetReadDeadline(time.Now().Add(30 * time.Second)) //set time before to read tracker
	_, err = conn.Read(received)
	if err != nil {
		fmt.Println("Error receiving announce response:", err)
	}
	r_action := binary.BigEndian.Uint32(received[0:4])
	r_transactionID := binary.BigEndian.Uint32(received[4:8])
	if r_action != 1 || r_transactionID != transactionID {
		fmt.Println("Action or transactionID mismatch")
	}
	//parse response
	fmt.Println("Announce response received from", addr.String())
	interval := binary.BigEndian.Uint32(received[8:12])
	leechers := binary.BigEndian.Uint32(received[12:16])
	seeders := binary.BigEndian.Uint32(received[16:20])
	fmt.Println("Interval:", interval)
	fmt.Println("Leechers:", leechers)
	fmt.Println("Seeders:", seeders)

}
