package pathfind

import (

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
)

type node struct {
	g, h, f  float64
	position nodePos
	parent   *node
}

func equal(node1, node2 *node) bool{
	return node1.position.nodeBlock == node2.position.nodeBlock
}

func (n *node) blockPos() cube.Pos{
	return n.position.nodeBlock
}

type nodePos struct{
	nodeBlock cube.Pos
	nodePos mgl64.Vec3
}

func nodePosByVec3(pos mgl64.Vec3) nodePos{
	return nodePos{
		nodeBlock: cube.PosFromVec3(pos),
		nodePos: pos,
	}
}

func (p nodePos) nodeBound() cube.BBox{
	return cube.Box(p.nodePos[0], p.nodePos[1], p.nodePos[2], p.nodePos[0]+1, p.nodePos[1]+1, p.nodePos[2]+1)
}

func (p nodePos) outOfNodeBound(pos mgl64.Vec3) bool{
	return !(p.nodeBound().Min()[0] <= pos[0] && pos[0] < p.nodeBound().Max()[0] && 
	p.nodeBound().Min()[1] <= pos[1] && pos[1] < p.nodeBound().Max()[1] &&
	p.nodeBound().Min()[2] <= pos[2] && pos[2] < p.nodeBound().Max()[2])
}

func (p *nodePos) setNodePos(pos mgl64.Vec3) bool{
	if p.outOfNodeBound(pos){
		return false
	}
	p.nodePos = pos
	return true
}
