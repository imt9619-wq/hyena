package utils

import (
	"encoding/binary"
	"hash/fnv"
	"iter"
	"math"
	"slices"
	"sync"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

func BlockInBBox(bbox cube.BBox, bs world.BlockSource) iter.Seq2[cube.Pos, world.Block]{
	return func(yield func(cube.Pos, world.Block) bool) {
		for blockPos := range blockPositionsInBBox(bbox){
			if !yield(blockPos, bs.Block(blockPos)){
				return 
			}
		}
	}
}

func BBoxMinMaxOnFace(box cube.BBox, face cube.Face) (mgl64.Vec3, mgl64.Vec3){
	switch face{
	case cube.FaceUp:
		return SetVec3AxisTo(box.Min(), 1, box.Max()[1]), box.Max()
	case cube.FaceDown:
		return box.Min(), SetVec3AxisTo(box.Max(), 1, box.Min()[1])
	case cube.FaceNorth:
		return box.Min(), SetVec3AxisTo(box.Max(), 2, box.Min()[2])
	case cube.FaceEast:
		return SetVec3AxisTo(box.Min(), 0, box.Max()[0]), box.Max()
	case cube.FaceSouth:
		return SetVec3AxisTo(box.Min(), 2, box.Max()[2]), box.Max()
	default:
		return box.Min(), SetVec3AxisTo(box.Max(), 0, box.Min()[0])
	}
}

func BBoxOnBBoxFaceWithThreshold(self cube.BBox, face cube.Face, threshold float64) cube.BBox{
	min, max := self.Min(), self.Max()
	switch face{
	case cube.FaceUp:
		min[1] = max[1]
		max[1] += threshold
	case cube.FaceDown:
		max[1] = min[1]
		min[1] -= threshold
	case cube.FaceNorth:
		max[2] = min[2]
		min[2] -= threshold
	case cube.FaceEast:
		min[0] = max[0]
		max[0] += threshold
	case cube.FaceSouth:
		min[2] = max[2]
		max[2] += threshold
	default:
		max[0] = min[0]
		min[0] -= threshold
	}
	return cube.Box(min[0], min[1], min[2], max[0], max[1], max[2])
}

func TinyBBoxOnBBoxFace(self cube.BBox, face cube.Face) cube.BBox{
	return BBoxOnBBoxFaceWithThreshold(self, face, ProbeOffset)
}

func Box(vec1, vec2 mgl64.Vec3) cube.BBox{
	return cube.Box(vec1[0],
					vec1[1],
					vec1[2],
					vec2[0],
					vec2[1],
					vec2[2],
					)
}

func IsBoxFlat(box cube.BBox) bool{
	return box.Height() == 0 || box.Width() == 0 || box.Length() == 0
}

func BBoxIntersectsSolid(bs BlockSourse, pBBox cube.BBox) bool {
	for _, blockBox := range SweptBBoxInBBox(pBBox, bs) {
		if pBBox.IntersectsWith(blockBox) {
			return true
		}
	}
	return false
}

func SweptBBoxInBBox(bbox cube.BBox, bs BlockSourse) iter.Seq2[cube.Pos, cube.BBox]{
	return func(yield func(cube.Pos, cube.BBox) bool) {
		for pos := range blockPositionsInBBox(bbox) {
			model, _ := bs.BlockModel(pos, 0)
			if model == nil {
				continue
			}
			for _, bbox := range BBoxes(model, pos, bs){
				if !yield(pos, bbox){
					return 
				}
			}
			
		}
	}
}

func blockPositionsInBBox(bbox cube.BBox) iter.Seq[cube.Pos]{
	min := bbox.Min()
	max := bbox.Max()
	
	return func(yield func(cube.Pos) bool) {
		for x := int(math.Floor(min[0])); x <= int(math.Floor(max[0])); x++ {
			for y := int(math.Floor(min[1])); y <= int(math.Floor(max[1])); y++ {
				for z := int(math.Floor(min[2])); z <= int(math.Floor(max[2])); z++ {
					if !yield(cube.Pos{x, y, z}){
						return 
					}
				}
			}
		}
	}
}

func BBoxes(model world.BlockModel, pos cube.Pos, s world.BlockSource) []cube.BBox{
	return TranslateBBoxes(model.BBox(pos, s), pos.Vec3())
}

func TranslateBBoxes(bboxes []cube.BBox, by mgl64.Vec3) []cube.BBox{
	for i, bbox := range bboxes{
		bboxes[i] = bbox.Translate(by)
	}
	return bboxes
}

func BBoxHash(bboxes []cube.BBox) uint64{
	hasher := fnv.New64a()
	var buf [8]byte
	for _, bbox := range bboxes{
		for i := range 3{
			binary.LittleEndian.PutUint64(buf[:], math.Float64bits(RoundFloat(bbox.Min()[i], 3)))
			hasher.Write(buf[:])
			binary.LittleEndian.PutUint64(buf[:], math.Float64bits(RoundFloat(bbox.Max()[i], 3)))
			hasher.Write(buf[:])
		}
	}
	return hasher.Sum64()
}

type ReverseBBoxCache struct{
	rmu *sync.RWMutex
	BBoxesToReverseBBoxes map[uint64][]cube.BBox
}

var reverseBBoxCache *ReverseBBoxCache = &ReverseBBoxCache{
	rmu: &sync.RWMutex{},
	BBoxesToReverseBBoxes: make(map[uint64][]cube.BBox, 200),
}

func ReverseBBoxOfBBoxes(bboxes []cube.BBox) []cube.BBox{
	if len(bboxes) == 0{
		return []cube.BBox{Box(mgl64.Vec3{}, mgl64.Vec3{1, 1, 1})}
	}
	offset := bboxes[0].Min()
	for _, bbox := range bboxes{
		for axis := range 3{
			offset[axis] = min(offset[axis], bbox.Min()[axis])
		}
	}
	return TranslateBBoxes(getReverseBB(TranslateBBoxes(bboxes, offset.Mul(-1))), offset)
}

func getReverseBB(normBBoxes []cube.BBox) []cube.BBox{
	key := BBoxHash(normBBoxes)
	reverseBBoxCache.rmu.RLock()
	reverseBB, ok := reverseBBoxCache.BBoxesToReverseBBoxes[key]
	result := slices.Clone(reverseBB)
	reverseBBoxCache.rmu.RUnlock()
	if ok{
		return result
	}
	reverseBB = IntersectionOfBoundWithComplementOfBBoxesUnion(Box(mgl64.Vec3{}, mgl64.Vec3{1, 1, 1}), normBBoxes)
	reverseBBoxCache.rmu.Lock()
	reverseBBoxCache.BBoxesToReverseBBoxes[key] = reverseBB
	result = slices.Clone(reverseBB)
	reverseBBoxCache.rmu.Unlock()
	return result
}

func ReverseBBoxes(model world.BlockModel, pos cube.Pos, s world.BlockSource) []cube.BBox{
	return TranslateBBoxes(getReverseBB(model.BBox(pos, s)), pos.Vec3())
}

func StickBBox(b1, b2 cube.BBox) (cube.BBox, bool) {
	forceStick := Box(b1.Min(), b2.Max())
	if b1.Height() == b2.Height() && b1.Width() == b2.Width() {
		if forceStick.Length() == 0 {
			return Box(b2.Min(), b1.Max()), true
		}
		return forceStick, true
	}
	if b1.Length() == b2.Length() && b1.Width() == b2.Width() {
		if forceStick.Height() == 0 {
			return Box(b2.Min(), b1.Max()), true
		}
		return forceStick, true
	}
	if b1.Height() == b2.Height() && b1.Length() == b2.Length() {
		if forceStick.Width() == 0 {
			return Box(b2.Min(), b1.Max()), true
		}
		return forceStick, true
	}
	return cube.BBox{}, false
}

func BBoxIntersection(b1, b2 cube.BBox) (cube.BBox, bool) {
	if !b1.IntersectsWith(b2) {
		return cube.BBox{}, false
	}
	minVec, maxVec := mgl64.Vec3{}, mgl64.Vec3{}
	for axis := range 3 {
		minVec[axis] = max(b1.Min()[axis], b2.Min()[axis])
		maxVec[axis] = min(b1.Max()[axis], b2.Max()[axis])
	}
	return Box(minVec, maxVec), true
}

func IntersectionOfBoundWithComplementOfBBoxesUnion(bound cube.BBox, bboxes []cube.BBox) []cube.BBox {
	if len(bboxes) == 0 {
		return []cube.BBox{bound}
	}
	getBoxCutOnAxisMax := func(axis int, bbox, bound cube.BBox) cube.BBox {
		return Box(bound.Min(), SetVec3AxisTo(bound.Max(), axis, max(bbox.Min()[axis], bound.Min()[axis])))
	}
	getBoxCutOnAxisMin := func(axis int, bbox, bound cube.BBox) cube.BBox {
		return Box(bound.Max(), SetVec3AxisTo(bound.Min(), axis, min(bound.Max()[axis], bbox.Max()[axis])))
	}
	cutOnFourDir := func(axis int, bbox, b cube.BBox) cube.BBox {
		b = getBoxCutOnAxisMax((axis+2)%3, bbox, b)
		b = getBoxCutOnAxisMin((axis+2)%3, bbox, b)
		b = getBoxCutOnAxisMax((axis+4)%3, bbox, b)
		return getBoxCutOnAxisMin((axis+4)%3, bbox, b)
	}
	resultBanch := map[int][]cube.BBox{
		0: make([]cube.BBox, 0, len(bboxes)*4),
		1: make([]cube.BBox, 0, len(bboxes)*4),
	}
	currBanch := 0
	complement := make([]cube.BBox, 0, 6)
	resultBanch[currBanch] = append(resultBanch[currBanch], bound)
	for _, bbox := range bboxes {
		complement = complement[:0]
		for axis := range 3 {
			bMax := getBoxCutOnAxisMax(axis, bbox, bound)
			for _, comBB := range complement{
				if comBB.Max()[axis] == bMax.Max()[axis] {
					bMax = cutOnFourDir(axis, comBB, bMax)
				}
				if IsBoxFlat(bMax) {
					break
				}
			}
			if !IsBoxFlat(bMax){
				complement = append(complement, bMax)
			}
			bMin := getBoxCutOnAxisMin(axis, bbox, bound)
			for _, comBB := range complement {
				if comBB.Min()[axis] == bMin.Min()[axis] {
					bMin = cutOnFourDir(axis, comBB, bMin)
				}
				if IsBoxFlat(bMin) {
					break
				}
			}
			if !IsBoxFlat(bMin){
				complement = append(complement, bMin)
			}
		}
		// get intersection
		currBanch = (currBanch + 1) % 2
		for _, otherComBox := range resultBanch[(currBanch+1)%2] {
			for _, comBB := range complement {
				intersection, hasIntersected := BBoxIntersection(otherComBox, comBB)
				if !hasIntersected{
					continue
				}
				resultBanch[currBanch] = append(resultBanch[currBanch], intersection)
			}
		}
		resultBanch[(currBanch+1)%2] = resultBanch[(currBanch+1)%2][:0]
	}
	if len(resultBanch[currBanch]) <= 1{
		return resultBanch[currBanch]
	}
	// stick the bboxes to a larger piece
	for{
		oldLen := len(resultBanch[currBanch])
		oldBanch := currBanch
		currBanch = (currBanch+1)%2
		for len(resultBanch[oldBanch]) != 0{
			i := len(resultBanch[oldBanch])-1
			appendBB := resultBanch[oldBanch][i]
			for j := 0; j < i; j++{
				if mergeBB, ok := StickBBox(resultBanch[oldBanch][j], appendBB); ok{
					appendBB = mergeBB
					resultBanch[oldBanch] = slices.Delete(resultBanch[oldBanch], j, j+1)
					i--
					break
				}
			}
			resultBanch[currBanch] = append(resultBanch[currBanch], appendBB)
			resultBanch[oldBanch] = resultBanch[oldBanch][:i]
		}
		if len(resultBanch[currBanch]) == oldLen{
			break
		}
	}
	return resultBanch[currBanch]
}

func B1InB2(b1, b2 cube.BBox) bool{
	if b2.Max()[0] >= b1.Max()[0] && b2.Min()[0] <= b1.Min()[0]{
		if b2.Max()[1] >= b1.Max()[1] && b2.Min()[1] <= b1.Min()[1]{
			return b2.Max()[2] >= b1.Max()[2] && b2.Min()[2] <= b1.Min()[2]
		}
	}
	return false
}