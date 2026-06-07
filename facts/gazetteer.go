package facts

import (
	_ "embed"
	"encoding/json"
	"math"
	"strings"
)

// World-map gazetteer: named places + typed points-of-interest, so a host
// can perceive WHERE IT IS in the world ("near Lumbridge, the bank is NE")
// rather than just raw tile coords. The data is vendored from the RSC+
// client's well-curated world map (~/Code/rscplus/assets/map) and converted
// from its map-image space into our world tile coords.

//go:embed mapdata/labels.json
var labelsJSON []byte

//go:embed mapdata/points.json
var pointsJSON []byte

// mapImageWidth is plane-0.png's width in the RSC+ world map. The map is
// drawn at 3 px/tile and X-MIRRORED (west on the right); mapToWorld inverts
// WorldMapWindow.convertWorldCoordsToMapRaw. Surface plane only (the data
// is all plane 0). Validated: Lumbridge label (2042,1964) -> world (135,654).
const mapImageWidth = 2448

func mapToWorld(mx, my int) (int, int) {
	return (mapImageWidth-mx-4)/3 + 1, my / 3
}

// Place is a named map location (town/landmark) in WORLD tile coords.
type Place struct {
	Name string `json:"name"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
}

// POI is a typed point of interest (bank, altar, furnace, fishing-point, …)
// in world tile coords.
type POI struct {
	Type string `json:"type"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
}

// Gazetteer holds the converted named places + POIs and answers
// "where am I / what's nearby / where is X" queries.
type Gazetteer struct {
	Places []Place
	POIs   []POI
}

// loadGazetteer parses the embedded RSC+ map data into world coords.
func loadGazetteer() *Gazetteer {
	g := &Gazetteer{}
	var labels []struct {
		Text string `json:"text"`
		X    int    `json:"x"`
		Y    int    `json:"y"`
	}
	if err := json.Unmarshal(labelsJSON, &labels); err == nil {
		for _, l := range labels {
			if strings.TrimSpace(l.Text) == "" {
				continue
			}
			wx, wy := mapToWorld(l.X, l.Y)
			g.Places = append(g.Places, Place{Name: strings.TrimSpace(l.Text), X: wx, Y: wy})
		}
	}
	var pts []struct {
		Type string `json:"type"`
		X    int    `json:"x"`
		Y    int    `json:"y"`
	}
	if err := json.Unmarshal(pointsJSON, &pts); err == nil {
		for _, p := range pts {
			wx, wy := mapToWorld(p.X, p.Y)
			g.POIs = append(g.POIs, POI{Type: p.Type, X: wx, Y: wy})
		}
	}
	return g
}

func gazCheb(ax, ay, bx, by int) int {
	dx := ax - bx
	if dx < 0 {
		dx = -dx
	}
	dy := ay - by
	if dy < 0 {
		dy = -dy
	}
	if dx > dy {
		return dx
	}
	return dy
}

// NearestPlace returns the named place closest to (x,y), its Chebyshev
// distance, and whether any place exists.
func (g *Gazetteer) NearestPlace(x, y int) (Place, int, bool) {
	best := -1
	bestD := 1 << 30
	for i, p := range g.Places {
		if d := gazCheb(x, y, p.X, p.Y); d < bestD {
			best, bestD = i, d
		}
	}
	if best < 0 {
		return Place{}, 0, false
	}
	return g.Places[best], bestD, true
}

// PlaceByName resolves a place by name (exact, case-insensitive; falls back
// to substring), returning the closest match to (x,y) on ties.
func (g *Gazetteer) PlaceByName(name string, x, y int) (Place, bool) {
	want := strings.ToLower(strings.TrimSpace(name))
	if want == "" {
		return Place{}, false
	}
	best := -1
	bestExact := false
	bestD := 1 << 30
	for i, p := range g.Places {
		n := strings.ToLower(p.Name)
		exact := n == want
		if !exact && !strings.Contains(n, want) {
			continue
		}
		d := gazCheb(x, y, p.X, p.Y)
		if best < 0 || (exact && !bestExact) || (exact == bestExact && d < bestD) {
			best, bestExact, bestD = i, exact, d
		}
	}
	if best < 0 {
		return Place{}, false
	}
	return g.Places[best], true
}

// NearestPOI returns the closest POI of the given type to (x,y).
func (g *Gazetteer) NearestPOI(typ string, x, y int) (POI, int, bool) {
	typ = strings.ToLower(strings.TrimSpace(typ))
	best := -1
	bestD := 1 << 30
	for i, p := range g.POIs {
		if typ != "" && !strings.Contains(strings.ToLower(p.Type), typ) {
			continue
		}
		if d := gazCheb(x, y, p.X, p.Y); d < bestD {
			best, bestD = i, d
		}
	}
	if best < 0 {
		return POI{}, 0, false
	}
	return g.POIs[best], bestD, true
}

// POIsWithin returns every POI within Chebyshev `radius` of (x,y).
func (g *Gazetteer) POIsWithin(x, y, radius int) []POI {
	var out []POI
	for _, p := range g.POIs {
		if gazCheb(x, y, p.X, p.Y) <= radius {
			out = append(out, p)
		}
	}
	return out
}

// Bearing returns an 8-point compass direction from (fx,fy) to (tx,ty)
// in RSC world space (north = -Y). Empty string if the points coincide.
func Bearing(fx, fy, tx, ty int) string {
	dx := tx - fx
	dy := ty - fy
	if dx == 0 && dy == 0 {
		return ""
	}
	// angle: 0°=E, 90°=N (north is -Y, so negate dy).
	ang := math.Atan2(float64(-dy), float64(dx)) * 180 / math.Pi
	switch {
	case ang >= -22.5 && ang < 22.5:
		return "E"
	case ang >= 22.5 && ang < 67.5:
		return "NE"
	case ang >= 67.5 && ang < 112.5:
		return "N"
	case ang >= 112.5 && ang < 157.5:
		return "NW"
	case ang >= -67.5 && ang < -22.5:
		return "SE"
	case ang >= -112.5 && ang < -67.5:
		return "S"
	case ang >= -157.5 && ang < -112.5:
		return "SW"
	default:
		return "W"
	}
}
