package game

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/imt9619-wq/hyena/game/input"
	"github.com/imt9619-wq/hyena/game/movements"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func (gs *GameState) moveTick() {
	in := gs.player.spiltInMovement(gs.in)
	out := gs.player.doMove(in)
	gs.moveBuf.addTick(in, out)
	gs.setMoveFlags(out)
	gs.setInputFlags()
	gs.packets.Append(gs.PlayerAuthInputWithState())
}

func (gs *GameState) setMoveFlags(nowOut *movements.OutMovement){
	flag := nowOut.Flag
	if flag.HorizontalCollision{
		gs.SetFlag(packet.InputFlagHorizontalCollision)
	}
	if flag.VerticalCollision{
		gs.SetFlag(packet.InputFlagVerticalCollision)
	}
	if flag.StartedJumping{
		gs.SetFlag(packet.InputFlagStartJumping)
	}
	if flag.WantDown{
		gs.SetFlag(packet.InputFlagWantDown)
	}
	if flag.WantUp{
		gs.SetFlag(packet.InputFlagWantUp)
	}
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
	pk.Pitch, pk.InteractPitch = gs.in.Pitch, gs.in.Pitch
	pk.Yaw, pk.InteractYaw, pk.HeadYaw = gs.in.Yaw, gs.in.Yaw, gs.in.Yaw
	pk.Position = gs.player.position
	out, ok := gs.moveBuf.outMoveWithTick(gs.tick - 1)
	if ok {
		pk.Delta = gs.player.position.Sub(out.simResult.Position)
	}
	return pk
}

func (gs *GameState) setInputFlags() {
	nowIn := gs.in
	in, ok := gs.moveBuf.outMoveWithTick(gs.tick - 1)
	if nowIn.Space.Pressed {
		gs.SetFlag(packet.InputFlagJumping)
		gs.SetFlag(packet.InputFlagJumpCurrentRaw)
		gs.SetFlag(packet.InputFlagJumpDown)
	}
	if nowIn.Sprint.Pressed {
		gs.SetFlag(packet.InputFlagSprinting)
		gs.SetFlag(packet.InputFlagSprintDown)
	}
	if nowIn.Shift.Pressed {
		gs.SetFlag(packet.InputFlagSneaking)
		gs.SetFlag(packet.InputFlagSneakDown)
		gs.SetFlag(packet.InputFlagSneakCurrentRaw)
	}
	lastIn := input.Inputs{}
	if ok {
		lastIn = in.simInMove.Input
	}
	if !lastIn.Space.Pressed && nowIn.Space.Pressed {
		gs.SetFlag(packet.InputFlagJumpPressedRaw)
	}
	if lastIn.Space.Pressed && !nowIn.Space.Pressed {
		gs.SetFlag(packet.InputFlagJumpReleasedRaw)
	}
	if !lastIn.Shift.Pressed && nowIn.Shift.Pressed {
		gs.SetFlag(packet.InputFlagSneakPressedRaw)
		gs.SetFlag(packet.InputFlagStartSneaking)
	}
	if lastIn.Shift.Pressed && !nowIn.Shift.Pressed {
		gs.SetFlag(packet.InputFlagStopSneaking)
		gs.SetFlag(packet.InputFlagSneakReleasedRaw)
	}
	if !lastIn.Sprint.Pressed && nowIn.Sprint.Pressed {
		gs.SetFlag(packet.InputFlagStartSprinting)
	}
	if lastIn.Sprint.Pressed && !nowIn.Sprint.Pressed {
		gs.SetFlag(packet.InputFlagStopSprinting)
	}
}

func (gs *GameState) SetFlag(flag int) {
	gs.tickInputDataFlags.Set(flag)
}

// Reset all bits in ps.tickInputDataFlags to 0
func (gs *GameState) resetFlags() {
	gs.tickInputDataFlags = protocol.NewBitset(packet.PlayerAuthInputBitsetSize)
}

func (gs *GameState) setInputFlagBlockBreakingDelayEnabled() {
	gs.SetFlag(packet.InputFlagBlockBreakingDelayEnabled)
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