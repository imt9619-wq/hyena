package pathfind

import (
	"container/heap"
	"slices"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
)

func (e *pathRunExecutor) astar(start, end mgl64.Vec3) []nodePos{
	startNode := &node{position: nodePosByVec3(start)}
	endNode := &node{position: nodePosByVec3(end)}
	if equal(startNode, endNode){
		return nil
	}

	elsSize := elsimate(startNode.blockPos(), endNode.blockPos())
	openList := make(Set, 0, elsSize/2)
	heap.Push(&openList, startNode)
	closeList := make(map[cube.Pos]struct{}, elsSize)
	for openList.Len() > 0 {
		currNode := heap.Pop(&openList).(*node)
		closeList[currNode.blockPos()] = struct{}{}
		if equal(currNode, endNode){
			path := make([]nodePos, 0, elsSize/8)
			for lastNode := currNode; !equal(lastNode, startNode); lastNode = lastNode.parent{
				path = append(path, lastNode.position)
			}
			path = append(path, startNode.position)
			slices.Reverse(path)
			return path
		}

		
	}
	return nil
}

func elsimate(start, end cube.Pos) int{
	return abs(start.Sub(end)[0] * start.Sub(end)[1] * start.Sub(end)[2]/2)
}

func abs(n int) int{
	if n < 0{
		return -n
	}
	return n
}