package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/pathfind"
	"github.com/gen0cide/westworld/render"
	orsc "github.com/gen0cide/westworld/render/orsc"
	"github.com/gen0cide/westworld/runtime"
	"github.com/gen0cide/westworld/world"
)

// spectatorShotDir is where /shot + /clip write captures for the agent to read.
const spectatorShotDir = "/tmp/render_out/spectator"

// --- motion interpolation (smooth walking) ---------------------------------
//
// The server moves actors once per ~600ms walk tick, snapping them a whole tile;
// the browser polls /frame at ~15fps, so without interpolation every actor
// freezes then teleports (the "1s tick" feel). motionTracker remembers each
// actor's previous tile + when it changed and reports a sub-tile WORLD-unit
// offset that GLIDES it from the previous tile to the current over walkGlideDur.
// bernard's glide also scrolls the whole world (he stays centred). This is a
// pure render-side presentation layer keyed on the server's stable actor index;
// no world state is touched, so combat/trade/etc. are unaffected.

// walkGlideDur is one tile-step's duration — locked to RSC's ~600ms walk tick so
// a continuously-walking actor moves at constant speed with no per-step stutter.
const walkGlideDur = 600 * time.Millisecond

// npcWalkModel is the authentic 4-phase leg-cycle frame order (deob drawNpc:
// npcWalkModel = {0,1,2,1}); phase p selects walk frame npcWalkModel[p].
var npcWalkModel = [4]int{0, 1, 2, 1}

type motionTrack struct {
	fromX, fromY int
	toX, toY     int
	since        time.Time
}

type motionTracker struct {
	mu sync.Mutex
	m  map[string]*motionTrack
}

func newMotionTracker() *motionTracker { return &motionTracker{m: map[string]*motionTrack{}} }

// offset records key's current tile (x,y) and returns the sub-tile WORLD-unit
// render offset (offX along world X, offZ along world Z) gliding from the prior
// tile to the current over walkGlideDur, plus the leg-cycle phase + moving flag.
// A non-adjacent jump (teleport / actor re-appearing in view) snaps with no
// glide so it doesn't streak across the map.
func (mt *motionTracker) offset(key string, x, y int, now time.Time) (offX, offZ, phase int, moving bool) {
	mt.mu.Lock()
	defer mt.mu.Unlock()
	t := mt.m[key]
	if t == nil {
		mt.m[key] = &motionTrack{fromX: x, fromY: y, toX: x, toY: y, since: now}
		return 0, 0, 0, false
	}
	if x != t.toX || y != t.toY {
		if dx, dy := x-t.toX, y-t.toY; dx >= -1 && dx <= 1 && dy >= -1 && dy <= 1 {
			t.fromX, t.fromY = t.toX, t.toY // adjacent step: glide from the old tile
		} else {
			t.fromX, t.fromY = x, y // teleport / re-appear: snap, no glide
		}
		t.toX, t.toY = x, y
		t.since = now
	}
	if t.fromX == t.toX && t.fromY == t.toY {
		return 0, 0, 0, false
	}
	frac := float64(now.Sub(t.since)) / float64(walkGlideDur)
	if frac >= 1 {
		return 0, 0, 0, false
	}
	inv := 1 - frac
	offX = int(float64(t.fromX-t.toX) * inv * 128)
	offZ = int(float64(t.fromY-t.toY) * inv * 128)
	phase = npcWalkModel[int(frac*4)&3]
	return offX, offZ, phase, true
}

// apply sets per-actor glide offsets on a freshly-built View in place. selfX/selfY
// are the host's CURRENT tile in the same (plane-local) coords as v.Y.
func (mt *motionTracker) apply(v *render.View, selfX, selfY int, now time.Time) {
	for i := range v.Entities {
		e := &v.Entities[i]
		key := "p"
		if e.Kind == render.EntityNPC {
			key = "n"
		}
		key += strconv.Itoa(e.Index)
		e.OffX, e.OffZ, e.StepPhase, e.Moving = mt.offset(key, e.X, e.Y, now)
	}
	v.SelfOffX, v.SelfOffZ, v.SelfStepPhase, v.SelfMoving = mt.offset("self", selfX, selfY, now)
}

// montage tiles PNG frames into one contact-sheet PNG (each frame downscaled
// 2x, laid out in `cols` columns) so a whole walk burst is one readable image.
func montage(frames [][]byte, cols int) ([]byte, error) {
	var imgs []*image.RGBA
	for _, b := range frames {
		im, err := png.Decode(bytes.NewReader(b))
		if err != nil {
			continue
		}
		bnd := im.Bounds()
		w, h := bnd.Dx()/2, bnd.Dy()/2
		half := image.NewRGBA(image.Rect(0, 0, w, h))
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				half.Set(x, y, im.At(bnd.Min.X+x*2, bnd.Min.Y+y*2))
			}
		}
		imgs = append(imgs, half)
	}
	if len(imgs) == 0 {
		return nil, fmt.Errorf("no decodable frames")
	}
	fw, fh := imgs[0].Bounds().Dx(), imgs[0].Bounds().Dy()
	rows := (len(imgs) + cols - 1) / cols
	sheet := image.NewRGBA(image.Rect(0, 0, fw*cols, fh*rows))
	for i, im := range imgs {
		cx, cy := (i%cols)*fw, (i/cols)*fh
		draw.Draw(sheet, image.Rect(cx, cy, cx+fw, cy+fh), im, image.Point{}, draw.Src)
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, sheet); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// waitForLivePosition polls until the host reports a real (non-spawn) position
// that has held steady for a beat — the world-state/region has loaded and the
// host isn't mid-teleport. Returns (pos, false) if it never loads within the
// timeout. Shared by -render-view and -spectate.
func waitForLivePosition(host *runtime.Host) (world.Coord, bool) {
	var pos, last world.Coord
	deadline := time.Now().Add(15 * time.Second)
	stable := 0
	for time.Now().Before(deadline) {
		pos = host.World().Self.Position()
		if pos.X > 0 && pos.Y > 0 {
			if pos == last {
				if stable++; stable >= 3 {
					return pos, true
				}
			} else {
				stable = 0
			}
			last = pos
		}
		time.Sleep(300 * time.Millisecond)
	}
	if pos.X == 0 && pos.Y == 0 {
		return pos, false
	}
	return pos, true
}

// renderFrame renders the live View through the faithful OpenRSC three/ port
// (render/orsc) — the renderer we standardized on (multi-story buildings, roofs,
// doorframes, animated + gliding actors). render/orsc.PickTile (below) maps clicks
// through the same camera, so the displayed frame and click->tile stay locked.
func renderFrame(land *pathfind.Landscape, f *facts.Facts, v render.View) ([]byte, error) {
	return orsc.RenderViewCached(land, f, v)
}

// buildLiveView snapshots everything the host currently perceives into a
// render.View — the local tile position, every nearby NPC/player billboard (in
// their real appearance), and the live dynamic-state mirrors (opened doors,
// depleted/regrown scenery, dropped items). It leaves the CAMERA fields
// (Rotation / Zoom / W / H) at their zero value for the caller to set, so the
// same snapshot can be rendered from any angle. Reads are lock-safe: every
// world accessor RLocks and returns a fresh copy, so calling this from an HTTP
// goroutine while the net loop mutates state is safe.
func buildLiveView(host *runtime.Host, pos world.Coord) render.View {
	plane := pos.Plane()
	planeOffset := plane * world.PlaneHeight
	localY := pos.Y - planeOffset

	var ents []render.Entity
	for _, npc := range host.World().Npcs.All() {
		if npc.X <= 0 && npc.Y <= 0 {
			continue
		}
		// Server NPC TypeID IS the OpenRSC npc id — used directly as the
		// facts.NpcDef key that drives the Authentic_Sprites.orsc composite.
		ents = append(ents, render.Entity{
			X: npc.X, Y: npc.Y - planeOffset, Kind: render.EntityNPC,
			NpcID: npc.TypeID, Heading: npc.Heading, Index: npc.Index,
		})
	}
	selfName := host.Username()
	for _, pl := range host.World().Players.All() {
		if pl.X <= 0 && pl.Y <= 0 {
			continue
		}
		// Skip the host's own appearance-mirror record by NAME, not by
		// index: the server keys players by a GLOBAL index where index 0
		// can be a legitimate other player (e.g. the first account to log
		// in after a server restart). Matching on index 0 dropped that
		// player from the rendered viewport. (Same fix as the /state dots
		// loop in remoteclient.go.)
		if pl.Name != "" && selfName != "" && strings.EqualFold(pl.Name, selfName) {
			continue
		}
		ent := render.Entity{X: pl.X, Y: pl.Y - planeOffset, Kind: render.EntityPlayer, Heading: pl.Heading, Index: pl.Index}
		if pl.HasEquip {
			ent.EquipSprites = pl.EquipBySlot
			ent.HasEquip = true
		}
		if pl.HasColours {
			ent.HairColour = pl.HairColour
			ent.TopColour = pl.TopColour
			ent.TrouserColour = pl.TrouserColour
			ent.SkinColour = pl.SkinColour
		}
		ents = append(ents, ent)
	}

	v := render.View{
		X:           pos.X,
		Y:           localY,
		Plane:       plane,
		Entities:    ents,
		SelfHeading: host.World().Self.Heading(),
	}
	if self := host.World().Self; self.HasEquip() {
		v.SelfEquipSprites = self.EquipSprites()
		v.SelfHasEquip = true
		if hair, top, trouser, skin, ok := self.AppearanceColours(); ok {
			v.SelfHairColour = hair
			v.SelfTopColour = top
			v.SelfTrouserColour = trouser
			v.SelfSkinColour = skin
		}
	}
	if w := host.World(); w != nil {
		v.BoundaryRemoved = func(x, ay, dir int) bool { return w.Boundaries.IsRemoved(x, ay, dir) }
		v.SceneryRemoved = func(x, ay int) bool { return w.Scenery.IsRemoved(x, ay) }
		for _, ds := range w.Scenery.All() {
			if ds.X <= 0 && ds.Y <= 0 {
				continue
			}
			v.DynamicScenery = append(v.DynamicScenery, render.DynamicSceneryItem{X: ds.X, Y: ds.Y - planeOffset, ID: ds.ID})
		}
		for _, gi := range w.GroundItems.All() {
			if gi.X <= 0 && gi.Y <= 0 {
				continue
			}
			v.GroundItems = append(v.GroundItems, render.GroundItemMarker{X: gi.X, Y: gi.Y - planeOffset, ItemID: gi.ItemID})
		}
	}
	return v
}

// spectate serves a LIVE browser viewport that follows the host around the
// world. It reuses the exact decoupled render path (buildLiveView ->
// render.RenderView -> PNG) the static -render-view uses, but re-snapshots +
// re-renders the host's current position on every frame request, so the view
// tracks the host as it walks and the camera angle/zoom are driven live from
// the browser. The browser IS the window — there is no native window and no
// CGo: the server only speaks stdlib net/http and ships PNG frames.
//
// Endpoints:
//
//	GET /            the viewport HTML/JS page
//	GET /frame?rot=&zoom=&w=&h=   one freshly-rendered PNG of the host's view
//	GET /pos         {"x","y","plane"} for the on-screen HUD
//
// It blocks (serving) until ctx is cancelled (Ctrl-C), keeping the host logged
// in the whole time; on cancel it shuts the server down cleanly so the deferred
// graceful-logout still runs.
func spectate(ctx context.Context, log *slog.Logger, cfg config, host *runtime.Host, land *pathfind.Landscape, f *facts.Facts) error {
	if land == nil {
		return fmt.Errorf("spectate: no landscape loaded (need -facts pointing at OpenRSC root)")
	}
	if _, ok := waitForLivePosition(host); !ok {
		return fmt.Errorf("spectate: host position never loaded (still 0,0)")
	}

	atoiOr := func(s string, def int) int {
		if n, err := strconv.Atoi(s); err == nil {
			return n
		}
		return def
	}

	// Serialized click-to-walk worker: a /walk click queues one target tile; the
	// worker runs host.WalkTo (the BFS pathfinder — routes around walls/scenery and
	// sends the waypoint path, instead of host.Walk's single straight walk-to-point)
	// one at a time. A new click cancels the in-flight WalkTo and retargets (RSC walk
	// overrides the previous anyway). ctx-scoped so it stops on shutdown.
	walkCh := make(chan [2]int, 1)
	go func() {
		var cancel context.CancelFunc
		var done chan struct{}
		stop := func() {
			if cancel != nil {
				cancel()
				<-done // wait for the in-flight WalkTo to actually return (no overlap)
				cancel, done = nil, nil
			}
		}
		for {
			select {
			case <-ctx.Done():
				stop()
				return
			case t := <-walkCh:
				stop() // retarget: drop any in-flight journey first
				wctx, c := context.WithTimeout(ctx, 90*time.Second)
				cancel, done = c, make(chan struct{})
				go func(t [2]int, wctx context.Context, done chan struct{}) {
					defer close(done)
					if err := host.WalkTo(wctx, t[0], t[1]); err != nil {
						log.Debug("spectate walk", "to", t, "err", err)
					}
				}(t, wctx, done)
			}
		}
	}()

	mux := http.NewServeMux()
	// motion tracks per-actor tile changes across frames to glide them smoothly
	// (shared by /frame + /shot + /clip; thread-safe).
	motion := newMotionTracker()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, spectatePage, cfg.username, cfg.renderZoom, cfg.renderW, cfg.renderH)
	})
	mux.HandleFunc("/pos", func(w http.ResponseWriter, r *http.Request) {
		p := host.World().Self.Position()
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		fmt.Fprintf(w, `{"x":%d,"y":%d,"plane":%d}`, p.X, p.Y, p.Plane())
	})

	// renderMu serializes every RenderView call. The render pipeline shares
	// mutable state (the model cache map, the landscape sector cache); the
	// browser polls /frame continuously and /shot|/clip can fire concurrently,
	// so without this two concurrent renders crash the process with "concurrent
	// map read and map write". Renders are ~28ms, so serializing is cheap.
	var renderMu sync.Mutex
	mux.HandleFunc("/frame", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		pos := host.World().Self.Position()
		if pos.X <= 0 && pos.Y <= 0 {
			http.Error(w, "host position not loaded", http.StatusServiceUnavailable)
			return
		}
		v := buildLiveView(host, pos)
		motion.apply(&v, v.X, v.Y, time.Now()) // smooth tile-to-tile glide
		v.Rotation = atoiOr(q.Get("rot"), cfg.renderRotation) & 0xff
		v.Zoom = atoiOr(q.Get("zoom"), cfg.renderZoom)
		v.W = atoiOr(q.Get("w"), cfg.renderW)
		v.H = atoiOr(q.Get("h"), cfg.renderH)
		v.AnimFrame = atoiOr(q.Get("anim"), 0) // model-swap frame (fires/torches flicker)
		renderMu.Lock()
		png, err := renderFrame(land, f, v)
		renderMu.Unlock()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write(png)
	})
	// /walk?px=&py=&rot=&zoom=&w=&h= : screen click -> world tile -> host.Walk.
	// PickTile rebuilds the exact camera (rot/zoom/w/h the page rendered with) +
	// the host's CURRENT position, then picks the window tile nearest the click.
	mux.HandleFunc("/walk", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		px, py := atoiOr(q.Get("px"), -1), atoiOr(q.Get("py"), -1)
		if px < 0 || py < 0 {
			http.Error(w, "missing px/py", http.StatusBadRequest)
			return
		}
		pos := host.World().Self.Position()
		if pos.X <= 0 && pos.Y <= 0 {
			http.Error(w, "host position not loaded", http.StatusServiceUnavailable)
			return
		}
		plane := pos.Plane()
		v := render.View{
			X:        pos.X,
			Y:        pos.Y - plane*world.PlaneHeight, // PickTile works in plane-local Y
			Plane:    plane,
			Rotation: atoiOr(q.Get("rot"), cfg.renderRotation) & 0xff,
			Zoom:     atoiOr(q.Get("zoom"), cfg.renderZoom),
			W:        atoiOr(q.Get("w"), cfg.renderW),
			H:        atoiOr(q.Get("h"), cfg.renderH),
		}
		tx, ty, ok := orsc.PickTile(land, v, px, py)
		if !ok {
			http.Error(w, "no tile under click", http.StatusNoContent)
			return
		}
		absY := ty + plane*world.PlaneHeight // PickTile returns plane-local Y
		// retarget: drop any pending, queue the latest (non-blocking).
		select {
		case <-walkCh:
		default:
		}
		select {
		case walkCh <- [2]int{tx, absY}:
		default:
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"x":%d,"y":%d}`, tx, absY)
	})
	// /teleport?x=&y= : ADMIN/TEST teleport via the host's `teleport <x> <y>`
	// command (the same one reset-on-exit uses). Requires an admin account.
	// Defaults to the Lumbridge spawn (120,649) when x/y are omitted.
	mux.HandleFunc("/teleport", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		tx, ty := atoiOr(q.Get("x"), 120), atoiOr(q.Get("y"), 649)
		cctx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
		if err := host.Command(cctx, fmt.Sprintf("teleport %d %d", tx, ty)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Info("spectate /teleport", "x", tx, "y", ty)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"teleport":[%d,%d]}`, tx, ty)
	})
	// renderOne builds the live view with the page's camera params + one PNG.
	renderOne := func(q map[string][]string, animFrame int) ([]byte, error) {
		get := func(k string) string {
			if v := q[k]; len(v) > 0 {
				return v[0]
			}
			return ""
		}
		pos := host.World().Self.Position()
		v := buildLiveView(host, pos)
		motion.apply(&v, v.X, v.Y, time.Now()) // smooth tile-to-tile glide
		v.Rotation = atoiOr(get("rot"), cfg.renderRotation) & 0xff
		v.Zoom = atoiOr(get("zoom"), cfg.renderZoom)
		v.W = atoiOr(get("w"), cfg.renderW)
		v.H = atoiOr(get("h"), cfg.renderH)
		v.AnimFrame = animFrame
		renderMu.Lock()
		defer renderMu.Unlock()
		return renderFrame(land, f, v)
	}
	// /shot : save the CURRENT frame to spectatorShotDir/shot.png (hotkey p).
	mux.HandleFunc("/shot", func(w http.ResponseWriter, r *http.Request) {
		img, err := renderOne(r.URL.Query(), atoiOr(r.URL.Query().Get("anim"), 0))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = os.MkdirAll(spectatorShotDir, 0o755)
		// Write a UNIQUE numbered file (shot_HHMMSS.png) so multiple captures
		// persist for review, PLUS shot.png as the always-latest alias. The old
		// fixed shot.png overwrote every press, losing earlier shots.
		stamp := time.Now().Format("150405")
		numbered := filepath.Join(spectatorShotDir, "shot_"+stamp+".png")
		_ = os.WriteFile(numbered, img, 0o644)
		latest := filepath.Join(spectatorShotDir, "shot.png")
		_ = os.WriteFile(latest, img, 0o644)
		log.Info("spectate /shot saved (for the agent)", "path", numbered)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"saved":%q}`, numbered)
	})
	// /clip : capture a 12-frame burst (~2.6s) of the LIVE view (walk while it
	// runs) + tile into one contact sheet clip_sheet.png (hotkey c). Per-frame
	// PNGs clip_00..11 also saved. AnimFrame advances so model anims cycle too.
	mux.HandleFunc("/clip", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		const nFrames = 12
		const gap = 220 * time.Millisecond
		_ = os.MkdirAll(spectatorShotDir, 0o755)
		var frames [][]byte
		for k := 0; k < nFrames; k++ {
			img, err := renderOne(q, k)
			if err == nil {
				frames = append(frames, img)
				_ = os.WriteFile(filepath.Join(spectatorShotDir, fmt.Sprintf("clip_%02d.png", k)), img, 0o644)
			}
			if k < nFrames-1 {
				time.Sleep(gap)
			}
		}
		sheet, err := montage(frames, 4)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		stamp := time.Now().Format("150405")
		p := filepath.Join(spectatorShotDir, "clip_sheet_"+stamp+".png")
		_ = os.WriteFile(p, sheet, 0o644)
		_ = os.WriteFile(filepath.Join(spectatorShotDir, "clip_sheet.png"), sheet, 0o644) // latest alias
		log.Info("spectate /clip saved (for the agent)", "sheet", p, "frames", len(frames))
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"saved":%q,"frames":%d}`, p, len(frames))
	})

	srv := &http.Server{Addr: cfg.spectateAddr, Handler: mux}
	go func() {
		<-ctx.Done()
		sc, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(sc)
	}()

	log.Info("spectator viewport serving — open this in a browser; Ctrl-C to stop",
		"url", "http://"+cfg.spectateAddr+"/", "host", cfg.username)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("spectate: serve %q: %w", cfg.spectateAddr, err)
	}
	return nil
}

// spectatePage is the viewport HTML. %s=host name, %d=zoom, %d=w, %d=h. The
// rendered frame is fetched as a PNG and CSS-scaled (pixelated) to fill the
// window, so the render resolution stays cheap while the display fills the
// screen with the authentic chunky RSC look. Arrow keys rotate (8-way, matching
// the renderer's deterministic yaw steps), +/- zoom.
const spectatePage = `<!doctype html><html><head><meta charset="utf-8">
<title>cradle spectator — %s</title>
<style>
  html,body{margin:0;height:100%%;background:#000;overflow:hidden;font:13px monospace;color:#9f9}
  #v{position:absolute;inset:0;width:100%%;height:100%%;object-fit:contain;image-rendering:pixelated}
  #hud{position:absolute;left:8px;top:6px;text-shadow:0 0 3px #000;white-space:pre;pointer-events:none}
  #flash{position:absolute;right:10px;top:8px;color:#ff8;text-shadow:0 0 4px #000;opacity:0;transition:opacity .2s;pointer-events:none;font-size:15px}
</style></head><body>
<img id="v">
<div id="hud"></div>
<div id="flash"></div>
<script>
let rot=64, zoom=%d, w=%d, h=%d, anim=0;
// Rotation is CONTINUOUS over all 256 camera angles (not 8-way): hold < or >
// to spin smoothly, RS-style. ROT = yaw units advanced per frame while held.
const ROT=4;
const keys={};
const img=document.getElementById('v'), hud=document.getElementById('hud');
let busy=false, pos={x:0,y:0,plane:0};
async function refreshPos(){ try{ pos=await (await fetch('/pos')).json(); }catch(e){} }
function drawHud(){ hud.textContent =
  'host ('+pos.x+', '+pos.y+')  plane '+pos.plane+'\n'+
  'rot '+rot+'   zoom '+zoom+'   '+w+'x'+h+'\n'+
  'hold < >  rotate    + -  zoom    [ ]  view size    CLICK walk    p shot  c clip'; }
function tick(){
  // advance the camera every tick while a rotate key is held (smooth spin),
  // independent of whether a frame is still in flight.
  if(keys['ArrowLeft'])  rot=(rot-ROT+256)&255;
  if(keys['ArrowRight']) rot=(rot+ROT)&255;
  if(busy) return; busy=true;
  const u='/frame?rot='+rot+'&zoom='+zoom+'&w='+w+'&h='+h+'&anim='+anim+'&t='+performance.now();
  const n=new Image();
  n.onload=()=>{ img.src=n.src; busy=false; drawHud(); };
  n.onerror=()=>{ busy=false; };
  n.src=u;
}
const HANDLED=['ArrowLeft','ArrowRight','+','=','-','_','[',']','p','c'];
function flash(msg){ const o=document.getElementById('flash'); if(!o)return; o.textContent=msg; o.style.opacity=1; setTimeout(()=>o.style.opacity=0,1200); }
document.addEventListener('keydown',e=>{
  if(!HANDLED.includes(e.key)) return;
  e.preventDefault();
  keys[e.key]=true;
  const cam='rot='+rot+'&zoom='+zoom+'&w='+w+'&h='+h;
  // + zooms IN (camera closer = SMALLER zoom/distance); - zooms OUT. (Was inverted.)
  if(e.key==='+'||e.key==='=') zoom=Math.max(zoom-150,250);
  else if(e.key==='-'||e.key==='_') zoom=Math.min(zoom+150,4000);
  else if(e.key==='[') { w=Math.max(w-128,256); h=Math.max(h-84,168); }
  else if(e.key===']') { w=Math.min(w+128,1280); h=Math.min(h+84,840); }
  else if(e.key==='p'){ flash('shot…'); fetch('/shot?'+cam).then(()=>flash('shot saved')); }
  else if(e.key==='c'){ flash('clip… walk now (~3s)'); fetch('/clip?'+cam).then(()=>flash('clip saved')); }
  drawHud();
});
document.addEventListener('keyup',e=>{ keys[e.key]=false; });
// CLICK TO WALK: map the click to a frame pixel (undo object-fit:contain
// letterboxing) and POST it; the server picks the tile under it + walks there.
img.addEventListener('click',e=>{
  const r=img.getBoundingClientRect();
  const scale=Math.min(r.width/w, r.height/h);
  if(scale<=0) return;
  const px=Math.round((e.clientX-(r.left+(r.width-w*scale)/2))/scale);
  const py=Math.round((e.clientY-(r.top+(r.height-h*scale)/2))/scale);
  if(px<0||py<0||px>=w||py>=h) return;
  fetch('/walk?px='+px+'&py='+py+'&rot='+rot+'&zoom='+zoom+'&w='+w+'&h='+h);
});
setInterval(refreshPos,500);
setInterval(()=>{anim=(anim+1)&0x3fffffff;},160); // model-swap animation pace (~6/s); fires/torches flicker
setInterval(tick,33);   // ~30fps target for smooth gliding; tick() no-ops while a frame is in flight
refreshPos(); drawHud();
</script></body></html>`
