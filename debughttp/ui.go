package debughttp

// dashboardHTML is the self-contained browser UI served at "/". It opens a
// WebSocket to /ws for the live event stream (EVERY bus event type, including
// agent_thought) and polls /state for the status panel. No build step, no
// external assets. "__USER__" is substituted with the host's username.
const dashboardHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>host · __USER__</title>
<style>
  :root { color-scheme: dark; }
  * { box-sizing: border-box; }
  body { margin:0; font:14px/1.5 ui-monospace,SFMono-Regular,Menlo,monospace; background:#0d1117; color:#c9d1d9; }
  header { padding:10px 16px; background:#161b22; border-bottom:1px solid #30363d; display:flex; align-items:center; gap:14px; }
  header h1 { font-size:15px; margin:0; font-weight:600; }
  #status { font-size:12px; padding:2px 8px; border-radius:10px; }
  #status.on { background:#1a472a; color:#7ee787; }
  #status.off { background:#5a1e1e; color:#ff7b72; }
  .ticker { color:#6e7681; font-size:11.5px; }
  .wrap { display:grid; grid-template-columns: 300px 1fr 380px; height:calc(100vh - 49px); }
  .panel { overflow:auto; padding:14px 16px; }
  #state { border-right:1px solid #30363d; }
  #stream { border-right:1px solid #30363d; }
  h2 { font-size:12px; text-transform:uppercase; letter-spacing:.5px; color:#8b949e; margin:0 0 8px; position:sticky; top:0; background:#0d1117; padding-bottom:4px; }
  .kv { display:flex; justify-content:space-between; padding:2px 0; }
  .kv span:last-child { color:#e6edf3; }
  .sec { margin:14px 0 6px; color:#58a6ff; font-size:12px; }
  .list { font-size:12.5px; }
  .list div { padding:1px 0; color:#adbac7; }
  .dialog { background:#1c2333; border:1px solid #30415d; border-radius:6px; padding:8px; margin-top:6px; font-size:12.5px; color:#cdd9e5; }
  .opt { color:#7ee787; }
  #thoughts { display:flex; flex-direction:column; gap:10px; }
  .thought { background:#161b22; border:1px solid #30363d; border-left:3px solid #58a6ff; border-radius:6px; padding:10px 12px; }
  .thought .top { display:flex; gap:10px; align-items:baseline; flex-wrap:wrap; }
  .thought .turn { color:#58a6ff; font-weight:600; }
  .thought .kind { font-size:11px; padding:1px 7px; border-radius:9px; background:#21262d; color:#8b949e; }
  .thought .pos { color:#6e7681; font-size:12px; margin-left:auto; }
  .thought .reason { margin:6px 0; color:#e6edf3; }
  .thought .percept { color:#8b949e; font-size:12px; margin:4px 0; white-space:pre-wrap; }
  .thought pre { background:#0d1117; border:1px solid #21262d; border-radius:5px; padding:8px; overflow:auto; font-size:12px; color:#a5d6ff; margin:6px 0 0; white-space:pre-wrap; }
  #feed { font-size:12px; }
  .ev { padding:3px 0; border-bottom:1px solid #161b22; display:flex; gap:8px; }
  .ev .k { flex:0 0 130px; color:#8b949e; }
  .ev .d { color:#adbac7; word-break:break-word; }
  .ev.k_agent_thought .k { color:#58a6ff; }
  .ev.k_chat_received .k, .ev.k_system_message .k { color:#7ee787; }
  .ev.k_experience_gain .k, .ev.k_level_up .k { color:#d2a8ff; }
  .ev.k_npc_damage .k, .ev.k_death .k { color:#ff7b72; }
  .ev.k_policy_veto { background:#2d1518; } .ev.k_policy_veto .k { color:#ff9492; }
  .filterbar { display:flex; gap:6px; margin-bottom:8px; flex-wrap:wrap; }
  .filterbar input { background:#0d1117; border:1px solid #30363d; color:#c9d1d9; border-radius:5px; padding:3px 7px; font:inherit; font-size:12px; width:140px; }
  label.chk { font-size:11.5px; color:#8b949e; }
</style>
</head>
<body>
<header>
  <h1>🤖 host · __USER__</h1>
  <span id="status" class="off">connecting…</span>
  <span class="ticker" id="evcount"></span>
</header>
<div class="wrap">
  <div class="panel" id="state">
    <h2>Live state</h2>
    <div id="vitals"></div>
    <div class="sec">Inventory</div><div class="list" id="inv"></div>
    <div class="sec">Nearby NPCs</div><div class="list" id="npcs"></div>
    <div class="sec">Dialog</div><div id="dialog"></div>
    <div class="sec">Recent messages</div><div class="list" id="msgs"></div>
  </div>
  <div class="panel" id="stream">
    <h2>Thought stream</h2>
    <div id="thoughts"></div>
  </div>
  <div class="panel" id="busp">
    <h2>Event bus (all types)</h2>
    <div class="filterbar">
      <input id="filter" placeholder="filter by kind…">
      <label class="chk"><input type="checkbox" id="hideThoughts"> hide thoughts</label>
    </div>
    <div id="feed"></div>
  </div>
</div>
<script>
(function(){
  var statusEl=document.getElementById('status'), evcount=document.getElementById('evcount');
  var thoughts=document.getElementById('thoughts'), feed=document.getElementById('feed');
  var filterEl=document.getElementById('filter'), hideThoughts=document.getElementById('hideThoughts');
  var nEvents=0, nThoughts=0;

  function esc(s){ return (s==null?'':String(s)).replace(/[&<>]/g,function(c){return {'&':'&amp;','<':'&lt;','>':'&gt;'}[c];}); }

  function addThought(d){
    nThoughts++;
    var el=document.createElement('div'); el.className='thought';
    var h='<div class="top"><span class="turn">turn '+esc(d.turn)+'</span>'
      +'<span class="kind">'+esc(d.move_kind)+'</span><span class="kind">'+esc(d.trigger)+'</span>'
      +'<span class="pos">'+esc(d.pos)+' · hp '+esc(d.hp)+'</span></div>';
    if(d.reasoning) h+='<div class="reason">“'+esc(d.reasoning)+'”</div>';
    if(d.perception) h+='<div class="percept">👁 '+esc(d.perception)+'</div>';
    if(d.dsl) h+='<pre>'+esc(d.dsl)+'</pre>';
    el.innerHTML=h;
    thoughts.insertBefore(el, thoughts.firstChild);
    while(thoughts.children.length>80) thoughts.removeChild(thoughts.lastChild);
  }

  // Compact one-line summary per event kind (falls back to JSON).
  function summarize(kind, d){
    if(!d) return '';
    switch(kind){
      case 'agent_thought': return (d.move_kind||'')+' — '+(d.reasoning||'');
      case 'policy_veto': return '⛔ '+(d.action||d.Action||'')+' blocked ['+(d.rule||d.Rule||'')+']: '+(d.reason||d.Reason||'');
      case 'chat_received': return (d.Speaker||d.speaker||'?')+': '+(d.Message||d.message||'');
      case 'private_message': return (d.Sender||d.sender||'?')+' » '+(d.Message||d.message||'');
      case 'system_message': return d.Message||d.message||'';
      case 'experience_gain': return 'skill '+(d.Skill!=null?d.Skill:d.skill)+' +'+(d.XP!=null?d.XP:d.xp);
      case 'level_up': return 'skill '+(d.Skill!=null?d.Skill:d.skill)+' → '+(d.Level!=null?d.Level:d.level);
      case 'npc_damage': return 'npc#'+(d.Index!=null?d.Index:d.index)+' dmg '+(d.Damage!=null?d.Damage:d.damage);
      default: var s=JSON.stringify(d); return s.length>160?s.slice(0,160)+'…':s;
    }
  }

  function passFilter(kind){
    if(hideThoughts.checked && kind==='agent_thought') return false;
    var f=filterEl.value.trim().toLowerCase();
    return !f || kind.indexOf(f)>=0;
  }

  function addEvent(rec){
    if(!passFilter(rec.kind)) return;
    var el=document.createElement('div'); el.className='ev k_'+rec.kind;
    el.innerHTML='<span class="k">'+esc(rec.kind)+'</span><span class="d">'+esc(summarize(rec.kind, rec.data))+'</span>';
    feed.insertBefore(el, feed.firstChild);
    while(feed.children.length>300) feed.removeChild(feed.lastChild);
  }

  function connect(){
    var proto=location.protocol==='https:'?'wss':'ws';
    var ws=new WebSocket(proto+'://'+location.host+'/ws');
    ws.onopen=function(){ statusEl.className='on'; statusEl.textContent='● live'; };
    ws.onclose=function(){ statusEl.className='off'; statusEl.textContent='○ disconnected — retrying'; setTimeout(connect,1500); };
    ws.onmessage=function(ev){
      nEvents++; evcount.textContent=nEvents+' events · '+nThoughts+' thoughts';
      var rec; try{ rec=JSON.parse(ev.data); }catch(e){ return; }
      addEvent(rec);
      if(rec.kind==='agent_thought' && rec.data) addThought(rec.data);
    };
  }
  connect();

  function row(k,v){ return '<div class="kv"><span>'+k+'</span><span>'+esc(v)+'</span></div>'; }
  function refreshState(){
    fetch('/state').then(function(r){return r.json();}).then(function(s){
      document.getElementById('vitals').innerHTML=row('position','('+s.x+', '+s.y+')')+row('hp',s.hp+' / '+s.max_hp)+row('combat lvl',s.combat_level)+row('fatigue',s.fatigue+'%')+row('inv free',s.inventory_free);
      document.getElementById('inv').innerHTML=(s.inventory&&s.inventory.length)?s.inventory.map(function(i){return '<div>'+esc(i.name||('item#'+i.item_id))+(i.amount>1?(' ×'+i.amount):'')+'</div>';}).join(''):'<div>(empty)</div>';
      var seen={},npcs=[]; (s.npcs||[]).forEach(function(n){var nm=n.name||('npc#'+n.type_id); if(!seen[nm]){seen[nm]=1;npcs.push(nm);}});
      document.getElementById('npcs').innerHTML=npcs.length?npcs.map(function(n){return '<div>'+esc(n)+'</div>';}).join(''):'<div>(none)</div>';
      var dl=''; if(s.dialog_text) dl+='<div class="dialog">'+esc(s.dialog_text)+'</div>';
      if(s.dialog_open&&s.dialog_options) dl+='<div class="dialog">'+s.dialog_options.map(function(o,i){return '<div class="opt">'+(i+1)+'. '+esc(o)+'</div>';}).join('')+'</div>';
      document.getElementById('dialog').innerHTML=dl||'<div class="list">(none open)</div>';
      var m=s.recent_server_messages||[];
      document.getElementById('msgs').innerHTML=m.slice(-8).reverse().map(function(x){return '<div>'+esc(x)+'</div>';}).join('')||'<div>(none)</div>';
    }).catch(function(){});
  }
  refreshState(); setInterval(refreshState,1500);
})();
</script>
</body>
</html>`
