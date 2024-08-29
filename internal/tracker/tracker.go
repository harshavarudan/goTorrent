package tracker

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"time"
)

// Single instance of tracker (only udp trackers)
// single socket for udp is enough
type AnnounceRequest struct {
}
type Tracker struct {
	conn       *net.UDPConn
	infoHash   [20]byte        //might not use
	peerSet    map[string]bool //might change type
	trackerSet map[*net.UDPAddr]bool
}

func (t Tracker) Init(trackerList ...string) {
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
		t.trackerSet = make(map[*net.UDPAddr]bool)
	}
	if t.peerSet == nil {
		t.peerSet = make(map[string]bool)
	}
	//adding to tracker set
	for _, val := range trackerList {
		udpAddr, ok := parseTrackerList(val)
		if ok {
			t.trackerSet[udpAddr] = true
		}
	}
	t.establishConnection()

}

// establish via go routines
// implement retry also
func (t Tracker) establishConnection() {
	for addresses, ok := range t.trackerSet {
		if ok {
			establishConnection(t.conn, addresses)
		}
	}
}

// TODO parse only udp ones and not others and return bool
func parseTrackerList(url string) (*net.UDPAddr, bool) {

	url, _ = strings.CutPrefix(url, "udp://")
	url, _ = strings.CutSuffix(url, "/announce")

	fmt.Println("Tracker URL:", url)
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

// TODO
func announce() {}
