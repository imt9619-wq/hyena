package game

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/imt9619-wq/hyena/game/movements"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// return a pointer to PlayerAuthInput packet where the fields are filled out based on the
// current GameState
func (gs *GameState) PlayerAuthInputWithState() *packet.PlayerAuthInput {
	nowOut, ok := gs.moveBuf.outMoveWithTick(gs.tick)
	pk := &packet.PlayerAuthInput{}
	if ok{
		pk.InputData = *nowOut.Input.InputFlags
		pk.RawMoveVector, pk.MoveVector = gs.RawAndMoveVector(nowOut)
	}
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
		pk.Delta = nowOut.Position.Sub(out.Position)
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

func (gs *GameState) RawAndMoveVector(nowOut *movements.OutMovement) (raw mgl32.Vec2, move mgl32.Vec2){
	if nowOut.Input.IsLeftWalk(){
		raw[0] = 1
	}
	if nowOut.Input.IsRightWalk(){
		raw[0] = -1
	}
	if nowOut.Input.IsUpWalk(){
		raw[1] = 1
	}
	if nowOut.Input.IsDownWalk(){
		raw[1] = -1
	}
	move = raw
	if nowOut.Input.IsSneak(){
		move = move.Mul(0.3)
	}
	if nowOut.Input.IsStrafe() && !nowOut.Input.IsSneak(){
		move = move.Mul(0.98)
	}
	return 
}