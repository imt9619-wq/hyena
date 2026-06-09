package blockmap

import (
	"testing"

	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

func TestNewPaletteMapsBlockByName(t *testing.T) {
	items := []protocol.ItemEntry{
		{Name: "minecraft:air", RuntimeID: 0},
		{Name: "minecraft:stone", RuntimeID: 5},
	}
	expected, _ := world.BlockByName("minecraft:stone", nil)
	p := newPalette(items)
	got, ok := p.model(5)
	if !ok || got != expected.Model() {
		t.Fatal("expected stone model from item name")
	}
}

func TestNewPaletteDoesNotUseDragonflyRuntimeIndex(t *testing.T) {
	rid := int16(1)
	wrongBlock, _ := world.BlockByRuntimeID(uint32(rid))
	wrongName, _ := wrongBlock.EncodeBlock()

	items := []protocol.ItemEntry{
		{Name: "minecraft:stone", RuntimeID: rid},
	}
	expected, _ := world.BlockByName("minecraft:stone", nil)
	p := newPalette(items)
	got, ok := p.model(rid)
	if !ok || got != expected.Model() {
		t.Fatalf("should map by item name; dragonfly[%d] is %q", rid, wrongName)
	}
}

func TestNewPaletteDisambiguatesVariants(t *testing.T) {
	items := []protocol.ItemEntry{
		{Name: "minecraft:oak_stairs", RuntimeID: 10},
		{Name: "minecraft:stick", RuntimeID: 11},
		{Name: "minecraft:oak_stairs", RuntimeID: 12},
	}
	states := dragonflyStatesByName()["minecraft:oak_stairs"]
	if len(states) < 2 {
		t.Fatal("need multiple oak_stairs states in dragonfly")
	}

	p := newPalette(items)
	first, ok := world.BlockByName("minecraft:oak_stairs", states[0])
	if !ok {
		t.Fatal("first stairs state should resolve")
	}
	second, ok := world.BlockByName("minecraft:oak_stairs", states[1])
	if !ok {
		t.Fatal("second stairs state should resolve")
	}

	got0, ok := p.model(10)
	if !ok || got0 != first.Model() {
		t.Fatal("first oak_stairs entry should use first variant")
	}
	got1, ok := p.model(12)
	if !ok || got1 != second.Model() {
		t.Fatal("second oak_stairs entry should use second variant")
	}
}

func TestNormalizePaletteName(t *testing.T) {
	if got := normalizePaletteName("minecraft:tallgrass"); got != "minecraft:short_grass" {
		t.Fatalf("got %q", got)
	}
}
