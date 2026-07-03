package dbgshape

import (
	"container/heap"
	"time"

	"github.com/google/uuid"
)

type shapeToDel struct{
	id uuid.UUID
	delTime time.Time
}

type deadlineQueue []*shapeToDel

func (s *ShapeCache) runDeadlineCountdown() {
	var timerC <-chan time.Time

	resetTimer := func(t time.Time) {
		d := time.Until(t)
		if d < 0 {
			d = 0
		}
		if s.oldTimer != nil {
			if !s.oldTimer.Stop() {
				select {
				case <-s.oldTimer.C:
				default:
				}
			}
		}
		s.oldTimer = time.NewTimer(d)
		timerC = s.oldTimer.C
	}

	stopTimer := func() {
		if s.oldTimer != nil {
			if !s.oldTimer.Stop() {
				select {
				case <-s.oldTimer.C:
				default:
				}
			}
			s.oldTimer = nil
		}
		timerC = nil
	}

	for {
		select {
		case <-s.closed:
			stopTimer()
			return
		case st, ok := <-s.addDq:
			if !ok {
				continue
			}
			wasEmpty := s.dq.Len() == 0
			var oldNext *shapeToDel
			if !wasEmpty {
				oldNext = s.dq[0]
			}
			heap.Push(&s.dq, st)
			if wasEmpty || st.delTime.Before(oldNext.delTime) {
				resetTimer(s.dq[0].delTime)
			}
		case <-timerC:
			if s.dq.Len() == 0 {
				stopTimer()
				continue
			}
			next := heap.Pop(&s.dq).(*shapeToDel)
			s.DeleteShapeWithId(s.conn, next.id)
			if s.dq.Len() == 0 {
				stopTimer()
				continue
			}
			resetTimer(s.dq[0].delTime)
		}
	}
}

func (dq deadlineQueue) Len() int { 
    return len(dq) 
}

func (dq deadlineQueue) Less(i, j int) bool {
    return dq[i].delTime.Before(dq[j].delTime)
}

func (dq deadlineQueue) Swap(i, j int) {
    dq[i], dq[j] = dq[j], dq[i]
}

func (dq *deadlineQueue) Push(x any) {
    *dq = append(*dq, x.(*shapeToDel))
}

func (dq *deadlineQueue) Pop() any {
    old := *dq
    n := len(old)
    item := old[n-1]
    old[n-1] = nil 
    *dq = old[0 : n-1]
    return item
}