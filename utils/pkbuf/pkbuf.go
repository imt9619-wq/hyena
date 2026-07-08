package pkbuf

import (
	"iter"
	"sync"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type PacketBuffer struct{
	mu *sync.RWMutex
	buf []packet.Packet
}

func NewPacketBuffer(size int) *PacketBuffer{
	return &PacketBuffer{
		mu: &sync.RWMutex{},
		buf: make([]packet.Packet, 0, size),
	}
}

func (pb *PacketBuffer) Append(pk packet.Packet){
	pb.mu.Lock()
	defer pb.mu.Unlock()
	pb.buf = append(pb.buf, pk)
}

func (pb *PacketBuffer) Reset(){
	pb.mu.Lock()
	defer pb.mu.Unlock()
	if len(pb.buf) == 0{
		return
	}
	(pb.buf)[0] = nil
	pb.buf = (pb.buf)[:0]
}

func (pb *PacketBuffer) Packets() iter.Seq[packet.Packet]{
	return func(yield func(packet.Packet) bool) {
		pb.mu.RLock()
		defer pb.mu.RUnlock()
		for _, pk := range pb.buf{
			if !yield(pk){
				return 
			}
		}
	}
} 