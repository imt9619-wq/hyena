package game

import (
	"iter"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type packetBuffer []packet.Packet

func (pb packetBuffer) append(pk packet.Packet){
	pb = append(pb, pk)
}

func (pb packetBuffer) reset(){
	if len(pb) == 0{
		return
	}
	pb[0] = nil
	pb = pb[:0]
}

// return a pointer to PlayerAuthInput packet where the fields are filled out based on the
// current GameState
func (gs *GameState) PlayerAuthInputWithState() *packet.PlayerAuthInput {
	pk := &packet.PlayerAuthInput{}
	pk.InputData = gs.tickInputDataFlags
	pk.RawMoveVector, pk.MoveVector = gs.RawAndMoveVector()
	pk.Tick = uint64(gs.tick)
	pk.InputMode = uint32(gs.clientData.CurrentInputMode)
	pk.PlayMode = packet.PlayModeNormal
	pk.InteractionModel = packet.InteractionModelClassic
	pk.BlockActions = nil
	pk.ItemInteractionData = protocol.UseItemTransactionData{}
	pk.ItemStackRequest = protocol.ItemStackRequest{}
	pk.VehicleRotation = mgl32.Vec2{}
	pk.ClientPredictedVehicle = 0
	pk.AnalogueMoveVector = mgl32.Vec2{}
	pk.CameraOrientation = mgl32.Vec3{}
	gs.Player().setPlayerAuthInputWithPlayerState(pk)
	out, ok := gs.moveBuf.outMoveWithTick(gs.tick-1)
	if ok{
		pk.Delta = gs.player.position.Sub(out.simResult.Position)
	}
	return pk
}

func (gs *GameState) SetFlag(flag int){
	gs.tickInputDataFlags.Set(flag)
}

// Reset all bits in ps.tickInputDataFlags to 0
func (gs *GameState) resetFlags() {
	gs.tickInputDataFlags = protocol.NewBitset(packet.PlayerAuthInputBitsetSize)
}

func (gs *GameState) setInputFlagBlockBreakingDelayEnabled() {
	gs.SetFlag(packet.InputFlagBlockBreakingDelayEnabled)
}

func (gs *GameState) RawAndMoveVector() (raw mgl32.Vec2, move mgl32.Vec2){
	if gs.in.IsLeftWalk(){
		raw[0] = 1
	}
	if gs.in.IsRightWalk(){
		raw[0] = -1
	}
	if gs.in.IsUpWalk(){
		raw[1] = 1
	}
	if gs.in.IsDownWalk(){
		raw[1] = -1
	}
	move = raw
	if gs.in.IsSneak(){
		move = move.Mul(0.3)
	}
	if gs.in.IsStrafe() && !gs.in.IsSneak(){
		move = move.Mul(0.98)
	}
	return 
}

func (gs *GameState) FlushPackets() iter.Seq[packet.Packet]{
	return func(yield func(packet.Packet) bool) {
		if len(gs.packets) == 0{
			return 
		}
		for n := len(gs.packets)-1; n >= 0; n--{
			if !yield(gs.packets[n]){
				return 
			}
			gs.packets = gs.packets[:n]
		}
	}
}