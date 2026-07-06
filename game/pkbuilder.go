package game

import (
	"iter"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/imt9619-wq/hyena/game/input"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type packetBuffer []packet.Packet

func (pb *packetBuffer) append(pk packet.Packet){
	*pb = append(*pb, pk)
}

func (pb *packetBuffer) reset(){
	if len(*pb) == 0{
		return
	}
	(*pb)[0] = nil
	*pb = (*pb)[:0]
}

func (pb *packetBuffer) flushPackets() iter.Seq[packet.Packet]{
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

// return a pointer to PlayerAuthInput packet where the fields are filled out based on the
// current GameState
func (gs *GameState) PlayerAuthInputWithState(in input.Inputs) *packet.PlayerAuthInput {
	pk := &packet.PlayerAuthInput{}
	pk.InputData = gs.tickInputDataFlags
	pk.RawMoveVector, pk.MoveVector = gs.RawAndMoveVector(in)
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
	pk.Pitch, pk.InteractPitch = in.Pitch, in.Pitch
	pk.Yaw, pk.InteractYaw, pk.HeadYaw = in.Yaw, in.Yaw, in.Yaw
	pk.Position = gs.player.position
	out, ok := gs.moveBuf.outMoveWithTick(gs.tick-1)
	if ok{
		pk.Delta = gs.player.position.Sub(out.simResult.Position)
	}
	return pk
}

func (gs *GameState) setInputFlags(nowIn input.Inputs){
	in, ok := gs.moveBuf.outMoveWithTick(gs.tick-1)
	if nowIn.Space.Pressed{
		gs.SetFlag(packet.InputFlagJumping)
		gs.SetFlag(packet.InputFlagJumpCurrentRaw)
		gs.SetFlag(packet.InputFlagJumpDown)
	}
	if nowIn.Sprint.Pressed{
		gs.SetFlag(packet.InputFlagSprinting)
		gs.SetFlag(packet.InputFlagSprintDown)
	}
	if nowIn.Shift.Pressed{
		gs.SetFlag(packet.InputFlagSneaking)
		gs.SetFlag(packet.InputFlagSneakDown)
		gs.SetFlag(packet.InputFlagSneakCurrentRaw)
	}
	lastIn := input.Inputs{}
	if ok{
		lastIn = in.simInMove.Input
	}
	if !lastIn.Space.Pressed && nowIn.Space.Pressed{
		gs.SetFlag(packet.InputFlagJumpPressedRaw)
	}
	if lastIn.Space.Pressed && !nowIn.Space.Pressed{
		gs.SetFlag(packet.InputFlagJumpReleasedRaw)
	}
	if !lastIn.Shift.Pressed && nowIn.Shift.Pressed{
		gs.SetFlag(packet.InputFlagSneakPressedRaw)
		gs.SetFlag(packet.InputFlagStartSneaking)
	}
	if lastIn.Shift.Pressed && !nowIn.Shift.Pressed{
		gs.SetFlag(packet.InputFlagStopSneaking)
		gs.SetFlag(packet.InputFlagSneakReleasedRaw)
	}
	if !lastIn.Sprint.Pressed && nowIn.Sprint.Pressed {
		gs.SetFlag(packet.InputFlagStartSprinting)
	}
	if lastIn.Sprint.Pressed && !nowIn.Sprint.Pressed{
		gs.SetFlag(packet.InputFlagStopSprinting)
	}
}

func flagLoad(flags *protocol.Bitset, flag int) bool{
	if flags == nil{
		return false
	}
	return (*flags).Load(flag)
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

func (gs *GameState) RawAndMoveVector(in input.Inputs) (raw mgl32.Vec2, move mgl32.Vec2){
	if in.IsLeftWalk(){
		raw[0] = 1
	}
	if in.IsRightWalk(){
		raw[0] = -1
	}
	if in.IsUpWalk(){
		raw[1] = 1
	}
	if in.IsDownWalk(){
		raw[1] = -1
	}
	move = raw
	if in.IsSneak(){
		move = move.Mul(0.3)
	}
	if in.IsStrafe() && !in.IsSneak(){
		move = move.Mul(0.98)
	}
	return 
}

func (gs *GameState) FlushPackets() iter.Seq[packet.Packet]{
	return gs.packets.flushPackets()
}