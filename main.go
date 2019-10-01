package main

import (
	"flag"
	"fmt"
	"github.com/dedis/protobuf"
	"github.com/tormey97/Peerster/messaging"
	"log"
	"net"
	"strings"
	"time"
)

type Origin int

const (
	Client Origin = iota
	Server
)

type Peerster struct {
	UIPort     string
	gossipAddr string
	knownPeers []string
	name       string
	simple     bool
}

func (peerster Peerster) String() string {
	return fmt.Sprintf(
		"UIPort: %s, gossipAddr: %s, knownPeers: %s, name: %s, simple: %s", peerster.UIPort, peerster.gossipAddr, peerster.knownPeers, peerster.name, peerster.simple)
}

func (peerster Peerster) createConnection(origin Origin) (net.PacketConn, error) {
	var addr string
	switch origin {
	case Client:
		addr = "127.0.0.1:" + peerster.UIPort
	case Server:
		addr = peerster.gossipAddr
	}
	return net.ListenPacket("udp4", addr)
}

func readFromConnection(conn net.PacketConn) ([]byte, error) {
	buffer := make([]byte, 1024)
	n, _, err := conn.ReadFrom(buffer)
	if err != nil {
		return nil, err
	}
	buffer = buffer[:n]
	return buffer, nil
}

func (peerster Peerster) clientReceive(buffer []byte, packet messaging.GossipPacket) {
	fmt.Println("CLIENT MESSAGE " + string(buffer))
	packet = messaging.GossipPacket{Simple: peerster.createMessage(string(buffer))}
	err := peerster.sendToKnownPeers(packet, []string{})
	if err != nil {
		fmt.Printf("Error: could not send packet from client, reason: %s", err)
	}
}

func (peerster Peerster) serverReceive(buffer []byte, packet messaging.GossipPacket) {
	receivedPacket := &messaging.GossipPacket{}
	err := protobuf.Decode(buffer, receivedPacket)
	if err != nil {
		fmt.Printf("Error: could not decode packet, reason: %s", err)
	}
	fmt.Printf("SIMPLE MESSAGE origin %s from %s contents %s \n", receivedPacket.Simple.OriginalName, receivedPacket.Simple.RelayPeerAddr, receivedPacket.Simple.Contents)
	blacklist := []string{receivedPacket.Simple.RelayPeerAddr}
	peerster.addToKnownPeers(receivedPacket.Simple.RelayPeerAddr)
	receivedPacket.Simple.RelayPeerAddr = peerster.gossipAddr
	err = peerster.sendToKnownPeers(*receivedPacket, blacklist)
	if err != nil {
		fmt.Printf("Error: could not send packet from some other peer, reason: %s", err)
	}
	peerster.listPeers()
}

func (peerster Peerster) listen(origin Origin) {
	conn, err := peerster.createConnection(origin)
	if err != nil {
		log.Fatalf("Error: could not listen. Origin: %s, error: %s", origin, err)
	}

	for {
		buffer, err := readFromConnection(conn)
		if err != nil {
			log.Printf("Could not read from connection, origin: %s, reason: %s", origin, err)
			break
		}
		var packet messaging.GossipPacket
		switch origin {
		case Client:
			peerster.clientReceive(buffer, packet)
		case Server:
			peerster.serverReceive(buffer, packet)
		}
	}
}

func (peerster *Peerster) addToKnownPeers(address string) {
	if address == peerster.gossipAddr {
		return
	}
	for i := range peerster.knownPeers {
		if address == peerster.knownPeers[i] {
			return
		}
	}
	peerster.knownPeers = append(peerster.knownPeers, address)
}

func (peerster Peerster) createMessage(msg string) *messaging.SimpleMessage {
	return &messaging.SimpleMessage{
		OriginalName:  peerster.name,
		RelayPeerAddr: peerster.gossipAddr,
		Contents:      msg,
	}
}

func (peerster Peerster) listPeers() {
	for i := range peerster.knownPeers {
		peer := peerster.knownPeers[i]
		fmt.Print(peer)
		if i < len(peerster.knownPeers)-1 {
			fmt.Print(",")
		} else {
			fmt.Println()
		}
	}

}

// Sends a GossipPacket to all known peers.
func (peerster Peerster) sendToKnownPeers(packet messaging.GossipPacket, blacklist []string) error {
	for i := range peerster.knownPeers {
		peer := peerster.knownPeers[i]
		if peer == peerster.gossipAddr {
			break
		}
		blacklisted := false
		for j := range blacklist {
			if peer == blacklist[j] {
				blacklisted = true
				break
			}
		}
		if blacklisted {
			break
		}

		conn, err := net.Dial("udp4", peer)
		if err != nil {
			return err
		}
		packetBytes, err := protobuf.Encode(&packet)
		if err != nil {
			return err
		}
		_, err = conn.Write(packetBytes)
		if err != nil {
			return err
		}
	}
	return nil
}

func createPeerster() Peerster {
	UIPort := flag.String("UIPort", "8080", "the port the client uses to communicate with peerster")
	gossipAddr := flag.String("gossipAddr", "127.0.0.1:5000", "the address of the peerster")
	name := flag.String("name", "nodeA", "the name of the node")
	peers := flag.String("peers", "", "known peers")
	simple := flag.Bool("simple", true, "simple mode")
	flag.Parse()
	return Peerster{
		UIPort:     *UIPort,
		gossipAddr: *gossipAddr,
		knownPeers: strings.Split(*peers, ","),
		name:       *name,
		simple:     *simple,
	}
}

func main() {
	peerster := createPeerster()
	//fmt.Println(peerster.String())
	go peerster.listen(Server)
	go peerster.listen(Client)
	time.Sleep(3000 * time.Millisecond)
}
