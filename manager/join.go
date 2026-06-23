package manager

import (
	"errors"

	"github.com/imt9619-wq/hyena/manager/handler"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/auth"
)

func (a *Account) JoinServer(serverAddress string, h handler.Handler) error {
	if h == nil{
		h = handler.DefaultHandler{}
	}
	src := auth.RefreshTokenSource(a.config.Token)
	serverConn, err := minecraft.Dialer{
		TokenSource:       src,
		EnableClientCache: false,
	}.Dial("raknet", serverAddress)
	if err != nil {
		return err
	}

	if err = serverConn.DoSpawn(); err != nil {
		serverConn.Close()
		return err
	}

	session := a.newSession(serverConn, h)

	select {
	case a.sessionQueue <- session:
	case <-a.managerClosed:
		session.markClosed()
		return errors.New("manager is closed")
	}

	return nil
}
