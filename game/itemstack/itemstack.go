package itemstack

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/imt9619-wq/hyena/utils/pkbuf"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func init() {
	world.DefaultBlockRegistry.Finalize()
}

type PlayerItemStack struct {   
	entityRuntimeID  uint64
	heldSlot         int
    inv, offHand, ui *Inventory
    armour           *Armour
	containersOpen   map[uint32]*Inventory
	packets          *pkbuf.PacketBuffer
}

func NewPlayerItemStack(conn *minecraft.Conn, pks *pkbuf.PacketBuffer) *PlayerItemStack{
	pi := &PlayerItemStack{}
	pi.inv = NewInventory(36)
	pi.offHand = NewInventory(1)
	pi.armour = NewArmour()
	pi.ui = NewInventory(54)
	pi.containersOpen = make(map[uint32]*Inventory, 10)
	pi.entityRuntimeID = conn.GameData().EntityRuntimeID
	pi.packets = pks
	return pi
}

func (pi *PlayerItemStack) HeldItem() (mainhand, offhand Stack){
	mainhand, _ = pi.inv.Item(pi.heldSlot)
	offhand, _ = pi.offHand.Item(0) 
	return 
}

func (pi *PlayerItemStack) HeldSlot() int{
	return pi.heldSlot
}

func (pi *PlayerItemStack) SyncInventoryContent(pk *packet.InventoryContent){
	inv, ok := pi.inventoryByWindowId(pk.WindowID)
	if !ok{
		inv = NewInventory(len(pk.Content))
		pi.containersOpen[pk.WindowID] = inv
	}
	pi.decodeItemInstanceToInv(pk.Content, inv)
}

func (pi *PlayerItemStack) inventoryByWindowId(windowId uint32) (*Inventory, bool){
	switch windowId{
	case protocol.WindowIDInventory:
		return pi.inv, true
	case protocol.WindowIDArmour:
		return pi.armour.Inventory(), true
	case protocol.WindowIDUI:
		return pi.ui, true
	case protocol.WindowIDOffHand:
		return pi.offHand, true
	default:
		inv, ok := pi.containersOpen[windowId]
		return inv, ok
	}
}

func (pi *PlayerItemStack) SetItemOnInvSlot(windowId uint32, slot uint32, ist protocol.ItemInstance){
	inv, ok := pi.inventoryByWindowId(windowId)
	if !ok{
		return
	}
	if slot > uint32(inv.Size()-1){
		return
	}
	inv.SetItem(int(slot), NewStack(world.DefaultBlockRegistry, ist))
}

func (pi *PlayerItemStack) Equip(pk *packet.MobEquipment){
	switch pk.WindowID{
	case protocol.WindowIDInventory:
		pi.inv.SetItem(int(pk.InventorySlot), NewStack(world.DefaultBlockRegistry, pk.NewItem))
	case protocol.WindowIDOffHand:
		pi.offHand.SetItem(int(pk.InventorySlot), NewStack(world.DefaultBlockRegistry, pk.NewItem))
	}
}

func (pi *PlayerItemStack) SetHoldSlot(slot int){
	if !(8 >= slot && 0 <= slot) || slot == pi.heldSlot{
		return
	}
	pi.heldSlot = slot
	pi.packets.Append(&packet.MobEquipment{
		EntityRuntimeID: pi.entityRuntimeID,
		InventorySlot: byte(slot),
		HotBarSlot: byte(slot),
		NewItem: pi.SlotInstance(pi.heldSlot),
	})
}

func (pi *PlayerItemStack) decodeItemInstanceToInv(ct []protocol.ItemInstance, inv *Inventory){
	if len(ct) != inv.Size(){
		return
	}
	for slot, ist := range ct{
		inv.SetItem(slot, NewStack(world.DefaultBlockRegistry, ist))
	}
}

func (pi *PlayerItemStack) SlotInstance(slot int) protocol.ItemInstance{
	mainhand, _ := pi.inv.Item(slot)
	return instanceFromStack(world.DefaultBlockRegistry, mainhand)
}

func InstanceFromItem(br world.BlockRegistry, it item.Stack) protocol.ItemInstance{
	return instanceFromItem(br, it)
}
