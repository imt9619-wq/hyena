package game

import (
	"iter"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/imt9619-wq/hyena/game/input"
	"github.com/imt9619-wq/hyena/game/movements"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// Qx idea is basically the same as Tx for a dragonfly world, but
// instead of Tx for a world, we have Tx for a GameState, we call
// it Qx instead of Tx so we wont confuse it with the Tx from dragonfly
type Qx struct {
	gs *GameState
	closed bool
}

func (qx *Qx) close() {
	qx.closed = true
}

type QueueFunc func(*Qx)

type queueTransition struct {
	c chan struct{}
	f QueueFunc
}

func (gs *GameState) startRunningQueue() {
	go func ()  {
		for {
			select {
			case <-gs.closed:
				return
			case q := <-gs.queue:
				q.Run(gs)
			}
		}
	}()
}

func (gs *GameState) Exec(f QueueFunc) chan struct{} {
	ch := make(chan struct{})
	gs.queue <- &queueTransition{c: ch, f: f}
	return ch
}

func (q *queueTransition) Run(gs *GameState) {
	qx := &Qx{gs: gs, closed: false}
	q.f(qx)
	qx.close()
	close(q.c)
}

func (qx *Qx) checkIsValidQx(){
	if qx.closed == true{
		panic("Calling Qx methods on closed Qx.")
	}
}

func (qx *Qx) Tick(){
	qx.checkIsValidQx()
	qx.gs.tick()
}

func (qx *Qx) SetItemOnInvSlot(windowId uint32, slot uint32, ist protocol.ItemInstance){
	qx.checkIsValidQx()
	qx.gs.items.SetItemOnInvSlot(windowId, slot, ist)
}

func (qx *Qx) SetInput(f func(*input.Inputs)){
	qx.checkIsValidQx()
	in := qx.gs.in
	f(&in)
	qx.gs.in = in
}

func (qx *Qx) Equip(pk *packet.MobEquipment){
	qx.checkIsValidQx()
	qx.gs.items.Equip(pk)
}

func (qx *Qx) SyncInventoryContent(pk *packet.InventoryContent){
	qx.checkIsValidQx()
	qx.gs.items.SyncInventoryContent(pk)
}

func (qx *Qx) UpdateChunkCentre(pos mgl32.Vec3){
	qx.checkIsValidQx()
	qx.gs.blockMap.UpdateChunkCentre(pos)
}

func (qx *Qx) SetInputFlag(flag int){
	qx.checkIsValidQx()
	qx.gs.setflag(flag)
}

func (qx *Qx) ResimMove(tick uint, modF func(*movements.InMovement)){
	qx.checkIsValidQx()
	qx.gs.reSimMoveAtTick(tick, modF)
}

func (qx *Qx) SetBlock(pos protocol.BlockPos, layer uint8, block uint32){
	qx.checkIsValidQx()
	qx.gs.blockMap.SetBlock(pos, layer, block)
}

func (qx *Qx) UpdateChunkRadius(r int32){
	qx.checkIsValidQx()
	qx.gs.blockMap.UpdateChunkRadius(r)
}

func (qx *Qx) InsertLevelChunk(pk *packet.LevelChunk){
	qx.checkIsValidQx()
	qx.gs.blockMap.InsertLevelChunk(pk)
}

func (qx *Qx) InsertSubChunk(pk *packet.SubChunk){
	qx.checkIsValidQx()
	qx.gs.blockMap.InsertSubChunk(pk)
}

func (qx *Qx) FlushPackets() iter.Seq[packet.Packet]{
	return qx.gs.packets.FlushPackets()
}