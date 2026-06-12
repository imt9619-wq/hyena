package sim

import (
	"math"
	"slices"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
)

const (
	playerHeight = float32(1.8)
	playerWidth  = float32(0.6)
	maxStepHeight  = 0.6
)

type collideOffsets struct {
	closestBox map[int]*leastBoxOffset
	indies     []int
}

type leastBoxOffset struct {
	leastOffset       float32
	leastOffsetBBoxes []cube.BBox
}

func (m *Movement) checkCollision() {
	p := m.session.Player
	posBeforeTick := p.Position.Sub(p.Velocity)
	bboxBeforeTick := playerBBox(posBeforeTick)
	c := m.walkPath(bboxBeforeTick)
	for _, index := range c.indies {
		p.Position[index] = p.Position[index] - p.Velocity[index] + c.closestBox[index].leastOffset
		if (index == 1 && p.Velocity[1] < c.closestBox[index].leastOffset && p.Velocity[1] < 0) || c.closestBox[index].leastOffset == 0 {
			p.OnGround = true
		}
		if p.Velocity[index] == c.closestBox[index].leastOffset && c.closestBox[index].leastOffset != 0 {
			p.OnGround = false
		}
		if p.Velocity[index] != c.closestBox[index].leastOffset || c.closestBox[index].leastOffset == 0 {
			p.Velocity[index] = 0
		}
	}
}

func playerBBox(pos mgl32.Vec3) cube.BBox {
	halfW := playerWidth / 2
	return cube.Box(
		float64(pos[0]-halfW),
		float64(pos[1]),
		float64(pos[2]-halfW),
		float64(pos[0]+halfW),
		float64(pos[1]+playerHeight),
		float64(pos[2]+halfW),
	)
}

func (m *Movement) walkPath(pBBox cube.BBox) collideOffsets {
	c := m.collideWithOffsets(pBBox, m.session.Player.Velocity)
	return m.tryStepUp(pBBox, c)
}

func (m *Movement) tryStepUp(pBBox cube.BBox, c collideOffsets) collideOffsets {
	p := m.session.Player
	if !((p.OnGround && p.Velocity[1] == 0) && !slices.Contains(c.indies, 1)) {
		return c
	}
	if len(c.indies) == 0 {
		return c
	}

	index := c.indies[0]
	offset := c.closestBox[index].leastOffset
	if len(c.indies) > 1 || p.Velocity[index] == offset {
		return c
	}

	bboxWithOffset := c.closestBox[index].leastOffsetBBoxes
	var stepHeight float64
	for _, bBBox := range bboxWithOffset {
		currentStepHeight := bBBox.Max()[1] - pBBox.Min()[1]
		if currentStepHeight > stepHeight {
			stepHeight = currentStepHeight
		}
		if currentStepHeight > maxStepHeight {
			return c
		}
	}
	if stepHeight == 0 {
		return c
	}

	walkStairDelta := p.Velocity
	walkStairDelta[1] = float32(stepHeight)
	upC := m.collideWithOffsets(pBBox, walkStairDelta)
	if upC.closestBox[1].leastOffset < float32(stepHeight) {
		return c
	}

	raisedBBox := pBBox.Translate(mgl64.Vec3{0, stepHeight, 0})
	if m.bboxIntersectsSolid(raisedBBox) {
		return c
	}

	horizontalVel := p.Velocity
	horizontalVel[1] = 0
	stepC := m.collideWithOffsets(raisedBBox, horizontalVel)
	finalBBox := raisedBBox.Translate(mgl64.Vec3{
		float64(stepC.closestBox[0].leastOffset),
		0,
		float64(stepC.closestBox[2].leastOffset),
	})
	if m.bboxIntersectsSolid(finalBBox) {
		return c
	}
	stepC.closestBox[1].leastOffset = float32(stepHeight)
	return m.firstCollideAxis(stepC.closestBox, p.Velocity)
}

func (m *Movement) firstCollideAxis(closestBox map[int]*leastBoxOffset, deltas mgl32.Vec3) collideOffsets {
	iToVRadio := make([]float32, 0, 3)
	c := collideOffsets{closestBox: closestBox}
	for i := 0; i < 3; i++ {
		radio := float32(1)
		if deltas[i] != 0 {
			radio = closestBox[i].leastOffset / deltas[i]
		}
		iToVRadio = append(iToVRadio, radio)
	}
	minOffsetRatio := min(iToVRadio[0], iToVRadio[1], iToVRadio[2])
	c.indies = make([]int, 0, 2)
	for i, radio := range iToVRadio {
		if radio == minOffsetRatio {
			c.indies = append(c.indies, i)
		}
	}
	return c
}

func newLeastBoxOffsetMap(deltas mgl32.Vec3) map[int]*leastBoxOffset {
	closestBox := make(map[int]*leastBoxOffset, 3)
	for i := 0; i < 3; i++ {
		closestBox[i] = &leastBoxOffset{
			leastOffset:       deltas[i],
			leastOffsetBBoxes: make([]cube.BBox, 0, 3),
		}
	}
	return closestBox
}

func isCloserToZero(currentOffset float64, leastOffset float64) float64 {
	return math.Abs(leastOffset) - math.Abs(currentOffset)
}

func (m *Movement) collideWithOffsets(pBBoxBeforeTick cube.BBox, deltas mgl32.Vec3) collideOffsets {
	world := m.session.BlockMap
	closestBox := newLeastBoxOffsetMap(deltas)
	var xSolid, ySolid, zSolid bool
	var xOffset, yOffset, zOffset float64
	xFace, yFace, zFace := cube.FaceWest, cube.FaceDown, cube.FaceNorth
	if deltas[0] > 0 {
		xFace = cube.FaceEast
	}
	if deltas[1] > 0 {
		yFace = cube.FaceUp
	}
	if deltas[2] > 0 {
		zFace = cube.FaceSouth
	}
	for pos := range m.intersection(pBBoxBeforeTick, deltas) {
		model, ok := world.BlockModel(pos, 0)
		if !ok {
			continue
		}
		bBoxes := model.BBox(pos, world)
		xSolid = model.FaceSolid(pos, xFace, world)
		ySolid = model.FaceSolid(pos, yFace, world)
		zSolid = model.FaceSolid(pos, zFace, world)
		for _, bBBox := range bBoxes {
			if xSolid {
				xOffset = pBBoxBeforeTick.XOffset(bBBox, float64(closestBox[0].leastOffset))
				closestBox[0].appendOffset(xOffset, bBBox)
			}
			if ySolid {
				yOffset = pBBoxBeforeTick.YOffset(bBBox, float64(closestBox[1].leastOffset))
				closestBox[1].appendOffset(yOffset, bBBox)
			}
			if zSolid {
				zOffset = pBBoxBeforeTick.ZOffset(bBBox, float64(closestBox[2].leastOffset))
				closestBox[2].appendOffset(zOffset, bBBox)
			}
		}
	}
	return m.firstCollideAxis(closestBox, deltas)
}

func (lb *leastBoxOffset) appendOffset(offset float64, bBBox cube.BBox) {
	if isCloserToZero(offset, float64(lb.leastOffset)) > 0 {
		clear(lb.leastOffsetBBoxes)
		lb.leastOffset = float32(offset)
	}
	if offset == float64(lb.leastOffset) {
		lb.leastOffsetBBoxes = append(lb.leastOffsetBBoxes, bBBox)
	}
}
