package pathfind

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type event struct {
	EventType
	payLoad any
}

type EventType uint

const (
	EventPlayerPosChanged EventType = iota
	EventBlockChanged
	EventChunkChanged
	EventSubChunkChanged
	EventGoalChanged
	EventShouldMove
)

type payLoadShouldMove struct{
	shouldMove bool
}

type payLoadSubChunkChanged struct{
	pos protocol.SubChunkPos
}

type payLoadBlockChanged struct{
	pos protocol.BlockPos
}

type payLoadChunkChanged struct{
	pos protocol.ChunkPos 
}

type payLoadPlayerPosChanged struct {
	before, after mgl32.Vec3
}

type payLoadGoalChanged struct {
	goal mgl32.Vec3
}

func (e *pathRunExecutor) newEvent(eventType EventType, payLoad any) {
	e.events <- event{
		EventType: eventType,
		payLoad:   payLoad,
	}
}