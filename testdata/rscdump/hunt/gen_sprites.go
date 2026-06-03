//go:build ignore

// gen_sprites authors the entity/ground-item sprite hunt fixtures as rscdump/1
// L1 JSON. The base scene is flat grass, host at the centre, camera looking
// across +Y (yaw 512 like the door fixture). Entities and ground items are
// placed a couple tiles in front of the host so they project into the frame.
//
//	go run gen_sprites.go
package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/gen0cide/westworld/internal/rscdump"
)

const (
	size  = 16
	baseX = 200
	baseY = 200
	plane = 0
	grass = byte(70)
)

func flatTerrain() *rscdump.Terrain {
	n := size * size
	ground := make([]byte, n)
	for i := range ground {
		ground[i] = grass
	}
	return &rscdump.Terrain{
		Size:         size,
		Elevation:    make([]byte, n),
		GroundColour: ground,
		Overlay:      make([]byte, n),
		Roof:         make([]byte, n),
		WallH:        make([]byte, n),
		WallV:        make([]byte, n),
		WallDiag:     make([]int32, n),
		TerrainSeed:  0,
	}
}

func baseCamera() rscdump.Camera {
	return rscdump.Camera{
		X: 0, Y: -384, Z: 0,
		Pitch:    912,
		Yaw:      512,
		Roll:     0,
		Distance: 1500,
		ViewDist: 9,
		ClipNear: 5,
		ClipFar:  7000,
		ScreenW:  512,
		ScreenH:  334,
	}
}

func save(name string, d *rscdump.Dump) {
	out, err := filepath.Abs(filepath.Join("testdata", "rscdump", "hunt", name))
	if err != nil {
		log.Fatal(err)
	}
	if err := d.Save(out); err != nil {
		log.Fatalf("%s: %v", name, err)
	}
	fmt.Printf("wrote %s\n", out)
}

func main() {
	hostX, hostY := baseX+size/2, baseY+size/2 // (208,208)

	// ---- ground item fixtures ----
	// One coin pile (item id 10 = "Coins" in classic config85) a couple tiles in
	// front of the host. We vary the item id across fixtures to test scaling.
	mkItem := func(name string, items []rscdump.GroundItem) {
		d := &rscdump.Dump{
			Schema: rscdump.SchemaID, Level: rscdump.LevelL1, Source: rscdump.SourceHandAuthored,
			Camera:      baseCamera(),
			Window:      rscdump.Window{BaseX: baseX, BaseY: baseY, Plane: plane, Size: size},
			Terrain:     flatTerrain(),
			GroundItems: items,
			Self:        &rscdump.Self{X: hostX, Y: hostY, NoSelf: true},
		}
		save(name, d)
	}

	// single item, item id 10 (coins). 3 tiles in front (+Y).
	mkItem("sprites_item_one.json", []rscdump.GroundItem{
		{X: hostX, Y: hostY + 3, ItemID: 10},
	})
	// twin: no item (control for "is anything drawn?")
	mkItem("sprites_item_none.json", nil)
	// a tall item (item id 77 = a longsword-ish) to test aspect/size vs a coin
	mkItem("sprites_item_tall.json", []rscdump.GroundItem{
		{X: hostX, Y: hostY + 3, ItemID: 77},
	})
	// THREE different items STACKED on the SAME tile (the deob stacks visually by
	// drawing each on top; tests how many are drawn + their lift).
	mkItem("sprites_item_stack3.json", []rscdump.GroundItem{
		{X: hostX, Y: hostY + 3, ItemID: 10},
		{X: hostX, Y: hostY + 3, ItemID: 20},
		{X: hostX, Y: hostY + 3, ItemID: 30},
	})
	// a row of the SAME item at increasing distance — to read the depth scaling.
	mkItem("sprites_item_row.json", []rscdump.GroundItem{
		{X: hostX, Y: hostY + 2, ItemID: 10},
		{X: hostX, Y: hostY + 4, ItemID: 10},
		{X: hostX, Y: hostY + 6, ItemID: 10},
	})

	// ---- entity (npc/player) fixtures ----
	mkEnt := func(name string, ents []rscdump.Entity, self *rscdump.Self) {
		if self == nil {
			self = &rscdump.Self{X: hostX, Y: hostY, NoSelf: true}
		}
		d := &rscdump.Dump{
			Schema: rscdump.SchemaID, Level: rscdump.LevelL1, Source: rscdump.SourceHandAuthored,
			Camera:   baseCamera(),
			Window:   rscdump.Window{BaseX: baseX, BaseY: baseY, Plane: plane, Size: size},
			Terrain:  flatTerrain(),
			Entities: ents,
			Self:     self,
		}
		save(name, d)
	}

	// one npc (id 1) 3 tiles in front
	mkEnt("sprites_npc_one.json", []rscdump.Entity{
		{X: hostX, Y: hostY + 3, Kind: "npc", ID: 1, Heading: 0},
	}, nil)
	// twin: no npc
	mkEnt("sprites_npc_none.json", nil, nil)
	// npc facing test: same npc, headings 0 and 4 (N vs S) at same tile
	mkEnt("sprites_npc_face0.json", []rscdump.Entity{
		{X: hostX, Y: hostY + 3, Kind: "npc", ID: 1, Heading: 0},
	}, nil)
	mkEnt("sprites_npc_face4.json", []rscdump.Entity{
		{X: hostX, Y: hostY + 3, Kind: "npc", ID: 1, Heading: 4},
	}, nil)
	// a default-human player drawn at his own tile (self), no equip
	mkEnt("sprites_self_human.json", nil, &rscdump.Self{X: hostX, Y: hostY, Heading: 0, NoSelf: false})
	// row of npcs at increasing depth, to read scale-by-distance
	mkEnt("sprites_npc_row.json", []rscdump.Entity{
		{X: hostX, Y: hostY + 2, Kind: "npc", ID: 1},
		{X: hostX, Y: hostY + 4, Kind: "npc", ID: 1},
		{X: hostX, Y: hostY + 6, Kind: "npc", ID: 1},
	}, nil)
}
