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

func SetVec3AxisTo(v mgl64.Vec3, axis int, to float64) mgl64.Vec3{
	v[axis] = to
	return v
}
