package messaging

import "sync"

type SimpleMessage struct {
	OriginalName  string
	RelayPeerAddr string
	Contents      string
}

type GossipPacket struct {
	Simple *SimpleMessage
	Rumor  *RumorMessage
	Status *StatusPacket
}

type RumorMessage struct {
	Origin string
	ID     uint32
	Text   string
}

type Message struct {
	Text string
}

type PeerStatus struct {
	Identifier string
	NextID     uint32
}

type StatusPacket struct {
	Want []PeerStatus
}

type RumormongeringSession struct {
	Message  RumorMessage
	TimeLeft int
	Active   bool
	Mutex    sync.Mutex
}

func (r *RumormongeringSession) DecrementTimer() {
	r.TimeLeft--
}

func (r *RumormongeringSession) ResetTimer() {
	r.TimeLeft = 10
}

func (r *RumormongeringSession) SetActive(value bool) {
	r.Active = value
}

func (r RumormongeringSession) Pointer() *RumormongeringSession {
	return &r
}
