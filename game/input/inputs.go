package input

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

// Represent the minic of keyBoard input
type Inputs struct{
	Keys
    Yaw   float32
    Pitch float32
    // ServerSpeedAdd refer to the velocity value that is added by knockback, explorsion...(Basically
    // velocity that is not contributed by client input)
    ServerSpeedAdd mgl32.Vec3
}

type Keys struct{
	W, A, S, D            KeyPress
    Space, Shift, Sprint  KeyPress
    RightClick, LeftClick KeyPress
}

type KeyPress struct{
	Pressed   bool
    PressOnce bool
}

func (in Inputs) NextTickPresses() Inputs{
	nextIn := Inputs{}
	nextIn.W = in.W.nextTickPress()
	nextIn.A = in.A.nextTickPress()
	nextIn.S = in.S.nextTickPress()
	nextIn.D = in.D.nextTickPress()
	nextIn.Sprint = in.Sprint.nextTickPress()
	nextIn.Shift = in.Shift.nextTickPress()
	nextIn.Space = in.Space.nextTickPress()
	nextIn.RightClick = in.RightClick.nextTickPress()
	nextIn.LeftClick = in.LeftClick.nextTickPress()
	nextIn.Yaw, nextIn.Pitch = in.Yaw, in.Pitch
	return nextIn
}

func (kp KeyPress) nextTickPress() KeyPress{
	if kp.Pressed && !kp.PressOnce{
		return kp
	}
	return KeyPress{}
}

var directionToOffsets = map[[2]int]float64{
    {-1, -1}: -135.0,
    {-1,  0}:  180.0,
    {-1,  1}:  135.0,
    { 0, -1}:  -90.0,
    { 0,  0}:    0.0, 
    { 0,  1}:   90.0,
    { 1, -1}:  -45.0,
    { 1,  0}:    0.0,
    { 1,  1}:   45.0,
}

func (k Keys) KeyOffsets() float64{
	var frontBack, rightLeft int
	if !(k.W.Pressed == k.S.Pressed){
		if k.W.Pressed{
			frontBack = 1
		}else{
			frontBack = -1
		}
	}
	if !(k.A.Pressed == k.D.Pressed){
		if k.D.Pressed{
			rightLeft = 1
		}else{
			rightLeft = -1
		}
	}	
	return directionToOffsets[[2]int{frontBack, rightLeft}]
}

func (k Keys) MovementMultiplier() float64{
	dirMul := func() float64{
		if k.KeyOffsets() == 45 || k.KeyOffsets() == -45{
			if k.Shift.Pressed{
				return math.Sqrt(2) * 0.98
			}
			return 1
		}
		if k.IsStop(){
			return 0
		}
		return 0.98
	}()
	moveMul := func() float64{
		if k.IsStop(){
			return 0
		}
		if k.Shift.Pressed{
			return 0.3
		}
		if k.IsSprinting(){
			return 1.3
		}
		return 1
	}()
	return moveMul * dirMul
}

func (k Keys) IsStop() bool{
	if k.A.Pressed == k.D.Pressed && k.W.Pressed == k.S.Pressed{
		return true
	}
	return false
}

func (k Keys) IsWalk() bool{
	return !(k.IsStop() || k.Sprint.Pressed)
}

func (k Keys) IsUpWalk() bool{
	return k.W.Pressed && !k.S.Pressed
}

func (k Keys) IsRightWalk() bool{
	return k.D.Pressed && !k.A.Pressed
}

func (k Keys) IsStrafe() bool{
	return k.KeyOffsets() == 45 || k.KeyOffsets() == -45
}

func (k Keys) IsLeftWalk() bool{
	return k.A.Pressed && !k.D.Pressed
}

func (k Keys) IsSneak() bool{
	return k.Shift.Pressed
}

func (k Keys) IsJump() bool{
	return k.Space.Pressed
}

func (k Keys) IsDownWalk() bool{
	return k.S.Pressed && !k.W.Pressed
}

func (k Keys) IsSprinting() bool{
	return k.Sprint.Pressed && k.IsUpWalk() && !k.IsSneak() 
}
