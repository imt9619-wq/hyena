package game

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// return a pointer to PlayerAuthInput packet where the fields are filled out based on the
// current GameState
func (gs *GameState) PlayerAuthInputWithState() *packet.PlayerAuthInput {
	pk := &packet.PlayerAuthInput{}
	pk.Tick = uint64(gs.tick)
	pk.InputMode = uint32(gs.clientData.CurrentInputMode)
	pk.PlayMode = packet.PlayModeNormal
	pk.InteractionModel = packet.InteractionModelClassic
	pk.BlockActions = nil
	pk.InputData = gs.tickInputDataFlags
	pk.ItemInteractionData = protocol.UseItemTransactionData{}
	pk.ItemStackRequest = protocol.ItemStackRequest{}
	pk.VehicleRotation = mgl32.Vec2{}
	pk.ClientPredictedVehicle = 0
	pk.AnalogueMoveVector = mgl32.Vec2{}
	pk.CameraOrientation = mgl32.Vec3{}
	pk.RawMoveVector, pk.MoveVector = gs.RawAndMoveVector()
	gs.Player().setPlayerAuthInputWithPlayerState(pk)
	return pk
}

// return a pointer to PlayerAuthInput packet where the fields are filled out based on the
// current GameState
func (gs *GameState) PlayerAuthInputWithStateWithResetInputs() *packet.PlayerAuthInput {
	pk := gs.PlayerAuthInputWithState()
	gs.resetFlags()
	gs.player.addedSpeed = mgl32.Vec3{}
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

func (gs *GameState) StartRunning() {
	gs.Exec(func(q *Qx) {
		gs.Inputs().W.Pressed, gs.Inputs().Sprint.Pressed = true, true
	})
}

func (gs *GameState) StopRunning() {
	gs.Exec(func(q *Qx) {
		gs.Inputs().W.Pressed, gs.Inputs().Sprint.Pressed = false, false
		gs.SetFlag(packet.InputFlagStopSprinting)
	})
}

func (gs *GameState) StartJumping() {
	gs.Exec(func(q *Qx) {
		gs.Inputs().Space.Pressed = true
		gs.SetFlag(packet.InputFlagJumpPressedRaw)
		gs.SetFlag(packet.InputFlagJumpCurrentRaw)
	})
}

func (gs *GameState) StopJumping() {
	gs.Exec(func(q *Qx) {
		gs.Inputs().Space.Pressed = false
		gs.SetFlag(packet.InputFlagJumpReleasedRaw)
	})
}

func (gs *GameState) RawAndMoveVector() (raw mgl32.Vec2, move mgl32.Vec2){
	if gs.tickInputDataFlags.Load(packet.InputFlagUp){
		raw[1] = 1
		move[1] = 1
		if gs.tickInputDataFlags.Load(packet.InputFlagSneakCurrentRaw){
			move[1] = 0.3
		}
	}
	return 
}