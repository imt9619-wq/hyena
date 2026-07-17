package utils

import (
	"iter"
	"math"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
)

// lineCoordAt returns the coordinates on the whole vector point where a movement line
// from origin with direction crosses the given plane on fixedAxis.
func LineCoordAt(origin, direction mgl64.Vec3, fixedAxis int, plane float64) (mgl64.Vec3, bool) {
	if direction[fixedAxis] == 0 {
		return mgl64.Vec3{}, false
	}
	t := (plane - origin[fixedAxis]) / direction[fixedAxis]
	var out mgl64.Vec3
	for axis, val := range origin {
		if axis == fixedAxis {
			out[axis] = plane
			continue
		}
		out[axis] = val + t*direction[axis]
	}
	return out, true
}

func MinOffset(offsets [3]float64, delta mgl64.Vec3) iter.Seq2[int, float64] {
	var radio [3]float64
	for axis, plane := range delta {
		if plane == 0 {
			radio[axis] = 1
			continue
		}
		radio[axis] = offsets[axis] / plane
	}
	minRadio := min(radio[0], radio[1], radio[2])
	return func(yield func(int, float64) bool) {
		for i, r := range radio {
			if r != minRadio {
				continue
			}
			if !yield(i, r) {
				return
			}
		}
	}
}

func OutOfPlane(self, nearby cube.BBox, axis int) bool{
	return self.Max()[(axis+2)%3] <= nearby.Min()[(axis+2)%3] || self.Min()[(axis+2)%3] >= nearby.Max()[(axis+2)%3] ||
	self.Max()[(axis+4)%3] <= nearby.Min()[(axis+4)%3] || self.Min()[(axis+4)%3] >= nearby.Max()[(axis+4)%3]
}

// floatPlanesBetween returns each integer boundary crossed between a and b.
func FloorFloatBetween(a, b float64) iter.Seq[float64] {
	if a > b {
		a, b = b, a
	}
	return func(yield func(float64) bool) {
		for i := math.Floor(a); i <= b; i++ {
			if !yield(i){
				return 
			}
		}
	}
}

// returns a Vec3 that is the smallest(with minimun len) is needed to apply to self to no longer intersect with nearby
func MinVec3InBoundWithinBBox(self, nearby, bound cube.BBox, selfPos mgl64.Vec3) (vec3 mgl64.Vec3, in bool){
	var dx, dy, dz float64 = 0, 0, 0
	var xDir, yDir, zDir float64 = 1, 1, 1
	in = true
	dF := func (axis int, dv, dir *float64) bool{
		if diff := self.Max()[axis] - nearby.Max()[axis]; diff >= 0 && diff < self.Max()[axis] - self.Min()[axis]{
			*dv = diff
		} 
		if diff := nearby.Min()[axis] - self.Min()[axis]; 
		diff >= 0 && diff < self.Max()[axis] - self.Min()[axis] && (diff < *dv || *dv == 0){
			*dv = diff
			*dir = -1
		}
		return *dv == 0
	}
	if dF(0, &dx, &xDir){
		return
	}
	if dF(1, &dy, &yDir){
		return
	}
	if dF(2, &dz, &zDir){
		return
	}
	outOfBound := func (axis int, dv float64, dir *float64) bool{
		if !(selfPos[axis] + dv < bound.Max()[axis] && bound.Min()[axis] <= selfPos[axis] + dv){
			*dir = 0
			return true
		}
		return false
	}
	if outOfBound(0, dx, &xDir) && outOfBound(0, dy, &yDir) && outOfBound(0, dz, &zDir){
		return vec3, false
	}
	eliminate := func (aDir, bDir, da, db *float64){
		if *aDir != 0 && *bDir != 0{
			if *da < *db && *da > 0{
				*db = 0
			}else if *db < *da && *db > 0{
				*da = 0
			}
		}
	}
	eliminate(&xDir, &yDir, &dx, &dy)
	eliminate(&xDir, &zDir, &dx, &dz)
	eliminate(&yDir, &zDir, &dy, &dz)
	vec3 = mgl64.Vec3{dx*xDir, dy*yDir, dz*zDir}
	return
}