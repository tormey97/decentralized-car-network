package gossiper

import (
	"fmt"
	"github.com/tormey97/Peerster/messaging"
	"time"
)

func (peerster *Peerster) addToNextHopTable(rumor messaging.RumorMessage, originAddr string) {
	peerster.NextHopTable.Mutex.Lock()
	defer peerster.NextHopTable.Mutex.Unlock()
	peerster.NextHopTable.Map[rumor.Origin] = originAddr
}

// Sends
func (peerster *Peerster) SendRouteMessage() {
	peerster.sendNewRumorMessage("")
}

// Starts a goroutine that sends route messages periodically (for discovery)
func (peerster *Peerster) SendRouteMessages() {
	if peerster.RTimer == 0 {
		return
	}
	peerster.SendRouteMessage()
	go func() {
		for {
			fmt.Println("SENDING ROUTE MSG")
			time.Sleep(time.Duration(peerster.RTimer) * time.Second)
			peerster.SendRouteMessage()
		}
	}()
}

// Sends a packet using the next hop table to find the path to the recipient.
func (peerster *Peerster) nextHopRoute(packet *messaging.GossipPacket, destination string) {
	peerster.NextHopTable.Mutex.RLock()
	nextHopAddr, ok := peerster.NextHopTable.Map[destination]
	for i := range peerster.NextHopTable.Map {
		fmt.Println(i, peerster.NextHopTable.Map[i], destination)
	}
	peerster.NextHopTable.Mutex.RUnlock()
	if ok {
		err := peerster.sendToPeer(nextHopAddr, *packet, []string{})
		if err != nil {
			fmt.Printf("Unable to send DSHV routed message to %s, reason: %s \n", nextHopAddr, err)
		}
	} else {
		fmt.Printf("Couldn't find %s in next hop table \n", destination)
	}
}
