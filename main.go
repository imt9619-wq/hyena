package main

import (
	"context"
	"fmt"

	"github.com/imt9619-wq/hyena/manager"
	"github.com/imt9619-wq/hyena/manager/handler"
)

func main() {
	mgr := manager.DefaultConfig().New(context.Background())

	acc, ok := mgr.AccountsByTag()["ms_token_cache"]
	if !ok {
		fmt.Println("no account found: add token JSON files to the tokens/ folder")
		return
	}

	go func() { 
		// play.venitymc.com:19132
		// 127.0.0.1:19135
		if err := acc.JoinServer("play.venitymc.com:19132", handler.DefaultHandler{}); err != nil {
			fmt.Println(err)
		}
	}()

	mgr.WaitTilClose()
}
