package dbgshape

import (
	"context"
	"fmt"
	"iter"
	"log/slog"
	"reflect"
	"sync"
	"unsafe"

	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/debug"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/google/uuid"
	"github.com/imt9619-wq/hyena/utils"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func DebugShapeListener(c server.UserConfig, log *slog.Logger) (server.Config, error) {
	conf, err := c.Config(log)
	if err != nil {
		return conf, err
	}
	conf.Listeners = []func(conf server.Config) (server.Listener, error){
		debugShapeListenerFunc(c),
	}
	return conf, nil
}

func debugShapeListenerFunc(c server.UserConfig) func(conf server.Config) (server.Listener, error) {
	return func(conf server.Config) (server.Listener, error) {
		cfg := minecraft.ListenConfig{
			MaximumPlayers:         conf.MaxPlayers,
			StatusProvider:         conf.StatusProvider,
			AuthenticationDisabled: conf.AuthDisabled,
			ResourcePacks:          conf.Resources,
			TexturePacksRequired:   conf.ResourcesRequired,
			Compression:            conf.Compression,
		}
		if conf.Log.Enabled(context.Background(), slog.LevelDebug) {
			cfg.ErrorLog = conf.Log.With("net origin", "gophertunnel")
		}
		l, err := cfg.Listen("raknet", c.Network.Address)
		if err != nil {
			return nil, fmt.Errorf("create minecraft listener: %w", err)
		}
		conf.Log.Info("Listener running.", "addr", l.Addr())
		return debugShapeListener{Listener: l}, nil
	}
}

type debugShapeListener struct {
	*minecraft.Listener
}

func (l debugShapeListener) Accept() (session.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return &debugShapeConn{
		Conn: conn.(session.Conn), 
		idMap: make(map[uuid.UUID]debug.Shape, 64), 
		idMapMu: &sync.Mutex{},
		}, nil
}

func (l debugShapeListener) Disconnect(conn session.Conn, reason string) error {
	if wrapped, ok := conn.(*debugShapeConn); ok {
		conn = wrapped.Conn
	}
	return l.Listener.Disconnect(conn.(*minecraft.Conn), reason)
}

type debugShapeConn struct {
	session.Conn
	srv *server.Server
	idMap map[uuid.UUID]debug.Shape
	idMapMu *sync.Mutex
}

func SetServerForConn(s *session.Session, srv *server.Server) {
	if srv == nil{
		return
	}
	if wrapped, ok := SessionDbgShapeConn(s); ok{
		wrapped.setServer(srv)
	}
}

func (c *debugShapeConn) setServer(srv *server.Server){
	c.srv = srv
}

func (c *debugShapeConn) ReadPacket() (packet.Packet, error) {
	for {
		pk, err := c.Conn.ReadPacket()
		if err != nil {
			return nil, err
		}
		if shape, ok := pk.(*DebugShape); ok {
			c.handleDebugShape(shape)
			continue
		}
		return pk, nil
	}
}

func (c *debugShapeConn) handleDebugShape(shape *DebugShape){
	if c.srv == nil{
		return
	}
	switch shape.Opts{
	case Add:
		c.addShape(shape.ShapeID, shape.Shape)
	case Delete:
		c.deleteShape(shape.ShapeID)
	}
}

func (c *debugShapeConn) addShape(shapeId uuid.UUID, shape AShape){
	c.idMapMu.Lock()
	_, ok := c.idMap[shapeId]
	c.idMapMu.Unlock()
	if ok{
		return
	}
	var s debug.Shape
	start := utils.Mgl32Vec3Tomgl64Vec3(shape.StartPoint)
	end := utils.Mgl32Vec3Tomgl64Vec3(shape.EndPoint)
	switch shape.Shape{
	case Box:
		s = &debug.Box{
			Position: start.Add(end).Mul(0.5),
			Bounds: end.Sub(start),
		}
	case Line:
		s = &debug.Line{
			Position: start,
			EndPosition: end,
		}
	default:
		return
	}
	if !c.execConnWorld(func(tx *world.Tx, e world.Entity) {
		for e := range tx.Players(){
			e.(*player.Player).AddDebugShape(s)
		}
	}){
		return
	}
	c.idMapMu.Lock()
	c.idMap[shapeId] = s
	c.idMapMu.Unlock()
}

func (c *debugShapeConn) deleteShape(shapeId uuid.UUID){
	c.idMapMu.Lock()
	s, ok := c.idMap[shapeId]
	c.idMapMu.Unlock()
	if !ok{
		return
	}
	if !c.execConnWorld(func(tx *world.Tx, e world.Entity) {
		for e := range tx.Players(){
			e.(*player.Player).RemoveDebugShape(s)
		}
	}){
		return
	}
	c.idMapMu.Lock()
	delete(c.idMap, shapeId)
	c.idMapMu.Unlock()
}

func (c *debugShapeConn) pID() (uuid.UUID, bool){
	parsedUUID, err := uuid.Parse(c.IdentityData().Identity)
	if err != nil {
		return parsedUUID, false
	}
	return parsedUUID, true
}

func (c *debugShapeConn) execConnWorld(f func(tx *world.Tx, e world.Entity)) bool{
	pid, ok := c.pID()
	if !ok{
		return false
	}
	e, ok := c.srv.Player(pid)
	if !ok{
		return false
	}
	if !e.ExecWorld(f){
		return false
	}
	return true
}

func SessionDbgShapeConn(s *session.Session) (*debugShapeConn, bool){
	if s == nil {
		return nil, false
	}
	reflectField := reflect.ValueOf(s).Elem().FieldByName("conn")
	if !reflectField.IsValid() {
		return nil, false
	}
	conn := reflect.NewAt(reflectField.Type(), unsafe.Pointer(reflectField.UnsafeAddr())).
		Elem().
		Interface().
		(session.Conn)
	wrapped, ok := conn.(*debugShapeConn)
	if !ok {
		return nil, false
	}
	return wrapped, true
}

func RemoveSessionShapes(s *session.Session, tx *world.Tx){
	d, ok := SessionDbgShapeConn(s)
	if !ok{
		return
	}
	for _, s := range d.Shapes(){
		for e := range tx.Players(){
			e.(*player.Player).RemoveDebugShape(s)
		}
	}
}

func (d *debugShapeConn) Shapes() iter.Seq2[uuid.UUID, debug.Shape]{
	return func(yield func(uuid.UUID, debug.Shape) bool) {
		d.idMapMu.Lock()
		defer d.idMapMu.Unlock()
		for id, s := range d.idMap{
			if !yield(id, s){
				return 
			}
		}
	}
}