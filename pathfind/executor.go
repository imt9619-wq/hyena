package pathfind

import (
	"sync"
	"sync/atomic"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/imt9619-wq/hyena/game/blockmap"
	"github.com/imt9619-wq/hyena/game/movements"
	"github.com/imt9619-wq/hyena/manager/handler"
)

type pathRunExecutor struct {
	goal              mgl32.Vec3
    onPathCal         *atomic.Bool
    onPathRunExcution *atomic.Bool
    world             *blockmap.BlockMap
    events            chan event
    c                 *handler.Connection
	stateMu           *sync.RWMutex
	state             *movements.PlayerState
}

func (h *PathFindHandler) setPathRunExector(){
	h.executor = &pathRunExecutor{
		onPathCal: &atomic.Bool{},
		onPathRunExcution: &atomic.Bool{},
		world: h.world,
		events: make(chan event, 1024),
		c: h.c,
		stateMu: &sync.RWMutex{},
		state: movements.NewPlayerState(h.c.Conn, movements.NewMovement(h.world)),
	}
}

func (h *PathFindHandler) startPathFindRunner() {
	e := h.executor
	for {
		select {
		case <-h.c.Closed():
			h.shape.Close()
			return
		case pk := <-h.packets:
			if ph, ok := packetToPacketHandler[pk.ID()]; ok{
				ph.handle(h, pk)
			}
		case en := <- e.events:
			switch en.EventType{
			case EventGoalChanged:
				e.goal = en.payLoad.(payLoadGoalChanged).goal
			case EventShouldMove:
				if en.payLoad.(payLoadShouldMove).shouldMove{
					e.onPathCal.Store(true)
				}
			}
		default:
			if e.onPathCal.Load(){
				if e.goal.Sub(e.pos()).Len() < 1{
					e.onPathCal.Store(false)
					continue
				}
				e.calculatePath()
			}
		}
	}
}

func (e *pathRunExecutor) syncState(move *movements.AMovement){
	e.newEvent(EventPlayerPosChanged, payLoadPlayerPosChanged{
		before: e.state.Position,
		after: move.Position,
	})
	e.state.CopyMovement(move)
}

func (e *pathRunExecutor) pos() mgl32.Vec3{
	e.stateMu.RLock()
	defer e.stateMu.RUnlock()
	return e.state.Position
}

func (e *pathRunExecutor) calculatePath(){
	defer e.onPathCal.Store(false)

}
