package pathfind

import (
	"iter"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/imt9619-wq/hyena/game/input"
	"github.com/imt9619-wq/hyena/game/movements"
)

func moveTilBreak(ps *movements.PlayerState, input *input.Inputs) iter.Seq[*movements.OutMovement]{
	return func(yield func(*movements.OutMovement) bool) {
		if !yield(ps.DoMove(ps.SpiltInMovement(*input))){
			return 
		}
	}
}

type moveType interface{
	getedge(ps movements.PlayerState, to cube.Pos) edge
}

type edge struct {
	kind moveType 
	withState movements.AMovement
	to   cube.Pos
	cost int       // the heuristic for our astar is in ticks, <0 means not reachable
}

type runMove struct{}
func (runMove) getedge(ps movements.PlayerState, to cube.Pos) edge{	
	var input input.Inputs
	for out := range moveTilBreak(&ps, &input){
		_ = out
	}
	return edge{}
}