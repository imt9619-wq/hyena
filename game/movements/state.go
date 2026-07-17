package movements

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/imt9619-wq/hyena/game/blockmap"
	"github.com/imt9619-wq/hyena/game/input"
	"github.com/sandertv/gophertunnel/minecraft"
)

type PlayerState struct {
	Position     mgl32.Vec3
    Velocity     mgl32.Vec3
    OnGround     bool
    BaseSpeed    float32
    JumpCooldown int

	world 		 *blockmap.BlockMap
}

func NewPlayerState(conn *minecraft.Conn, world *blockmap.BlockMap) *PlayerState {
	ps := &PlayerState{
		Position: conn.GameData().PlayerPosition,
		Velocity: mgl32.Vec3{},
		OnGround: false,
		BaseSpeed: float32(DefaultBaseSpeed),
		world: world,
	}
	return ps
}

func (ps *PlayerState) DoMove(in *InMovement) *OutMovement{
	out := SimMovementsInWorld(in, ps.world)
	ps.CopyMovement(&out.AMovement)
	return out
}

func (ps *PlayerState) SplitAMovement() *AMovement{
	a := &AMovement{}
	a.Position = ps.Position
	a.BaseSpeed = ps.BaseSpeed
	a.OnGround = ps.OnGround
	a.Velocity = ps.Velocity
	a.JumpCooldown = ps.JumpCooldown
	return a
}

func (ps *PlayerState) SpiltInMovement(input input.Inputs) *InMovement{
	in := &InMovement{}
	in.Input = input
	in.AMovement = *ps.SplitAMovement()
	return in
}

func (ps *PlayerState) CopyMovement(out *AMovement){
	ps.Position = out.Position
	ps.Velocity = out.Velocity
	ps.OnGround = out.OnGround
	ps.BaseSpeed = out.BaseSpeed
	ps.JumpCooldown = out.JumpCooldown
}
