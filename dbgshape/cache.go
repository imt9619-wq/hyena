package dbgshape

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft"
)

const (
	Line uint8 = iota
	Box
)

type AShape struct{
	Shape      uint8
    StartPoint mgl32.Vec3
    EndPoint   mgl32.Vec3
}

type ShapeCache struct {
	idToShape   map[uuid.UUID]AShape
    idToShapeMu *sync.Mutex
    conn        *minecraft.Conn
    closeOnce   *sync.Once
    closed      chan struct{}
    addDq       chan *shapeToDel
    dq          deadlineQueue
    oldTimer    *time.Timer
}

func NewShapeCache() *ShapeCache{
	s :=  &ShapeCache{
		idToShapeMu: &sync.Mutex{},
		idToShape: make(map[uuid.UUID]AShape, 64),
		closeOnce: &sync.Once{},
		closed: make(chan struct{}),
		addDq: make(chan *shapeToDel, 10),
		dq: make(deadlineQueue, 0, 10),
		oldTimer: nil,
	}
	go s.runDeadlineCountdown()
	return s
}

func (s *ShapeCache) AddDeadlineShape(conn *minecraft.Conn, shape AShape, d time.Duration){
	select{
	case <- s.closed:
		return
	default:
		s.conn = conn
		id := uuid.New()
		s.AddShapeOnId(conn, shape, id)
		s.addDq <- &shapeToDel{
			id: id,
			delTime: time.Now().Add(d),
		}
	}
}

func (s *ShapeCache) AddShape(conn *minecraft.Conn, shape AShape){
	s.AddShapeOnId(conn, shape, uuid.New())
} 

func (s *ShapeCache) AddShapeOnId(conn *minecraft.Conn, shape AShape, id uuid.UUID){
	s.idToShapeMu.Lock()
	defer s.idToShapeMu.Unlock()
	if _, ok := s.idToShape[id]; ok{
		fmt.Printf("Tried to add an existing debug shape")
		return
	}
	s.idToShape[id] = shape
	conn.WritePacket(&DebugShape{
		Opts: Add,
		ShapeID: id,
		Shape: shape,
	})
}

func (s *ShapeCache) DeleteShapeWithId(conn *minecraft.Conn, id uuid.UUID) bool{
	s.idToShapeMu.Lock()
	defer s.idToShapeMu.Unlock()
	if _, ok := s.idToShape[id]; !ok{
		return false
	}
	conn.WritePacket(&DebugShape{
		Opts: Delete,
		ShapeID: id,
	})
	delete(s.idToShape, id)
	return true
}

func (s *ShapeCache) DeleteShapeWithShape(conn *minecraft.Conn, shape AShape){
	s.idToShapeMu.Lock()
	defer s.idToShapeMu.Unlock()
	for id, sh := range s.idToShape{
		if sh == shape{
			conn.WritePacket(&DebugShape{
				Opts: Delete,
				ShapeID: id,
			})
			delete(s.idToShape, id)
		}
	}
}

func (s *ShapeCache) IdToShape() map[uuid.UUID]AShape{
	return s.idToShape
}

func (s *ShapeCache) Close(){
	s.closeOnce.Do(func() {
		close(s.closed)
	})
}