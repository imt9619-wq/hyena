package main

import (
	"context"
	"fmt"
	"time"

	"github.com/imt9619-wq/hyena/manager"
	"github.com/imt9619-wq/hyena/pathfind"
)

func main() {
	mgr := manager.DefaultConfig().New(context.Background())
	now := time.Now()
	acc, ok := mgr.AccountsByTag()["ms_token_cache"]
	if !ok {
		fmt.Println("no account found: add token JSON files to the tokens/ folder")
		return
	}

	go func() { 
		// play.venitymc.com:19132
		// 127.0.0.1:19134
		// 127.0.0.1:19135
		for {
			closed, err := acc.JoinServer("play.venitymc.com:19132", pathfind.NewPathHandler())
			if err != nil {
				fmt.Println(err)
				return	
			}
			<-closed
		}
	}()
		
	mgr.WaitTilClose()
	fmt.Printf("Ran for %v seconds\n", time.Since(now).Seconds())
}
