package game

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/imt9619-wq/hyena/game/input"
	"github.com/imt9619-wq/hyena/game/movements"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func (gs *GameState) moveTick() {
	in := gs.player.SpiltInMovement(gs.in)
	out := gs.player.DoMove(in)
	gs.moveBuf.addTick(in, out)
	gs.setMoveFlags(out)
	gs.setInputFlags()
	gs.packets.Append(gs.PlayerAuthInputWithState())
}

func (gs *GameState) setMoveFlags(nowOut *movements.OutMovement){
	flag := nowOut.Flag
	if flag.HorizontalCollision{
		gs.setflag(packet.InputFlagHorizontalCollision)
	}
	if flag.VerticalCollision{
		gs.setflag(packet.InputFlagVerticalCollision)
	}
	if flag.StartedJumping{
		gs.setflag(packet.InputFlagStartJumping)
	}
	if flag.WantDown{
		gs.setflag(packet.InputFlagWantDown)
	}
	if flag.WantUp{
		gs.setflag(packet.InputFlagWantUp)
	}
}

// return a pointer to PlayerAuthInput packet where the fields are filled out based on the
// current GameState
func (gs *GameState) PlayerAuthInputWithState() *packet.PlayerAuthInput {
	pk := &packet.PlayerAuthInput{}
	pk.InputData = gs.tickInputDataFlags
	pk.RawMoveVector, pk.MoveVector = gs.RawAndMoveVector()
	pk.Tick = uint64(gs.currTick)
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
	pk.Pitch, pk.InteractPitch = gs.in.Pitch, gs.in.Pitch
	pk.Yaw, pk.InteractYaw, pk.HeadYaw = gs.in.Yaw, gs.in.Yaw, gs.in.Yaw
	pk.Position = gs.player.Position
	out, ok := gs.moveBuf.outMoveWithTick(gs.currTick - 1)
	if ok {
		pk.Delta = gs.player.Position.Sub(out.simResult.Position)
	}
	return pk
}

func (gs *GameState) setInputFlags() {
	nowIn := gs.in
	in, ok := gs.moveBuf.outMoveWithTick(gs.currTick - 1)
	if nowIn.Space.Pressed {
		gs.setflag(packet.InputFlagJumping)
		gs.setflag(packet.InputFlagJumpCurrentRaw)
		gs.setflag(packet.InputFlagJumpDown)
	}
	if nowIn.Sprint.Pressed {
		gs.setflag(packet.InputFlagSprinting)
		gs.setflag(packet.InputFlagSprintDown)
	}
	if nowIn.Shift.Pressed {
		gs.setflag(packet.InputFlagSneaking)
		gs.setflag(packet.InputFlagSneakDown)
		gs.setflag(packet.InputFlagSneakCurrentRaw)
	}
	lastIn := input.Inputs{}
	if ok {
		lastIn = in.simInMove.Input
	}
	if !lastIn.Space.Pressed && nowIn.Space.Pressed {
		gs.setflag(packet.InputFlagJumpPressedRaw)
	}
	if lastIn.Space.Pressed && !nowIn.Space.Pressed {
		gs.setflag(packet.InputFlagJumpReleasedRaw)
	}
	if !lastIn.Shift.Pressed && nowIn.Shift.Pressed {
		gs.setflag(packet.InputFlagSneakPressedRaw)
		gs.setflag(packet.InputFlagStartSneaking)
	}
	if lastIn.Shift.Pressed && !nowIn.Shift.Pressed {
		gs.setflag(packet.InputFlagStopSneaking)
		gs.setflag(packet.InputFlagSneakReleasedRaw)
	}
	if !lastIn.Sprint.Pressed && nowIn.Sprint.Pressed {
		gs.setflag(packet.InputFlagStartSprinting)
	}
	if lastIn.Sprint.Pressed && !nowIn.Sprint.Pressed {
		gs.setflag(packet.InputFlagStopSprinting)
	}
}

func (gs *GameState) setflag(flag int) {
	gs.tickInputDataFlags.Set(flag)
}

// Reset all bits in ps.tickInputDataFlags to 0
func (gs *GameState) resetFlags() {
	gs.tickInputDataFlags = protocol.NewBitset(packet.PlayerAuthInputBitsetSize)
}

func (gs *GameState) setInputFlagBlockBreakingDelayEnabled() {
	gs.setflag(packet.InputFlagBlockBreakingDelayEnabled)
}

func (gs *GameState) RawAndMoveVector() (raw mgl32.Vec2, move mgl32.Vec2) {
	in := gs.in
	if in.IsLeftWalk() {
		raw[0] = 1
	}
	if in.IsRightWalk() {
		raw[0] = -1
	}
	if in.IsUpWalk() {
		raw[1] = 1
	}
	if in.IsDownWalk() {
		raw[1] = -1
	}
	move = raw
	if in.IsSneak() {
		move = move.Mul(0.3)
	}
	if in.IsStrafe() && !in.IsSneak() {
		move = move.Mul(0.98)
	}
	return
}