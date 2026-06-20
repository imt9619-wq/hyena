package physics

import (
	"fmt"
	"iter"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/imt9619-wq/hyena/utils"
)

// floatPlanesBetween returns each integer boundary crossed between a and b.
func floatIntBetween(a, b float64) iter.Seq[float64] {
	if a > b {
		a, b = b, a
	}
	iterTimes := int(b-a)+1
	return func(yield func(float64) bool) {
		for i, j := a, 0; j <= iterTimes; j++{
			if !yield(i){
				return 
			}
			i++
			if i > b{
				i = b
			}
		}
	}
}

func AOffset(self, nearby cube.BBox, axis int, delta mgl64.Vec3) (offset float64, reachable bool){	
	reachable = false
	if delta[axis] == 0{
		offset = 0
		return
	}
	if delta[axis] > 0 && self.Max()[axis] <= nearby.Min()[axis]{
		offset = min(nearby.Min()[axis] - self.Max()[axis], delta[axis])
		fmt.Printf("Index: %d, offset: %v\n", axis, offset)
	}else if delta[axis] < 0 && self.Min()[axis] >= nearby.Max()[axis]{
		offset = max(nearby.Max()[axis] - self.Min()[axis], delta[axis])
		fmt.Printf("Index: %d, offset: %v\n", axis, offset)
	}else{
		return
	}
	var radio float64 = 1
	if delta[axis] != 0{
		radio = offset/delta[axis]
	}
	if !utils.OutOfPlane(self.Translate(delta.Mul(radio)), nearby, axis){
		if mgl64.FloatEqualThreshold(offset, 0, utils.Negligible){
			offset = 0
		}
		reachable = true
	}
	return
}

func bboxes(model world.BlockModel, pos cube.Pos, s world.BlockSource) []cube.BBox{
	blockBoxes := model.BBox(pos, s)
	for i, bbox := range blockBoxes{
		blockBoxes[i] = bbox.Translate(pos.Vec3())
	}
	return blockBoxes
}