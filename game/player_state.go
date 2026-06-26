package game

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/imt9619-wq/hyena/game/movements"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type playerState struct {
	Position mgl32.Vec3
    Pitch    float32
    Yaw      float32
    Velocity mgl32.Vec3
    OnGround bool
	baseSpeed float32

    isJumping bool
	isRunning bool

    lastTickPos mgl32.Vec3
	addedSpeed mgl32.Vec3
}

func newPlayerState(conn *minecraft.Conn) *playerState {
	ps := &playerState{
		Position: conn.GameData().PlayerPosition,
		Pitch: conn.GameData().Pitch,
		Yaw: conn.GameData().Yaw,
		Velocity: mgl32.Vec3{},
		OnGround: true,
		lastTickPos: conn.GameData().PlayerPosition,
		baseSpeed: float32(movements.DefaultBaseSpeed),
	}
	return ps
}

func (ps *playerState) tick() {
	ps.lastTickPos = ps.Position
}

func (ps *playerState) setPlayerAuthInputWithPlayerState(pk *packet.PlayerAuthInput){
	pk.Pitch, pk.InteractPitch = ps.Pitch, ps.Pitch
	pk.Yaw, pk.InteractYaw, pk.HeadYaw = ps.Yaw, ps.Yaw, ps.Yaw
	pk.Position = ps.Position
	pk.Delta = ps.Position.Sub(ps.lastTickPos)
}

func (gs *GameState) Player() *playerState {
	return gs.player
}