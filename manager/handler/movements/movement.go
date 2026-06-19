package movements

import (
	"fmt"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/imt9619-wq/hyena/game"
	"github.com/imt9619-wq/hyena/utils"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type Movement struct {
	state         *game.GameState
	position      mgl64.Vec3
	velocity      mgl64.Vec3
	onGround      bool
	isrunning     bool
	isjumping     bool
	moveVector    [3]int

	scratch *collisionScratch
}

func NewMovement(state *game.GameState) *Movement {
	return &Movement{
		state:    state,
		scratch:  newCollisionScratch(),
	}
}

func (m *Movement) playerPosBeforeVelocityApply() mgl64.Vec3 {
	return m.position.Sub(m.velocity)
}

func (m *Movement) Tick() {
	now := time.Now()
	m.copyPlayerState()
	m.doMotions()
	m.resetVelocityIfBlocked()
	m.applyVelocity()
	m.applyCollision(m.getCollision())
	m.setOnGround()
	m.pasteToPlayerState()
	fmt.Printf("Movement on tick %d: {position: %v velocity: %v onGrond: %v}\n", m.state.GStick(), m.position, m.velocity, m.onGround)
	fmt.Printf("Time used for tick %d: %0.2fms\n", m.state.GStick(), time.Since(now).Seconds()*1000)
	fmt.Printf("Block pos based on pPos: %v\n\n", cube.PosFromVec3(m.position))
}

func (m *Movement) pasteToPlayerState() {
	ps := m.state.Player()
	ps.Velocity = utils.Mgl64Vec3Tomgl32Vec3(m.velocity)
	ps.Position = utils.Mgl64Vec3Tomgl32Vec3(m.position.Add(mgl64.Vec3{0, utils.NetworkOffset, 0}))
	ps.OnGround = m.onGround 
}

func (m *Movement) copyPlayerState() {
	ps := m.state.Player()
	m.velocity = utils.Mgl32Vec3Tomgl64Vec3(ps.Velocity)
	m.setMoveVector(m.velocity)
	m.position = utils.Mgl32Vec3Tomgl64Vec3(ps.Position).Sub(mgl64.Vec3{0, utils.NetworkOffset, 0})

	m.onGround = ps.OnGround
}

func (m *Movement) setMoveVector(v mgl64.Vec3){
	for axis, plane := range v{
		if plane == 0{
			m.moveVector[axis] = 0
		}else if plane > 0{
			m.moveVector[axis] = 1
		}else{
			m.moveVector[axis] = -1
		}
	}
}

// checkIfBlocked will set certain moveVector axis(except Y) to 0 if on that axis the player movement is blocked
// (like right in front of the wall)
func (m *Movement) checkIfBlocked() {
	halfHori := utils.HoriProbeOffset / 2
	for axis, dir := range m.moveVector{
		if axis == 1 || dir == 0{
			continue
		}
		bbpos := m.position
		bbpos[axis] += (utils.PlayerWidth/2 + halfHori) * float64(m.moveVector[axis])
		otherAxe := (2*axis+2)%6 // 0 if axis is 2, 2 if axis is 0
        if m.bboxIntersectsSolid(cube.Box(
			bbpos[0]-halfHori,
			bbpos[1]+utils.PlayerHeight,
			bbpos[2]-halfHori,
			bbpos[0]+halfHori,
			bbpos[1]+stepHeight,
			bbpos[2]+halfHori,
			).Stretch(cube.Axis(otherAxe), utils.PlayerWidth/2-halfHori)){
				m.moveVector[axis] = 0
		}
	}
}

func (m *Movement) resetVelocityIfBlocked(){
	m.checkIfBlocked()
	for axis, dir := range m.moveVector{
		if dir == 0{
			m.velocity[axis] = 0
		}
	}
}

func (m *Movement) setOnGround() {
	m.onGround= false
	halfW := utils.PlayerWidth / 2
	pos := m.position
	tinyBBox := cube.Box(
		pos[0]-halfW,
		pos[1]-utils.GroundProbeOffset,
		pos[2]-halfW,
		pos[0]+halfW,
		pos[1],
		pos[2]+halfW,
	)
	if m.velocity[1] == 0 && m.bboxIntersectsSolid(tinyBBox) {
		m.onGround = true
		m.state.Player().SetFlag(packet.InputFlagVerticalCollision)
	}
}