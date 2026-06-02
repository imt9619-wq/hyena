package main

import (
	"context"
	"fmt"

	"github.com/imt9619-wq/hyena/manager"
	"github.com/imt9619-wq/hyena/manager/handler"
)

func main() {
	mgr := manager.DefaultConfig().New(context.Background())

	clt, ok := mgr.ClientsByTag()["ms_token_cache"]
	if !ok {
		return
	}

	go func() {
		err := clt.JoinServer("play.venitymc.com:19132", handler.DefaultHandler{});
		if err != nil {
			fmt.Println(err)
		}
	}()

	mgr.WaitTilClose()
}
