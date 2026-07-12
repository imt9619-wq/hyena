// helpers for inventory are copied from github.com/df-mc/dragonfly/server/item/inventory/inventory.go
package itemstack

import (
	"sync"
	
	"github.com/df-mc/dragonfly/server/item/inventory"
)

type Container struct{
	windowId uint32
	inv *Inventory
}

type Inventory struct {
	mu    *sync.RWMutex
	slots []Stack
}

type Armour struct {
    inv *Inventory
}

func NewArmour() *Armour{
	return &Armour{inv: NewInventory(4)}
}

func (a *Armour) Inventory() *Inventory{
	return a.inv
}

func NewInventory(size int) *Inventory{
	if size <= 0 {
		panic("inventory size must be at least 1")
	}
	return &Inventory{mu: &sync.RWMutex{}, slots: make([]Stack, size)}
}

func (inv *Inventory) Item(slot int) (Stack, error) {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	inv.check()
	if !inv.validSlot(slot) {
		return Stack{}, inventory.ErrSlotOutOfRange
	}
	return inv.slots[slot], nil
}

func (inv *Inventory) Size() int {
	inv.mu.RLock()
	defer inv.mu.RUnlock()
	return inv.size()
}

func (inv *Inventory) validSlot(slot int) bool {
	return slot >= 0 && slot < inv.size()
}

func (inv *Inventory) size() int {
	return len(inv.slots)
}

func (inv *Inventory) SetItem(slot int, it Stack) error{
	inv.mu.Lock()

	inv.check()
	if !inv.validSlot(slot) {
		inv.mu.Unlock()
		return inventory.ErrSlotOutOfRange
	}

	inv.setitem(slot, it)
	inv.mu.Unlock()
	return nil
}

func (inv *Inventory) setitem(slot int, it Stack){
	 if it.Count() > it.MaxCount() {
		it.Stack = it.Grow(it.MaxCount() - it.Count())
	}
	inv.slots[slot] = it
}

func (inv *Inventory) check() {
	if inv.size() == 0 {
		panic("uninitialised inventory: inventory must be constructed using inventory.New()")
	}
}
