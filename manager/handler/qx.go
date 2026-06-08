package handler

// Qx idea is basically the same as Tx for a dragonfly world, but 
// instead of Tx for a world, we have Tx for a gameState, we call 
// it Qx instead of Tx so we wont confuse it with the Tx from dragonfly
type Qx struct {
	gs *gameState
	closed bool
}

func (qx *Qx) close() {
	qx.closed = true
}

type QueueFunc func(*Qx)

type queueTransition struct {
	c chan struct{}
	f QueueFunc
}

func (gs *gameState) startRunningQueue(c *Connection) {
	go func ()  {
		for {
			select {
			case <-c.closed:
				return
			case q := <-gs.queue:
				q.Run(gs)
			}
		}
	}()
}

func (gs *gameState) Exec(f QueueFunc) chan struct{} {
	ch := make(chan struct{})
	gs.queue <- &queueTransition{c: ch, f: f}
	return ch
}

func (q *queueTransition) Run(gs *gameState) {
	qx := &Qx{gs: gs, closed: false}
	q.f(qx)
	qx.close()
	close(q.c)
}