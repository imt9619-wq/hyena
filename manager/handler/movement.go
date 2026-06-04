package handler

import (
	"math"
	"sync/atomic"

	"github.com/go-gl/mathgl/mgl32"
)

const(
	Gravity = -0.08 // blocks per ticks^2
	Drag = 0.98 
)


type doMovement interface{
	// doAction simply add a force to the player, for example if a player is on jumping 
	// action, a Y force will be added if onGround is true if only gravity is acting 
	// on the player, the Y force will be 0 instead of the force of the gravity
	doAction(*playerState, map[doMovement]struct{}) 
}

// playerMovement is the force and momentnum applied onto the client, 
// currentAction funcs is going to use the playerMovement and manlapulate it,
// then a playerAuthInput packet will be sent after the force and other field is applied
// and return the player next position to be used in the playerAuthInput packet.
type playerMovement struct {
	sc *sessionConf
	currentAction map[doMovement]struct{}
}


func (pm *playerMovement) tick(){
	defer pm.sc.playerState.Unlock()
	pm.sc.playerState.Lock()
	if !pm.sc.playerState.onReset.CompareAndSwap(true, false){
		for aMove := range pm.currentAction{
			aMove.doAction(pm.sc.playerState, pm.currentAction)
		}
	}
	pm.applyVelocityOnState()
}


// the movement physics caluation is done on this function, playerState
// is changed then will get writen into playerAuthInput
func (pm *playerMovement) applyVelocityOnState(){
	ps := pm.sc.playerState
	ps.RLock()
	defer ps.RUnlock()
	ps.playerPosition.Add(ps.velocity)
}

func (pm *playerMovement) startRunning(){
	pm.currentAction[doRun{}] = struct{}{}
}

func (pm *playerMovement) stopRunning(){
	delete(pm.currentAction, doRun{})
}

type doRun struct{}
// The manittube of the velocity while running (on ground) is as follow: 
// Velocity.. * Slipperiness Multiplier.. * 0.91 + 
// 0.1 * Movement Multiplier(1.3 when running) * Effects Multiplier * (0.6/Slipperiness Multiplier.)^3
// .. means last tick, and . means current tick
// most of the time Slipperiness Multiplier from last tick to the current tick is the same anyway, so
// we just gonna have one slipperinessMultiplier instead of two for each tick
func (dr doRun) doAction(ps *playerState, ca map[doMovement]struct{}){
	slipperinessMultiplier := float32(0.6)
	movementMultiplier := float32(1.3)
	effectsMultiplier := float32(1)

	jumpBoost := float32(0.2)
	if _, ok := ca[doJump{}]; !ok{
		jumpBoost = 0
	}

	ps.Lock()
	defer ps.Unlock()
	yawInradius := float64(ps.yaw) * (math.Pi / 180)

	xVelocity := ps.velocity[0]
	zVelocity := ps.velocity[2]
	velocityValue := xzValue(ps.velocity)

	momentum := velocityValue * slipperinessMultiplier * 0.91
	acceleration := 0.1 * movementMultiplier * effectsMultiplier * float32(math.Pow(0.6/float64(slipperinessMultiplier), 3))
	if !ps.onGround{
		acceleration = 0
	}
	
	ps.velocity[0] = momentum + acceleration * xVelocity / velocityValue + jumpBoost * float32(math.Sin(yawInradius))
	ps.velocity[2] = momentum + acceleration * zVelocity / velocityValue + jumpBoost * float32(math.Cos(yawInradius))
}


func rotationToPitchAndYaw(r mgl32.Vec3) (yaw, pitch float32){
	xzRotateValue := math.Sqrt(math.Pow(float64(r[0]), 2) + math.Pow(float64(r[2]), 2))
	rotateValue := math.Cbrt(math.Pow(xzRotateValue, 2) + math.Pow(float64(r[1]), 2))
	pitch = float32(math.Acos(xzRotateValue/ rotateValue) * 180/math.Pi)
	yaw = float32(math.Acos(float64(r[2])/xzRotateValue) * 180/math.Pi)
	return
}

func xzValue(v mgl32.Vec3) float32{
	return float32(math.Sqrt(math.Pow(float64(v[0]), 2) + math.Pow(float64(v[2]), 2)))
}

type doJump struct{}
func (dj doJump) doAction(ps *playerState, ca map[doMovement]struct{}){

}



func newPlayerMovement(sc *sessionConf) *playerMovement{
	onGround := &atomic.Bool{}
	onGround.Store(true)
	
	return &playerMovement{
		sc: sc,
		currentAction: make(map[doMovement]struct{}, 3),
	}
}
