package pkbuf

import (
	"iter"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type PacketBuffer []packet.Packet

func NewPacketBuffer(size int, nilBuf bool)

func (pb *PacketBuffer) Append(pk packet.Packet){
	*pb = append(*pb, pk)
}

func (pb *PacketBuffer) Reset(){
	if len(*pb) == 0{
		return
	}
	(*pb)[0] = nil
	*pb = (*pb)[:0]
}

func (pb *PacketBuffer) FlushPackets() iter.Seq[packet.Packet]{
	return func(yield func(packet.Packet) bool) {
		if len(*pb) == 0{
			return 
		}
		for n := len(*pb)-1; n >= 0; n--{
			if !yield((*pb)[n]){
				return 
			}
			(*pb) = (*pb)[:n]
		}
	}
} 