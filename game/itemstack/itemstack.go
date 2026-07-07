package itemstack

import (
	_ "unsafe"

	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
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
    inv, offHand, ui *inventory.Inventory
    armour           *inventory.Armour
	containersOpen   map[uint32]*inventory.Inventory
	packets          *pkbuf.PacketBuffer
}

func NewPlayerItemStack(conn *minecraft.Conn, pks *pkbuf.PacketBuffer) *PlayerItemStack{
	pi := &PlayerItemStack{}
	pi.inv = inventory.New(36, nil)
	pi.offHand = inventory.New(1, nil)
	pi.armour = inventory.NewArmour(nil)
	pi.ui = inventory.New(54, nil)
	pi.entityRuntimeID = conn.GameData().EntityRuntimeID
	pi.packets = pks
	return pi
}

func (pi *PlayerItemStack) HeldItem() (mainhand, offhand item.Stack){
	mainhand, _ = pi.inv.Item(pi.heldSlot)
	offhand, _ = pi.offHand.Item(0) 
	return 
}

func (pi *PlayerItemStack) HeldSlot() int{
	return pi.heldSlot
}

func (pi *PlayerItemStack) SyncInventoryContent(pk *packet.InventoryContent){
	switch pk.WindowID{
	case protocol.WindowIDInventory:
		pi.decodeItemInstanceToInv(pk.Content, pi.inv)
	case protocol.WindowIDArmour:
		pi.decodeItemInstanceToInv(pk.Content, pi.armour.Inventory())
	case protocol.WindowIDUI:
		pi.decodeItemInstanceToInv(pk.Content, pi.ui)
	case protocol.WindowIDOffHand:
		pi.decodeItemInstanceToInv(pk.Content, pi.offHand)
	}
}

func (pi *PlayerItemStack) Equip(pk *packet.MobEquipment){
	switch pk.WindowID{
	case protocol.WindowIDInventory:
		pi.inv.SetItem(int(pk.InventorySlot), stackToItem(world.DefaultBlockRegistry, pk.NewItem.Stack))
	case protocol.WindowIDOffHand:
		pi.offHand.SetItem(int(pk.InventorySlot), stackToItem(world.DefaultBlockRegistry, pk.NewItem.Stack))
	}
}

func (pi *PlayerItemStack) SetHoldSlot(slot int){
	if !(8 >= slot || 0 <= slot) || slot == pi.heldSlot{
		return
	}
	pi.heldSlot = slot
	mainhand, _ := pi.HeldItem()
	pi.packets.Append(&packet.MobEquipment{
		EntityRuntimeID: pi.entityRuntimeID,
		InventorySlot: byte(pi.inv.Size()-1),
		HotBarSlot: byte(pi.inv.Size()-1),
		NewItem: InstanceFromItem(world.DefaultBlockRegistry, mainhand),
	})
}

func (pi *PlayerItemStack) decodeItemInstanceToInv(ct []protocol.ItemInstance, inv *inventory.Inventory){
	if len(ct) != inv.Size(){
		return
	}
	for slot, ist := range ct{
		inv.SetItem(slot, stackToItem(world.DefaultBlockRegistry, ist.Stack))
	}
}

// noinspection ALL
//
//go:linkname stackToItem github.com/df-mc/dragonfly/server/session.stackToItem
func stackToItem(br world.BlockRegistry, it protocol.ItemStack) item.Stack 

// noinspection ALL
//
//go:linkname instanceFromItem github.com/df-mc/dragonfly/server/session.instanceFromItem
func InstanceFromItem(br world.BlockRegistry, it item.Stack) protocol.ItemInstance