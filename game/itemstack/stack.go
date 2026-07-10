package itemstack

import (
	_ "unsafe"
	
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type Stack struct{
	item.Stack
    instance *protocol.ItemInstance
}

func NewStack(br world.BlockRegistry, ist protocol.ItemInstance) Stack {
	return Stack{
		instance: &ist,
		Stack: stackToItem(br, ist.Stack),
	}
}

func instanceFromStack(br world.BlockRegistry, it Stack) protocol.ItemInstance{
	if it.instance == nil{
		return instanceFromItem(br, it.Stack)
	}
	return *it.instance
}

// noinspection ALL
//
//go:linkname stackToItem github.com/df-mc/dragonfly/server/session.stackToItem
func stackToItem(br world.BlockRegistry, it protocol.ItemStack) item.Stack 

// noinspection ALL
//
//go:linkname instanceFromItem github.com/df-mc/dragonfly/server/session.instanceFromItem
func instanceFromItem(br world.BlockRegistry, it item.Stack) protocol.ItemInstance