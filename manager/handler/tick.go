package handler

import (
	"time"
)



func (cb *ConnBuf) tick(){
	defer cb.sc.flush()
	cb.movements.tick()
}



func (cb *ConnBuf) startTicking() {
	ticker := time.NewTicker(50 * time.Millisecond)
	
	go func() {
		for {
			select {
			case <-cb.closed:
				ticker.Stop()
				return
			case <-ticker.C:
				cb.tick()
			}
		}
	}()
	cb.tick()
}
