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
		// as.venity.net:19132
		// 127.0.0.1:19134
		// 127.0.0.1:19135
		closed, err := acc.JoinServer("as.venity.net:19132", pathfind.NewPathFindHandler("ruMEme"))
		if err != nil {
			fmt.Println(err)
			return	
		}
		<-closed
	
	}()
		
	mgr.WaitTilClose()
	fmt.Printf("Ran for %v seconds\n", time.Since(now).Seconds())
}
