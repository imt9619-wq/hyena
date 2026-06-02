package main

import (
	"context"
	"fmt"

	"github.com/imt9619-wq/hyena/manager"
)

func main() {
	mgr := manager.DefaultConfig().New(context.Background())

	clt, ok := mgr.Clients()[0]
	if !ok {
		fmt.Println("no clients found: add token JSON files to the tokens/ folder")
		return
	}

	go func() {
		if _, err := clt.JoinServer("play.venitymc.com:19132"); err != nil {
			fmt.Println(err)
		}

	}()

	mgr.WaitTilClose()
}
