package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/pathfind"
	"github.com/gen0cide/westworld/render"
	"github.com/gen0cide/westworld/runtime"
	"github.com/gen0cide/westworld/world"
)

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
		// Server NPC TypeID IS the config85 sprite-id space directly.
		ents = append(ents, render.Entity{
			X: npc.X, Y: npc.Y - planeOffset, Kind: render.EntityNPC,
			NpcID: npc.TypeID, Heading: npc.Heading,
		})
	}
	for _, pl := range host.World().Players.All() {
		if pl.Index == 0 || (pl.X <= 0 && pl.Y <= 0) {
			continue // index 0 is self; the camera sits on it
		}
		ent := render.Entity{X: pl.X, Y: pl.Y - planeOffset, Kind: render.EntityPlayer, Heading: pl.Heading}
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
	modelsPath := filepath.Join(cfg.factsRoot, "Client_Base", "Cache", "video", "models.orsc")
	bundle, err := render.OpenBundle(modelsPath)
	if err != nil {
		return fmt.Errorf("spectate: open models %q: %w", modelsPath, err)
	}

	atoiOr := func(s string, def int) int {
		if n, err := strconv.Atoi(s); err == nil {
			return n
		}
		return def
	}

	mux := http.NewServeMux()
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
	mux.HandleFunc("/frame", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		pos := host.World().Self.Position()
		if pos.X <= 0 && pos.Y <= 0 {
			http.Error(w, "host position not loaded", http.StatusServiceUnavailable)
			return
		}
		v := buildLiveView(host, pos)
		v.Rotation = atoiOr(q.Get("rot"), cfg.renderRotation) & 0xff
		v.Zoom = atoiOr(q.Get("zoom"), cfg.renderZoom)
		v.W = atoiOr(q.Get("w"), cfg.renderW)
		v.H = atoiOr(q.Get("h"), cfg.renderH)
		v.AnimFrame = atoiOr(q.Get("anim"), 0) // model-swap frame (fires/torches flicker)
		png, err := render.RenderView(land, f, bundle, v)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write(png)
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
</style></head><body>
<img id="v">
<div id="hud"></div>
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
  'hold < >  rotate    + -  zoom    [ ]  view size'; }
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
const HANDLED=['ArrowLeft','ArrowRight','+','=','-','_','[',']'];
document.addEventListener('keydown',e=>{
  if(!HANDLED.includes(e.key)) return;
  e.preventDefault();
  keys[e.key]=true;
  // + zooms IN (camera closer = SMALLER zoom/distance); - zooms OUT. (Was inverted.)
  if(e.key==='+'||e.key==='=') zoom=Math.max(zoom-150,250);
  else if(e.key==='-'||e.key==='_') zoom=Math.min(zoom+150,4000);
  else if(e.key==='[') { w=Math.max(w-128,256); h=Math.max(h-84,168); }
  else if(e.key===']') { w=Math.min(w+128,1280); h=Math.min(h+84,840); }
  drawHud();
});
document.addEventListener('keyup',e=>{ keys[e.key]=false; });
setInterval(refreshPos,500);
setInterval(()=>{anim=(anim+1)&0x3fffffff;},160); // model-swap animation pace (~6/s); fires/torches flicker
setInterval(tick,66);   // ~15fps target; tick() no-ops while a frame is in flight
refreshPos(); drawHud();
</script></body></html>`
