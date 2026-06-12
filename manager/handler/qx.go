package handler

// Qx is the serialised mutation context for gameState, similar to dragonfly's Tx.
type Qx struct {
	gs     *gameState
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