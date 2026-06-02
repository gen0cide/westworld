package facts

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
)

// Sources tells the loader where to find OpenRSC's data files. All
// paths are absolute (or relative to PWD); typically callers set
// Root to the OpenRSC server directory's conf/server/defs/ and the
// individual filenames default to OpenRSC's conventions.
//
// Optional secondary files (XxxLocs14.json, XxxLocs27.json, etc.) can
// be added via Extra*Loc lists when the host's location_data config
// includes them.
type Sources struct {
	Root string // e.g., "/Users/flint/Code/openrsc/server/conf/server/defs"

	SceneryDefsXML  string // GameObjectDef.xml
	BoundaryDefsXML string // DoorDef.xml
	NpcDefsJSON     string // NpcDefs.json
	ItemDefsJSON    string // ItemDefs.json

	SceneryLocsJSON  string // locs/SceneryLocs.json
	BoundaryLocsJSON string // locs/BoundaryLocs.json
	NpcLocsJSON      string // locs/NpcLocs.json
	GroundItemsJSON  string // locs/GroundItems.json
}

// DefaultSources returns Sources pointing at the OpenRSC server tree
// at the given root path.
func DefaultSources(openRSCRoot string) Sources {
	defs := filepath.Join(openRSCRoot, "server", "conf", "server", "defs")
	return Sources{
		Root:             defs,
		SceneryDefsXML:   filepath.Join(defs, "GameObjectDef.xml"),
		BoundaryDefsXML:  filepath.Join(defs, "DoorDef.xml"),
		NpcDefsJSON:      filepath.Join(defs, "NpcDefs.json"),
		ItemDefsJSON:     filepath.Join(defs, "ItemDefs.json"),
		SceneryLocsJSON:  filepath.Join(defs, "locs", "SceneryLocs.json"),
		BoundaryLocsJSON: filepath.Join(defs, "locs", "BoundaryLocs.json"),
		NpcLocsJSON:      filepath.Join(defs, "locs", "NpcLocs.json"),
		GroundItemsJSON:  filepath.Join(defs, "locs", "GroundItems.json"),
	}
}

// Load reads all defs and placements into an in-memory Facts store.
// Returns errors for missing required files; missing optional files
// are reported as nil entries.
func Load(s Sources) (*Facts, error) {
	f := &Facts{
		SceneryDefs:  map[int]*SceneryDef{},
		BoundaryDefs: map[int]*BoundaryDef{},
		NpcDefs:      map[int]*NpcDef{},
		ItemDefs:     map[int]*ItemDef{},
	}

	// Defs — loaded first so loc IDs can be resolved.
	if err := loadSceneryDefsXML(s.SceneryDefsXML, f); err != nil {
		return nil, fmt.Errorf("scenery defs: %w", err)
	}
	if err := loadBoundaryDefsXML(s.BoundaryDefsXML, f); err != nil {
		return nil, fmt.Errorf("boundary defs: %w", err)
	}
	if err := loadNpcDefsJSON(s.NpcDefsJSON, f); err != nil {
		return nil, fmt.Errorf("npc defs: %w", err)
	}
	if err := loadItemDefsJSON(s.ItemDefsJSON, f); err != nil {
		return nil, fmt.Errorf("item defs: %w", err)
	}

	// Placements.
	if err := loadSceneryLocsJSON(s.SceneryLocsJSON, f); err != nil {
		return nil, fmt.Errorf("scenery locs: %w", err)
	}
	if err := loadBoundaryLocsJSON(s.BoundaryLocsJSON, f); err != nil {
		return nil, fmt.Errorf("boundary locs: %w", err)
	}
	if err := loadNpcLocsJSON(s.NpcLocsJSON, f); err != nil {
		return nil, fmt.Errorf("npc locs: %w", err)
	}
	if err := loadGroundItemsJSON(s.GroundItemsJSON, f); err != nil {
		return nil, fmt.Errorf("ground items: %w", err)
	}

	// Build spatial indexes.
	f.buildIndex()
	return f, nil
}

// ----- XML structures for unmarshalling -----

type gameObjectDefArrayXML struct {
	Defs []gameObjectDefXML `xml:"GameObjectDef"`
}
type gameObjectDefXML struct {
	Name           string `xml:"name"`
	Description    string `xml:"description"`
	Command1       string `xml:"command1"`
	Command2       string `xml:"command2"`
	Type           int    `xml:"type"`
	Width          int    `xml:"width"`
	Height         int    `xml:"height"`
	GroundItemVar  int    `xml:"groundItemVar"`
	ObjectModel    string `xml:"objectModel"`
}

func loadSceneryDefsXML(path string, f *Facts) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var arr gameObjectDefArrayXML
	if err := xml.Unmarshal(b, &arr); err != nil {
		return err
	}
	for i, d := range arr.Defs {
		f.SceneryDefs[i] = &SceneryDef{
			ID:          i,
			Name:        d.Name,
			Description: d.Description,
			Command1:    d.Command1,
			Command2:    d.Command2,
			Type:        d.Type,
			Width:       d.Width,
			Height:      d.Height,
			Model:       d.ObjectModel,
		}
	}
	return nil
}

type doorDefArrayXML struct {
	Defs []doorDefXML `xml:"DoorDef"`
}
type doorDefXML struct {
	Name        string `xml:"name"`
	Description string `xml:"description"`
	Command1    string `xml:"command1"`
	Command2    string `xml:"command2"`
	ModelVar1   int    `xml:"modelVar1"`
	ModelVar2   int    `xml:"modelVar2"`
	ModelVar3   int    `xml:"modelVar3"`
	DoorType    int    `xml:"doorType"`
	Unknown     int    `xml:"unknown"`
}

func loadBoundaryDefsXML(path string, f *Facts) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var arr doorDefArrayXML
	if err := xml.Unmarshal(b, &arr); err != nil {
		return err
	}
	for i, d := range arr.Defs {
		f.BoundaryDefs[i] = &BoundaryDef{
			ID:          i,
			Name:        d.Name,
			Description: d.Description,
			Command1:    d.Command1,
			Command2:    d.Command2,
			Height:      d.ModelVar1,
			FrontDeco:   d.ModelVar2,
			BackDeco:    d.ModelVar3,
			DoorType:    d.DoorType,
			Unknown:     d.Unknown,
		}
	}
	return nil
}

type npcDefsContainer struct {
	Defs []npcDefJSON `json:"npcs"`
}
type npcDefJSON struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Command     string `json:"command"`
	Command2    string `json:"command2"`
	Hits        int    `json:"hits"`
	Attack      int    `json:"attack"`
	Defense     int    `json:"defense"`
	Strength    int    `json:"strength"`
	Attackable  int    `json:"attackable"`
	Aggressive  int    `json:"aggressive"`
	Sprites1    int    `json:"sprites1"`
	Sprites2    int    `json:"sprites2"`
	Sprites3    int    `json:"sprites3"`
	Sprites4    int    `json:"sprites4"`
	Sprites5    int    `json:"sprites5"`
	Sprites6    int    `json:"sprites6"`
	Sprites7    int    `json:"sprites7"`
	Sprites8    int    `json:"sprites8"`
	Sprites9    int    `json:"sprites9"`
	Sprites10   int    `json:"sprites10"`
	Sprites11   int    `json:"sprites11"`
	Sprites12   int    `json:"sprites12"`
	HairColour   int   `json:"hairColour"`
	TopColour    int   `json:"topColour"`
	BottomColour int   `json:"bottomColour"`
	SkinColour   int   `json:"skinColour"`
	Camera1      int   `json:"camera1"`
	Camera2      int   `json:"camera2"`
}

// toNpcDef builds the in-memory NpcDef (rendering fields included) from the JSON
// record. Sprite slots arrive as 12 individual fields; assemble them into the
// [12]int the renderer indexes.
func (d npcDefJSON) toNpcDef() *NpcDef {
	return &NpcDef{
		ID: d.ID, Name: d.Name, Description: d.Description,
		Command1: d.Command, Command2: d.Command2,
		Hits: d.Hits, Attack: d.Attack, Defense: d.Defense, Strength: d.Strength,
		Attackable: d.Attackable != 0, Aggressive: d.Aggressive != 0,
		Sprites: [12]int{
			d.Sprites1, d.Sprites2, d.Sprites3, d.Sprites4, d.Sprites5, d.Sprites6,
			d.Sprites7, d.Sprites8, d.Sprites9, d.Sprites10, d.Sprites11, d.Sprites12,
		},
		HairColour: d.HairColour, TopColour: d.TopColour, BottomColour: d.BottomColour,
		SkinColour: d.SkinColour, Camera1: d.Camera1, Camera2: d.Camera2,
	}
}

func loadNpcDefsJSON(path string, f *Facts) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	// File may have outer key like "npc" wrapping the array. Try the
	// container first; fall back to bare array.
	var c npcDefsContainer
	if err := json.Unmarshal(b, &c); err == nil && len(c.Defs) > 0 {
		for _, d := range c.Defs {
			f.NpcDefs[d.ID] = d.toNpcDef()
		}
		return nil
	}
	var arr []npcDefJSON
	if err := json.Unmarshal(b, &arr); err != nil {
		return err
	}
	for _, d := range arr {
		f.NpcDefs[d.ID] = d.toNpcDef()
	}
	return nil
}

type itemDefsContainer struct {
	Items []itemDefJSON `json:"item"`
}
type itemDefJSON struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Command       string `json:"command"`
	IsMembersOnly int    `json:"isMembersOnly"`
	IsStackable   int    `json:"isStackable"`
	IsUntradable  int    `json:"isUntradable"`
	IsWearable    int    `json:"isWearable"`
	BasePrice     int    `json:"basePrice"`
	AppearanceID  int    `json:"appearanceID"`
}

func loadItemDefsJSON(path string, f *Facts) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var c itemDefsContainer
	if err := json.Unmarshal(b, &c); err == nil && len(c.Items) > 0 {
		for _, d := range c.Items {
			f.ItemDefs[d.ID] = &ItemDef{
				ID: d.ID, Name: d.Name, Description: d.Description, Command: d.Command,
				IsMembersOnly: d.IsMembersOnly != 0, IsStackable: d.IsStackable != 0,
				IsUntradable: d.IsUntradable != 0, IsWearable: d.IsWearable != 0,
				BasePrice: d.BasePrice, AppearanceID: d.AppearanceID,
			}
		}
		return nil
	}
	var arr []itemDefJSON
	if err := json.Unmarshal(b, &arr); err != nil {
		return err
	}
	for _, d := range arr {
		f.ItemDefs[d.ID] = &ItemDef{
			ID: d.ID, Name: d.Name, Description: d.Description, Command: d.Command,
			IsMembersOnly: d.IsMembersOnly != 0, IsStackable: d.IsStackable != 0,
			IsUntradable: d.IsUntradable != 0, IsWearable: d.IsWearable != 0,
			BasePrice: d.BasePrice,
		}
	}
	return nil
}

// ----- Placement JSON structures -----

type sceneryLocsJSON struct {
	Sceneries []struct {
		ID        int     `json:"id"`
		Pos       posJSON `json:"pos"`
		Direction int     `json:"direction"`
	} `json:"sceneries"`
}

type boundaryLocsJSON struct {
	Boundaries []struct {
		ID        int     `json:"id"`
		Pos       posJSON `json:"pos"`
		Direction int     `json:"direction"`
	} `json:"boundaries"`
}

type npcLocsJSON struct {
	NpcLocs []struct {
		ID    int     `json:"id"`
		Start posJSON `json:"start"`
		Min   posJSON `json:"min"`
		Max   posJSON `json:"max"`
	} `json:"npclocs"`
}

type groundItemsJSON struct {
	GroundItems []struct {
		ID        int     `json:"id"`
		Pos       posJSON `json:"pos"`
		Respawn   int     `json:"respawnTime"`
	} `json:"grounditems"`
}

type posJSON struct {
	X int `json:"X"`
	Y int `json:"Y"`
}

func loadSceneryLocsJSON(path string, f *Facts) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var data sceneryLocsJSON
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	f.SceneryLocs = make([]SceneryLoc, len(data.Sceneries))
	for i, s := range data.Sceneries {
		f.SceneryLocs[i] = SceneryLoc{DefID: s.ID, X: s.Pos.X, Y: s.Pos.Y, Direction: s.Direction}
	}
	return nil
}

func loadBoundaryLocsJSON(path string, f *Facts) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var data boundaryLocsJSON
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	f.BoundaryLocs = make([]BoundaryLoc, len(data.Boundaries))
	for i, s := range data.Boundaries {
		f.BoundaryLocs[i] = BoundaryLoc{DefID: s.ID, X: s.Pos.X, Y: s.Pos.Y, Direction: s.Direction}
	}
	return nil
}

func loadNpcLocsJSON(path string, f *Facts) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var data npcLocsJSON
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	f.NpcLocs = make([]NpcLoc, len(data.NpcLocs))
	for i, s := range data.NpcLocs {
		f.NpcLocs[i] = NpcLoc{
			DefID: s.ID,
			StartX: s.Start.X, StartY: s.Start.Y,
			MinX: s.Min.X, MinY: s.Min.Y,
			MaxX: s.Max.X, MaxY: s.Max.Y,
		}
	}
	return nil
}

func loadGroundItemsJSON(path string, f *Facts) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var data groundItemsJSON
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	f.GroundItemLocs = make([]GroundItemLoc, len(data.GroundItems))
	for i, s := range data.GroundItems {
		f.GroundItemLocs[i] = GroundItemLoc{DefID: s.ID, X: s.Pos.X, Y: s.Pos.Y, RespawnMs: s.Respawn}
	}
	return nil
}
