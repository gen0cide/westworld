package main

// clientPage is the full remote-client SPA.  Served at GET / by serveClient.
// fmt.Fprintf verbs (in order): %s=username (title), %d=zoom, %d=w, %d=h.
// Every literal percent in the CSS/JS is escaped as %%.  No net/http, no Go
// logic lives in this file — it is a pure page asset (doc 40-browser-ui §4).
const clientPage = `<!doctype html><html><head><meta charset="utf-8">
<title>cradle client &#8212; %s</title>
<style>
html,body{margin:0;height:100%%;background:#000;overflow:hidden;
          font:13px monospace;color:#9f9}
#app{position:absolute;inset:0;display:grid;
     grid-template-columns:1fr 232px;
     grid-template-rows:1fr 168px;
     grid-template-areas:"view side" "chat side"}
#view{grid-area:view;position:relative;overflow:hidden;background:#000}
#v{position:absolute;inset:0;width:100%%;height:100%%;
   object-fit:contain;image-rendering:pixelated;cursor:crosshair}
#chat{grid-area:chat;display:flex;flex-direction:column;
      border-top:2px solid #1a3d1a;background:#0a140a}
#side{grid-area:side;display:flex;flex-direction:column;
      border-left:2px solid #1a3d1a;background:#0a140a;overflow:hidden}
#hud{position:absolute;left:8px;top:6px;white-space:pre;
     text-shadow:0 0 3px #000;pointer-events:none}
#label{position:absolute;left:8px;bottom:8px;color:#ff8;
       text-shadow:0 0 3px #000;pointer-events:none}
#flash{position:absolute;right:10px;top:8px;color:#ff8;opacity:0;
       transition:opacity .2s;pointer-events:none;font-size:15px}
#menu{position:absolute;display:none;z-index:50;min-width:120px;
      background:#23282b;border:1px solid #000;
      box-shadow:2px 2px 0 #000;font:13px monospace}
#menu .hdr{background:#0a0;color:#000;padding:2px 6px;font-weight:bold}
#menu .opt{padding:2px 8px;color:#fff;cursor:pointer;white-space:nowrap}
#menu .opt:hover{background:#0a0;color:#000}
#inv{display:grid;grid-template-columns:repeat(6,36px);gap:1px;padding:4px}
.cell{width:36px;height:36px;background:#1a1f1a;border:1px solid #000;
      position:relative;cursor:pointer;display:flex;align-items:center;
      justify-content:center;color:#cfc;font-size:11px;text-align:center}
.cell.wield{outline:1px solid #ff0}
.cell .qty{position:absolute;right:1px;bottom:0;color:#ff0;font-size:9px}
#equip{padding:4px;border-bottom:2px solid #1a3d1a;font-size:12px;line-height:1.4}
#equip .row{display:flex;justify-content:space-between}
#chatlog{flex:1;overflow-y:auto;padding:4px 6px;line-height:1.35;
         white-space:pre-wrap;word-break:break-word}
#chatlog .you{color:#0ff}
#chatlog .npc{color:#ff8}
#chatlog .pm{color:#f9f}
#chatlog .sys{color:#9a9}
#chatinput{border:0;border-top:1px solid #1a3d1a;background:#06100a;
           color:#9f9;font:13px monospace;padding:4px 6px;outline:none;
           width:100%%;box-sizing:border-box}
</style>
</head><body>
<div id="app">
  <div id="view">
    <img id="v" draggable="false">
    <div id="hud"></div>
    <div id="label"></div>
    <div id="flash"></div>
    <div id="menu"></div>
  </div>
  <div id="chat">
    <div id="chatlog"></div>
    <input id="chatinput" autocomplete="off" spellcheck="false"
           placeholder="say&#8230;  ::cmd  @name pm">
  </div>
  <div id="side">
    <div id="equip"></div>
    <div id="inv"></div>
  </div>
</div>
<script>
// ---- module-level state ------------------------------------------------
// fmt.Fprintf injects zoom, w, h (in that order, matching §6.3 contract)
let rot=64, zoom=%d, w=%d, h=%d, anim=0;
const ROT=4;           // yaw units advanced per tick while a rotate key is held
const keys={};
// Stretch-goal toggles (doc 40 §2.3 / §2.4) — disabled for M1
const HOVER=false, DRAG=false;

const img       = document.getElementById('v');
const hud       = document.getElementById('hud');
const lbl       = document.getElementById('label');
const menu      = document.getElementById('menu');
const chatlog   = document.getElementById('chatlog');
const chatinput = document.getElementById('chatinput');
const invEl     = document.getElementById('inv');
const equipEl   = document.getElementById('equip');

let busy=false, stateBusy=false;
let pos={x:0,y:0,plane:0};
let lastChatSeq=-1;
let stats={combatLevel:0,hp:0,maxHp:0};

// ---- flash toast (identical to spectatePage) ---------------------------
function flash(msg){
  const o=document.getElementById('flash');
  if(!o) return;
  o.textContent=msg;
  o.style.opacity=1;
  setTimeout(()=>o.style.opacity=0, 1500);
}

// ---- HUD ---------------------------------------------------------------
function drawHud(){
  hud.textContent=
    'host ('+pos.x+', '+pos.y+')  plane '+pos.plane+'\n'+
    'combat '+stats.combatLevel+'  hp '+stats.hp+'/'+stats.maxHp+'\n'+
    'rot '+rot+'  zoom '+zoom+'  '+w+'x'+h+'\n'+
    'left-click act \xB7 right-click menu \xB7 < >  rotate \xB7 +− zoom';
}

// ---- pos poll ----------------------------------------------------------
async function refreshPos(){
  try{ pos=await (await fetch('/pos')).json(); }catch(e){}
}

// ---- frame loop (mirrors spectatePage tick() exactly) ------------------
function tick(){
  if(keys['ArrowLeft'])  rot=(rot-ROT+256)&255;
  if(keys['ArrowRight']) rot=(rot+ROT)&255;
  if(busy) return;
  busy=true;
  const u='/frame?rot='+rot+'&zoom='+zoom+'&w='+w+'&h='+h+'&anim='+anim+'&t='+performance.now();
  const n=new Image();
  n.onload=()=>{ img.src=n.src; busy=false; drawHud(); };
  n.onerror=()=>{ busy=false; };
  n.src=u;
}

// ---- screenToFrame: undo object-fit:contain letterboxing ---------------
// Mirrors spectatePage's click handler exactly; the single source of (px,py)
// fed to /walk and /pick.
function screenToFrame(e){
  const r=img.getBoundingClientRect();
  const scale=Math.min(r.width/w, r.height/h);
  if(scale<=0) return null;
  const px=Math.round((e.clientX-(r.left+(r.width-w*scale)/2))/scale);
  const py=Math.round((e.clientY-(r.top+(r.height-h*scale)/2))/scale);
  if(px<0||py<0||px>=w||py>=h) return null;
  return {px,py};
}

// ---- walk (terrain fallback / retarget) --------------------------------
function walk(px,py){
  fetch('/walk?px='+px+'&py='+py+'&rot='+rot+'&zoom='+zoom+'&w='+w+'&h='+h);
}

// ---- pick: POST /pick -> PickResponse {candidates:[]} ------------------
// extra is merged into the request body (e.g. {slot:N} for inv fallback).
// slot defaults to -1 (no inventory slot — a plain screen pick); the inventory
// right-click fallback overrides it via extra={slot} so the server takes the
// inventory branch (spec 50-impl-spec §6.1: PickRequest.Slot, -1 default).
async function pick(px,py,extra){
  const body=Object.assign({px,py,rot,zoom,w,h,slot:-1},extra||{});
  try{
    const r=await fetch('/pick',{
      method:'POST',
      headers:{'Content-Type':'application/json'},
      body:JSON.stringify(body)
    });
    if(!r.ok) return {candidates:[]};
    return r.json();
  }catch(_){ return {candidates:[]}; }
}

// ---- act: POST /act {ref, optionId} -> {ok, message?} -----------------
// optionId is the int index into that candidate's options[] (50-impl-spec C2).
async function act(ref,optionId,extra){
  const body=Object.assign({ref,optionId},extra||{});
  try{
    const r=await fetch('/act',{
      method:'POST',
      headers:{'Content-Type':'application/json'},
      body:JSON.stringify(body)
    });
    const j=await r.json();
    if(j.message) flash(j.message);
  }catch(e){
    flash('act: '+e.message);
  }
}

// ---- context menu ------------------------------------------------------
function closeMenu(){
  menu.style.display='none';
  while(menu.firstChild) menu.removeChild(menu.firstChild);
}

// openMenu(screenX, screenY, candidates[])
// candidates: [{ref, label, options:[{id,verb}]}]
// Layout (doc 40 §5.5):
//   single candidate: option text = "Verb Label" (compact, no header)
//   multiple candidates: .hdr separator per candidate, option text = "Verb"
function openMenu(x,y,candidates){
  closeMenu();
  if(!candidates||!candidates.length) return;
  const multi=candidates.length>1;
  for(const cand of candidates){
    if(multi){
      const hdr=document.createElement('div');
      hdr.className='hdr';
      hdr.textContent=cand.label||'';
      menu.appendChild(hdr);
    }
    for(const opt of (cand.options||[])){
      const row=document.createElement('div');
      row.className='opt';
      row.textContent=multi ? opt.verb : (opt.verb+' '+cand.label);
      // Capture loop variables explicitly for stable handler closure
      const captRef=cand.ref;
      const captId=opt.id;
      row.addEventListener('click',ev=>{
        ev.stopPropagation();
        act(captRef,captId);
        closeMenu();
      });
      menu.appendChild(row);
    }
  }
  // Clamp to window (children must be in DOM first for scrollWidth to be valid)
  const mw=menu.scrollWidth||140, mh=menu.scrollHeight||80;
  menu.style.left=Math.min(x,window.innerWidth-mw-4)+'px';
  menu.style.top =Math.min(y,window.innerHeight-mh-4)+'px';
  menu.style.display='block';
}

// ---- defaultAct: left-click fires the top option of nearest candidate --
// Uniform path (doc 40 §2.1): always act top option — Layer 2 decides whether
// that is walk-here or an interaction, keeping one code path in the UI.
function defaultAct(menuJSON,fallbackPx,fallbackPy){
  const cands=(menuJSON&&menuJSON.candidates)||[];
  if(!cands.length){ if(fallbackPx!=null) walk(fallbackPx,fallbackPy); return; }
  const opts=(cands[0]||{}).options||[];
  if(!opts.length){  if(fallbackPx!=null) walk(fallbackPx,fallbackPy); return; }
  act(cands[0].ref, opts[0].id);
}

// ---- inventory ---------------------------------------------------------
const INV_SLOTS=30;  // RSC inventory is exactly 30 slots (doc 40 §1)

function initInventory(){
  // Build the fixed 30-cell grid once; event listeners delegate via closest().
  for(let i=0;i<INV_SLOTS;i++){
    const c=document.createElement('div');
    c.className='cell';
    c.dataset.slot=String(i);
    invEl.appendChild(c);
  }

  // Left-click: fire the slot's default option (doc 40 §2.4).
  invEl.addEventListener('click',e=>{
    const cell=e.target.closest('.cell');
    if(!cell) return;
    const itemId=parseInt(cell.dataset.itemId||'0');
    if(!itemId) return;  // empty slot
    const slot=parseInt(cell.dataset.slot);
    const defId=parseInt(cell.dataset.defaultOptionId||'0');
    const ref={kind:'inventory_item',index:0,x:0,y:0,dir:0,id:itemId,slot};
    act(ref,defId);
  });

  // Right-click: open item context menu (doc 40 §2.4).
  invEl.addEventListener('contextmenu',e=>{
    e.preventDefault();
    const cell=e.target.closest('.cell');
    if(!cell) return;
    const itemId=parseInt(cell.dataset.itemId||'0');
    if(!itemId) return;  // empty slot
    const slot=parseInt(cell.dataset.slot);
    const name=cell.dataset.name||('item '+itemId);
    let opts=[];
    try{ opts=JSON.parse(cell.dataset.options||'[]'); }catch(_){}
    const ref={kind:'inventory_item',index:0,x:0,y:0,dir:0,id:itemId,slot};
    if(opts.length){
      // Inline options from /state — no extra round-trip (doc 40 §2.4 preferred path)
      openMenu(e.pageX,e.pageY,[{ref,label:name,options:opts}]);
    }else{
      // Fallback: ask server via /pick with slot field
      pick(0,0,{slot}).then(j=>openMenu(e.pageX,e.pageY,j.candidates||[]));
    }
  });
}

// renderInventory: update the 30-cell grid from /state inventory[].
// Inventory is sparse (occupied slots only, keyed by slot).
function renderInventory(inv){
  const bySlot={};
  for(const s of (inv||[])) bySlot[s.slot]=s;
  const cells=invEl.querySelectorAll('.cell');
  cells.forEach(c=>{
    const i=parseInt(c.dataset.slot);
    const s=bySlot[i];
    while(c.firstChild) c.removeChild(c.firstChild);
    c.classList.remove('wield');
    if(!s||!s.itemId){
      c.dataset.itemId='0';
      c.dataset.name='';
      c.dataset.defaultOptionId='0';
      c.dataset.options='[]';
      return;
    }
    c.dataset.itemId=String(s.itemId);
    c.dataset.name=s.name||('item '+s.itemId);
    c.dataset.defaultOptionId=String(s.defaultOptionId||0);
    c.dataset.options=JSON.stringify(s.options||[]);
    if(s.wielded) c.classList.add('wield');
    // Short name to fit 36px cell
    const name=s.name||('item '+s.itemId);
    const span=document.createElement('span');
    span.textContent=name.length>8 ? name.slice(0,7)+'…' : name;
    c.appendChild(span);
    // Stack badge
    if(s.amount>1){
      const q=document.createElement('div');
      q.className='qty';
      q.textContent=s.amount>=1000 ? (s.amount/1000).toFixed(1)+'k' : String(s.amount);
      c.appendChild(q);
    }
  });
}

// renderEquipment: rebuild the equipment summary from /state equipment[].
// Each entry: {slot:"Weapon", sprite:13, itemId:0} (doc 50-impl-spec §C6).
function renderEquipment(eq){
  while(equipEl.firstChild) equipEl.removeChild(equipEl.firstChild);
  for(const e of (eq||[])){
    const row=document.createElement('div');
    row.className='row';
    const sl=document.createElement('span');
    sl.style.color='#9a9';
    sl.textContent=e.slot||'';
    const nm=document.createElement('span');
    nm.textContent=e.name||(e.sprite?'spr '+e.sprite:'');
    row.appendChild(sl);
    row.appendChild(nm);
    equipEl.appendChild(row);
  }
}

// ---- chat --------------------------------------------------------------
function initChat(){
  chatinput.addEventListener('keydown',e=>{
    if(e.key==='Enter'){
      const raw=chatinput.value;
      chatinput.value='';
      sendChat(raw);
    }else if(e.key==='Escape'){
      chatinput.blur();
    }
  });
}

// sendChat classifies raw input and POSTs /chat (doc 40 §2.5).
//   ::cmd or /cmd  -> kind:"command"
//   @name msg      -> kind:"pm"
//   else           -> kind:"say"
function sendChat(raw){
  const text=raw.trim();
  if(!text) return;
  let kind,to='',body=text;
  if(text.startsWith('::')||text.startsWith('/')){
    kind='command';
    body=text.startsWith('::') ? text.slice(2).trim() : text.slice(1).trim();
    if(!body){ flash('empty command'); return; }
  }else if(text.startsWith('@')){
    kind='pm';
    const sp=text.indexOf(' ');
    if(sp<0){ flash('usage: @name message'); return; }
    to=text.slice(1,sp).trim();
    body=text.slice(sp+1).trim();
    if(!to||!body){ flash('usage: @name message'); return; }
  }else{
    kind='say';
  }
  const req={kind,text:body};
  if(kind==='pm') req.to=to;
  fetch('/chat',{
    method:'POST',
    headers:{'Content-Type':'application/json'},
    body:JSON.stringify(req)
  }).then(r=>r.json())
    .then(j=>{ if(!j.ok&&j.message) flash(j.message); })
    .catch(e=>flash('chat: '+e.message));
}

// appendChat: append only lines with seq > lastChatSeq (doc 40 §5.3).
// Stays pinned to the newest line when the log was already at the bottom.
function appendChat(lines){
  if(!lines||!lines.length) return;
  const atBottom=chatlog.scrollTop+chatlog.clientHeight>=chatlog.scrollHeight-4;
  let added=false;
  for(const line of lines){
    if(line.seq<=lastChatSeq) continue;
    lastChatSeq=line.seq;
    const div=document.createElement('div');
    // Colour class by kind (50-impl-spec §C6: public/npc/private/system/self)
    switch(line.kind){
      case 'npc':     div.className='npc'; break;
      case 'private': div.className='pm';  break;
      case 'system':  div.className='sys'; break;
      case 'self':    div.className='you'; break;
      // 'public': no class — inherits default #9f9
    }
    const who=line.who ? '['+line.who+'] ' : '';
    div.textContent=who+line.text;
    chatlog.appendChild(div);
    added=true;
  }
  if(added&&atBottom) chatlog.scrollTop=chatlog.scrollHeight;
}

// ---- state poll --------------------------------------------------------
// GET /state -> {self, inventory, equipment, chat} (doc 40 §5.3, 50-impl §5)
async function pollState(){
  if(stateBusy) return;
  stateBusy=true;
  try{
    const r=await fetch('/state');
    if(r.ok) applyState(await r.json());
  }catch(_){}
  stateBusy=false;
}

function applyState(s){
  if(!s) return;
  if(s.inventory)  renderInventory(s.inventory);
  if(s.equipment)  renderEquipment(s.equipment);
  if(s.chat)       appendChat(s.chat);
  if(s.self){
    stats.combatLevel=s.self.combatLevel||0;
    stats.hp=s.self.hp||0;
    stats.maxHp=s.self.maxHp||0;
    // Keep pos in sync with the richer /state.self (more precise than /pos poll)
    if(s.self.x)        pos.x=s.self.x;
    if(s.self.y)        pos.y=s.self.y;
    if(s.self.plane!=null) pos.plane=s.self.plane;
  }
  drawHud();
}

// ---- keyboard ----------------------------------------------------------
// Mirror spectatePage's key set. Camera keys are suppressed while the chat
// input is focused so typing never rotates or zooms (doc 40 §2.5).
const HANDLED=['ArrowLeft','ArrowRight','+','=','-','_','[',']','p','c'];
document.addEventListener('keydown',e=>{
  if(document.activeElement===chatinput) return;
  if(!HANDLED.includes(e.key)) return;
  e.preventDefault();
  keys[e.key]=true;
  const cam='rot='+rot+'&zoom='+zoom+'&w='+w+'&h='+h;
  if     (e.key==='+'||e.key==='=')     zoom=Math.max(zoom-150,250);
  else if(e.key==='-'||e.key==='_')     zoom=Math.min(zoom+150,4000);
  else if(e.key==='['){w=Math.max(w-128,256);h=Math.max(h-84,168);}
  else if(e.key===']'){w=Math.min(w+128,1280);h=Math.min(h+84,840);}
  else if(e.key==='p'){ flash('shot…'); fetch('/shot?'+cam).then(()=>flash('shot saved')); }
  else if(e.key==='c'){ flash('clip… walk now (~3s)'); fetch('/clip?'+cam).then(()=>flash('clip saved')); }
  drawHud();
});
document.addEventListener('keyup',e=>{ keys[e.key]=false; });

// ---- close-menu triggers -----------------------------------------------
document.addEventListener('click',()=>closeMenu(),true);
document.addEventListener('scroll',()=>closeMenu(),true);
document.addEventListener('keydown',e=>{
  if(e.key==='Escape'){ closeMenu(); chatinput.blur(); }
});

// ---- viewport interactions ---------------------------------------------
img.addEventListener('click',e=>{
  const fp=screenToFrame(e);
  if(!fp) return;
  pick(fp.px,fp.py)
    .then(j=>defaultAct(j,fp.px,fp.py))
    .catch(()=>walk(fp.px,fp.py));
});

img.addEventListener('contextmenu',e=>{
  e.preventDefault();
  const fp=screenToFrame(e);
  if(!fp) return;
  const sx=e.pageX,sy=e.pageY;
  pick(fp.px,fp.py)
    .then(j=>openMenu(sx,sy,j.candidates||[]))
    .catch(()=>{});
});

// ---- init --------------------------------------------------------------
function init(){
  initInventory();
  initChat();
  refreshPos();
  pollState();
  drawHud();
}

// Three independent loops, none stacking on each other (doc 40 §3).
setInterval(tick,33);           // ~30fps frame; busy-guarded inside tick()
setInterval(pollState,400);     // inventory + chat + stats; stateBusy-guarded
setInterval(refreshPos,500);    // HUD coords; tiny, unguarded
setInterval(()=>{ anim=(anim+1)&0x3fffffff; },160); // model-swap animation

init();
</script>
</body></html>`
