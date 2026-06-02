package manager

import (
	"errors"

	"github.com/google/uuid"
	"github.com/imt9619-wq/hyena/manager/handler"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/auth"
)

func (c *Client) JoinServer(serverAddress string, h handler.ConnHandler) (err error) {
	src := auth.RefreshTokenSource(c.conf.Token)
	serverConn, err := minecraft.Dialer{
		TokenSource:         src,
		EnableClientCache:   false,
	}.Dial("raknet", serverAddress)

	if err != nil {
		return
	}
	
	err = serverConn.DoSpawn()
	if err != nil {
		serverConn.Close()
		return
	}

	conn := &ClientConn{
		connBuf: &handler.ConnBuf{
			Conn: serverConn,
			H: h,
		},
		client: c,
		id: uuid.New(),
	}

	select {
	case c.outgoingConn <- conn:
	case <-c.managerClosed:
		serverConn.Close()
		return errors.New("manager is closed")
	}
	
	return
}


