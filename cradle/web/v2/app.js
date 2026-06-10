// cradle v2 — app.js: hash router, global selection, the polling data store
// (the Stage-1 bridge until the SSE endpoints land — spec §3, last paragraph),
// and the FLEET LIST view (§2.1). Vanilla JS, no build, no CDN.
//
// ===== store + mount contract (read me, tablet.js) ===========================
// window.CradleV2.renderTablet(el, hostName, store) is called by the router
// each time #/host/{name} is entered, with a fresh empty container `el`. The
// router has ALREADY called store.select(hostName), so the per-host pollers
// are live. renderTablet may return a cleanup function; it is invoked on route
// change / host switch. While renderTablet is missing the route renders a stub.
//
// store API:
//   store.subscribe(key, fn) -> unsubscribe()
//        fn(value, meta) fires immediately with the cached value (if any) and
//        again on every update. meta = {at, stale, error, hidden}:
//          stale=true  → the fetch failed but `value` is the LAST GOOD data;
//                        render it with a staleness badge, NEVER throw (§4).
//          hidden=true → the endpoint answered 404/501 (not built yet / mesa
//                        proxy unimplemented); hide the pane entirely.
//   store.get(key) / store.status(key)      snapshot reads
//   store.selected                          selected host name or null (global
//                                           selection state — survives routes)
//   store.vitals(name) -> {hp:[], fat:[]}   20-sample history @1s per host
//   store.lifecycle(name, 'pause'|'resume'|'restart'|'stop')
//   store.refreshRoster()                   force an immediate roster poll
//   store.fetchJSON(path, {method, body})   shared fetch (eval/script/analysis)
//
// keys & shapes (cadence; endpoint — Go source of truth in parens):
//   'roster' (1s; GET /api/hosts) -> [HostStatus] (cradle/registry.go):
//        {name, status, mesa?, goal?, autonomous, restarts, started_at, err?,
//         current_routine?, current_routine_source?, current_line?,
//         line_trace?[], line_seq?, live, x?, y?, hp?, max_hp?, analysis?}
//        floor = Math.floor(y/944); east = smaller x (spec Appendix B).
//   'host' (1s) -> the selected host's HostStatus row out of 'roster'.
//   'routine_line' (paced ~90ms) -> int. The REPLAYED current line for the
//        selected host (the v1 line-trace pacing trick): the 1s poll delivers
//        line_trace (last 64 executed lines) + line_seq; the store enqueues the
//        delta and steps the published pointer so fast statements between waits
//        are SEEN. Highlight this, not host.current_line.
//   'state' (2s; GET /api/hosts/{n}/debug/state — debughttp/server.go
//        stateSnapshot) -> {x, y, hp, max_hp, fatigue, combat_level,
//        inventory_used, inventory_free, inventory[{slot,item_id,name,amount}],
//        npcs[{server_index,type_id,name,x,y}], players[{index,name,x,y}],
//        dialog_open, dialog_text?, dialog_options?[],
//        recent_server_messages?[], skills[{name,level,cur,xp}],
//        bus_subs_star, bus_subs_total, event_seq}
//   'mind' (5s; GET .../debug/mind?full=1 — mindSnapshot) ->
//        {knowledge[{subject,kind,top_claim?,confidence,provenance?,familiar,
//                    beliefs,tags?}],
//         relationships[{name,grade,trust,affinity,grievance,familiar,tags?,
//                        interactions?,last_at?}],  // G-16; last_at unix secs
//         goal_nodes[{id,kind,label,status,progress,tags?}],
//         goal_edges[{from,to,rel}], open_questions[goal-node], event_seq}
//        With ?full=1 (gap G-3) knowledge[].beliefs becomes
//        [{claim,provenance,alpha,beta,at}] (at = unix secs); today it is
//        still the count int — branch on Array.isArray(f.beliefs).
//   'events' (1s incremental; GET .../debug/events?since=&limit= — the 4000
//        ring, debughttp handleEvents) -> capped array (500) of
//        {seq, at, kind, type, data}, oldest first. Kinds of interest:
//        agent_thought{turn,trigger,goal?,pos,hp,perception,reasoning,
//        move_kind,dsl?}, routine_note{Text}, policy_veto{action,rule,reason},
//        system_message{Message}, chat_received{Speaker,Message},
//        other_player_chat{PlayerIndex,MessageText},
//        private_message{Sender,Message} (event/events.go, event/agent.go).
//   'aspirations' (5s; GET .../debug/aspirations — gap G-1, hidden until it
//        lands) -> [{id, label, goals_done, goals_active, goals_open,
//        last_touched, neglected}] (runtime.AspirationStatus + json tags).
//   'fog' (5s; GET .../debug/fog — gap G-2n) -> {coverage:{frac,seen,total},
//        known_pois, frontiers[{dir,x,y,dist}], here:{terrain,contents}}.
//   'where' (5s; GET .../debug/where — gap G-11) -> object, or {text:"…"} when
//        the endpoint returns plain text (LocationSummary narration).
//   'decisions' (5s; GET .../debug/decisions?tail=200 — gap G-4) ->
//        [{at, trigger, kind, reasoning, goal?}] (runtime decisionRecord; a
//        {decisions:[…]} wrapper is unwrapped by the store).
//   'persona' (once per selection; GET /api/hosts/{n}/persona — gap G-7) ->
//        persona JSON; 501 (mesa proxy off) and 404 (not landed) → hidden.
//
// §4 helpers live in render.js: renderKeyed/esc/setText/setClass, followPane
// (pointer-follow panes), feedPin (tail-pinned feeds + '▼ N new' pill),
// drawSpark. The DOM scroll-into-view API is BANNED in live panes — it scrolls
// every ancestor; move only a pane's own scrollTop.
// =============================================================================
(function () {
  'use strict';
  const NS = (window.CradleV2 = window.CradleV2 || {});
  const { esc, setText, setClass, renderKeyed, drawSpark } = NS;

  // ---- data store -----------------------------------------------------------
  const store = NS.store = (() => {
    const data = {};   // key -> last good value
    const meta = {};   // key -> {at, stale, error, hidden}
    const subs = {};   // key -> Set<fn>
    const vitals = {}; // host -> {hp:[], fat:[]} 20-sample history
    let selected = null;
    let detailTimers = [];
    let eventsLastSeq = 0;
    let lineState = null; // paced routine-line replay for the selected host

    function emit(key) {
      const fns = subs[key];
      if (!fns) return;
      for (const fn of fns) {
        try { fn(data[key], meta[key] || {}); }
        catch (e) { console.error('subscriber error', key, e); } // a bad pane never breaks the store
      }
    }
    function set(key, value) {
      data[key] = value;
      meta[key] = { at: Date.now(), stale: false, error: null, hidden: false };
      emit(key);
    }
    // fail degrades to stale-with-badge: the last good value stays on screen.
    function fail(key, err) {
      meta[key] = Object.assign({ at: 0 }, meta[key], { stale: true, error: String((err && err.message) || err) });
      emit(key);
    }
    function hide(key) {
      data[key] = null;
      meta[key] = { at: Date.now(), stale: false, error: null, hidden: true };
      emit(key);
    }

    async function fetchJSON(path, opts) {
      const r = await fetch(path, opts);
      const t = await r.text();
      let j = null;
      try { j = t ? JSON.parse(t) : null; } catch (_) { j = t; }
      if (!r.ok) {
        const e = new Error((j && j.error) || ('HTTP ' + r.status));
        e.status = r.status;
        throw e;
      }
      return j;
    }

    function unwrapList(v, names) {
      if (Array.isArray(v)) return v;
      if (v && typeof v === 'object') {
        for (const n of names) if (Array.isArray(v[n])) return v[n];
      }
      return v == null ? [] : v;
    }

    // -- roster: always-on @1s ------------------------------------------------
    async function pollRoster() {
      let hosts;
      try { hosts = await fetchJSON('/api/hosts'); }
      catch (e) { fail('roster', e); return; }
      if (!Array.isArray(hosts)) hosts = [];
      sampleVitals(hosts);
      set('roster', hosts);
      if (selected) {
        const h = hosts.find(x => x.name === selected);
        if (h) { set('host', h); feedLine(h); }
      }
    }
    function sampleVitals(hosts) {
      for (const h of hosts) {
        const v = vitals[h.name] || (vitals[h.name] = { hp: [], fat: [] });
        v.hp.push(h.live ? (h.hp || 0) : null); // hp omitempty: absent at exactly 0
        // fatigue is not on HostStatus yet (fleet-level fatigue = gap G-A,
        // Stage 2); read it defensively so the sparkline lights up when it lands.
        v.fat.push(typeof h.fatigue === 'number' ? h.fatigue : null);
        if (v.hp.length > 20) v.hp.shift();
        if (v.fat.length > 20) v.fat.shift();
      }
    }
    pollRoster();
    setInterval(pollRoster, 1000);

    // -- paced routine-line replay (the v1 pacing trick, store-side) ----------
    function feedLine(h) {
      if (!lineState) return;
      const src = h.current_routine_source || '';
      if (src !== lineState.src) { // new routine — jump, don't replay history
        lineState = { src, queue: [], lastSeq: h.line_seq || 0 };
        set('routine_line', h.current_line || 0);
        return;
      }
      const seq = h.line_seq || 0, trace = h.line_trace || [];
      if (seq > lineState.lastSeq) {
        const n = Math.min(seq - lineState.lastSeq, trace.length);
        for (let i = trace.length - n; i < trace.length; i++) lineState.queue.push(trace[i]);
        lineState.lastSeq = seq;
        if (lineState.queue.length > 64) lineState.queue = lineState.queue.slice(-64); // bound the backlog
      }
    }
    setInterval(() => { // replay stepper: one line per tick, 3 when behind
      if (!lineState || !lineState.queue.length) return;
      const step = lineState.queue.length > 12 ? 3 : 1;
      let line;
      for (let i = 0; i < step && lineState.queue.length; i++) line = lineState.queue.shift();
      if (line != null) set('routine_line', line);
    }, 90);

    // -- selection + per-host pollers ------------------------------------------
    function select(name) {
      if (selected !== name) {
        selected = name;
        eventsLastSeq = 0;
        lineState = { src: null, queue: [], lastSeq: 0 };
        // clear per-selection keys so a host switch never shows stale neighbors
        for (const k of ['host', 'state', 'mind', 'events', 'aspirations', 'fog', 'where', 'decisions', 'persona', 'routine_line']) {
          data[k] = undefined;
          meta[k] = {};
          emit(k);
        }
      }
      attach();
    }
    function detach() {
      for (const t of detailTimers) clearInterval(t);
      detailTimers = [];
    }
    function attach() {
      detach();
      const name = selected;
      if (!name) return;
      const enc = encodeURIComponent(name);
      const dbg = p => `/api/hosts/${enc}/debug/${p}`;

      // generic poller; opt.optional → 404/501 means "endpoint not built yet"
      // (gaps G-1/2n/4/7/11): hide the pane instead of erroring it.
      const poll = (key, ms, url, opt) => {
        const tick = async () => {
          if (selected !== name) return;
          try {
            let v = await fetchJSON(url);
            if (opt && opt.unwrap) v = unwrapList(v, opt.unwrap);
            if (opt && opt.textWrap && typeof v === 'string') v = { text: v };
            if (selected === name) set(key, v);
          } catch (e) {
            if (selected !== name) return;
            if (opt && opt.optional && (e.status === 404 || e.status === 501)) hide(key);
            else fail(key, e);
          }
        };
        tick();
        detailTimers.push(setInterval(tick, ms));
      };

      poll('state', 2000, dbg('state'));
      poll('mind', 5000, dbg('mind?full=1'));
      poll('aspirations', 5000, dbg('aspirations'), { optional: true, unwrap: ['aspirations'] });
      poll('fog', 5000, dbg('fog'), { optional: true });
      poll('where', 5000, dbg('where'), { optional: true, textWrap: true });
      poll('decisions', 5000, dbg('decisions?tail=200'), { optional: true, unwrap: ['decisions', 'records'] });

      // persona: once per selection (gap G-7). 501 = mesa proxy unimplemented,
      // 404 = endpoint not landed — both hide the persona band.
      if (data.persona === undefined) {
        (async () => {
          try {
            const v = await fetchJSON(`/api/hosts/${enc}/persona`);
            if (selected === name) set('persona', v);
          } catch (e) {
            if (selected !== name) return;
            if (e.status === 501 || e.status === 404) hide('persona');
            else fail('persona', e);
          }
        })();
      }

      // events: incremental tail @1s. The ring serves oldest-first, so a naive
      // limit fetch returns the OLDEST records — probe latest_seq first, then
      // backfill the last ~200 and follow with since=.
      const pollEvents = async () => {
        if (selected !== name) return;
        try {
          if (!eventsLastSeq) {
            const probe = await fetchJSON(dbg('events?limit=1'));
            // re-check after EVERY await: a slow probe from the previous host
            // must not clobber the new selection's (shared) cursor — that
            // strands the next host's feed behind a far-future seq forever.
            if (selected !== name) return;
            eventsLastSeq = Math.max(0, (probe.latest_seq || 0) - 200);
          }
          const r = await fetchJSON(dbg(`events?since=${eventsLastSeq}&limit=1000`));
          if (selected !== name) return;
          const recs = (r && r.events) || [];
          if (recs.length) {
            eventsLastSeq = recs[recs.length - 1].seq;
            const ring = (data.events || []).concat(recs);
            set('events', ring.length > 500 ? ring.slice(ring.length - 500) : ring);
          } else if (data.events === undefined || (meta.events && meta.events.stale)) {
            set('events', data.events || []); // first sight / recovery refresh
          }
        } catch (e) {
          if (selected === name) fail('events', e);
        }
      };
      pollEvents();
      detailTimers.push(setInterval(pollEvents, 1000));
    }

    async function lifecycle(name, action) {
      if (action === 'stop' && !window.confirm('stop ' + name + '?')) return;
      try { await fetchJSON(`/api/hosts/${encodeURIComponent(name)}/${action}`, { method: 'POST' }); }
      catch (e) { window.alert(action + ' ' + name + ': ' + ((e && e.message) || e)); }
      pollRoster();
    }

    function subscribe(key, fn) {
      (subs[key] || (subs[key] = new Set())).add(fn);
      if (data[key] !== undefined || (meta[key] && meta[key].at !== undefined)) {
        try { fn(data[key], meta[key] || {}); } catch (e) { console.error('subscriber error', key, e); }
      }
      return () => subs[key].delete(fn);
    }

    return {
      subscribe,
      get: k => data[k],
      status: k => Object.assign({}, meta[k]),
      select,
      detach,
      get selected() { return selected; },
      vitals: n => vitals[n],
      lifecycle,
      refreshRoster: pollRoster,
      fetchJSON,
    };
  })();

  // ---- shared derivations ----------------------------------------------------
  // floor band-encoding + RSC compass mirror: spec Appendix B.
  NS.floorOf = y => Math.floor((y || 0) / 944);

  // Activity glyph derived client-side from goal/current_routine text until the
  // server-side glyph lands (gap G-A, Stage 2).
  NS.activityGlyph = function (h) {
    if (h.status === 'crashed' || h.status === 'restarting' || h.status === 'stopped') return '✕';
    if (h.stalled) return '▣'; // not exposed yet — defensive (alert engine, Stage 2)
    const t = ((h.current_routine || '') + ' ' + (h.goal || '')).toLowerCase();
    if (/sleep|\brest\b/.test(t)) return 'zZ';
    if (/attack|fight|kill|combat|slay|duel/.test(t)) return '⚔';
    if (/\bsay\b|chat|\bask\b|talk|greet|befriend|reply|converse/.test(t)) return '✋';
    if (/forage|find|\bget\b|gather|fetch|collect|mine|chop|fish|buy|acquire|search|explore/.test(t)) return '◔';
    return '·';
  };

  // ---- chrome: rollup + staleness badge ---------------------------------------
  function wireChrome() {
    const rollup = document.getElementById('rollup');
    const conn = document.getElementById('conn');
    store.subscribe('roster', (hosts, m) => {
      conn.hidden = !m.stale;
      if (!hosts) return;
      const c = {};
      for (const h of hosts) c[h.status] = (c[h.status] || 0) + 1;
      setText(rollup, `●${c.running || 0} up · ${c.paused || 0} paused · ${(c.crashed || 0) + (c.restarting || 0)} crashed`);
    });
    // alert strip stays empty: the alert engine + GET /api/alerts is Stage 2
    // (gap G-B). The band is fixed-height so its arrival shifts nothing.
  }

  // ---- router ------------------------------------------------------------------
  let viewCleanup = null;
  function route() {
    const hash = location.hash || '#/fleet';
    const m = hash.match(/^#\/host\/([^/?]+)/);
    if (viewCleanup) { try { viewCleanup(); } catch (e) { console.error(e); } viewCleanup = null; }
    const view = document.getElementById('view');
    const active = m ? 'fleet' : (hash.startsWith('#/archive') ? 'archive' : 'fleet');
    document.querySelectorAll('#tabs .tab').forEach(t => setClass(t, 'active', t.dataset.route === active));
    if (m) {
      const name = decodeURIComponent(m[1]);
      store.select(name); // selection is global; pollers attach here
      view.innerHTML = '';
      viewCleanup = mountTablet(view, name);
    } else if (hash.startsWith('#/archive')) {
      store.detach(); // keep the selection highlight, stop detail polling
      view.innerHTML = '<div class="stub-panel">Archive — Stage 3</div>';
    } else {
      store.detach();
      view.innerHTML = '';
      viewCleanup = mountFleet(view);
    }
  }

  function mountTablet(el, name) {
    if (typeof NS.renderTablet === 'function') return NS.renderTablet(el, name, store) || null;
    el.innerHTML = `<div class="stub-panel">HOST ${esc(name)} — tablet module not loaded · <a href="#/fleet">← fleet</a></div>`;
    return null;
  }

  // ---- FLEET LIST view (§2.1) ----------------------------------------------------
  // UI state lives OUTSIDE the render path (§4.7): sort/filter/search/density/
  // keyboard cursor survive every data refresh and every remount.
  const fleetUI = {
    sort: localStorage.getItem('cradle2.sort') || 'name',
    filter: 'all',
    q: '',
    dense: localStorage.getItem('cradle2.dense') === '1',
    cursor: -1, // j/k cursor index into the visible row order
  };

  const SEV = { crashed: 0, restarting: 1, starting: 2, paused: 3, stopped: 4, running: 5 };
  // hp is omitempty on HostStatus: a live host at exactly 0 hp omits the key —
  // treat absent as 0 so triage sorts the most-critical host FIRST, not NaN.
  const hpFrac = h => (h.live && h.max_hp ? (h.hp || 0) / h.max_hp : 2); // not-live sorts last
  const fatOf = h => (typeof h.fatigue === 'number' ? h.fatigue : -1);
  const CMP = {
    name: (a, b) => a.name.localeCompare(b.name),
    hp: (a, b) => hpFrac(a) - hpFrac(b) || a.name.localeCompare(b.name), // lowest first: triage
    fatigue: (a, b) => fatOf(b) - fatOf(a) || a.name.localeCompare(b.name),
    restarts: (a, b) => (b.restarts || 0) - (a.restarts || 0) || a.name.localeCompare(b.name),
    status: (a, b) => (SEV[a.status] ?? 9) - (SEV[b.status] ?? 9) || a.name.localeCompare(b.name),
  };

  function fleetRows(hosts) {
    let rows = hosts.slice();
    if (fleetUI.filter !== 'all') rows = rows.filter(h => h.status === fleetUI.filter);
    if (fleetUI.q) {
      const q = fleetUI.q.toLowerCase();
      rows = rows.filter(h => (h.name + ' ' + (h.goal || '') + ' ' + (h.current_routine || '')).toLowerCase().includes(q));
    }
    rows.sort(CMP[fleetUI.sort] || CMP.name);
    return rows;
  }

  function mountFleet(root) {
    root.innerHTML = `
      <div id="fleet">
        <div class="toolbar">
          <label>sort <select class="f-sort">
            <option value="name">name</option><option value="hp">hp</option>
            <option value="fatigue">fatigue</option><option value="restarts">restarts</option>
            <option value="status">status</option>
          </select></label>
          <label>filter <select class="f-filter">
            <option value="all">all</option><option value="running">running</option>
            <option value="paused">paused</option><option value="starting">starting</option>
            <option value="stopped">stopped</option><option value="crashed">crashed</option>
            <option value="restarting">restarting</option>
          </select></label>
          <input type="search" class="f-q" placeholder="search name / goal / routine">
          <button class="f-dense" title="one-line dense mode">≡ dense</button>
          <span class="count"><span class="f-count">0</span> hosts</span>
        </div>
        <div class="group-head"><span class="f-group">▼ FLEET (0)</span><span class="f-group-roll"></span></div>
        <div class="list" tabindex="-1"></div>
        <div class="legend">▣stall ◔forage ⚔fight ✋talk zZsleep ·idle ✕down — nominal dim · click row → tablet · j/k navigate · Enter opens</div>
      </div>`;
    // One flat group until HostStatus grows a `community` field (gap G-15,
    // Stage 2) — then this becomes community-grouped headers with rollups.
    const list = root.querySelector('.list');
    const elSort = root.querySelector('.f-sort');
    const elFilter = root.querySelector('.f-filter');
    const elQ = root.querySelector('.f-q');
    const elDense = root.querySelector('.f-dense');
    elSort.value = fleetUI.sort;
    elFilter.value = fleetUI.filter;
    elQ.value = fleetUI.q;
    setClass(list, 'dense', fleetUI.dense);
    setClass(elDense, 'on', fleetUI.dense);

    let lastOrder = []; // visible row order (names), for j/k

    function createRow(h) {
      const el = document.createElement('div');
      el.className = 'row';
      el.dataset.name = h.name;
      el.innerHTML = `
        <div class="r1">
          <span class="dot"></span><span class="glyph"></span><span class="hname"></span>
          <span class="badge-a" hidden>ANLZ</span><span class="badge-r"></span>
          <span class="hp-num"></span><span class="fat-num"></span>
          <span class="oneliner"></span>
          <span class="btns">
            <button data-act="pause" title="pause">⏸</button><button data-act="resume" title="resume">▶</button><button data-act="restart" title="restart">↻</button><button data-act="stop" title="stop" class="danger">⏹</button>
          </span>
        </div>
        <div class="r2">
          <span class="loc"></span>
          <span class="spark-wrap">hp <canvas class="spark hp" width="48" height="11"></canvas></span>
          <span class="spark-wrap">fat <canvas class="spark fat" width="48" height="11"></canvas></span>
          <span class="errtxt"></span>
        </div>`;
      el.__r = {
        dot: el.querySelector('.dot'),
        glyph: el.querySelector('.glyph'),
        name: el.querySelector('.hname'),
        anlz: el.querySelector('.badge-a'),
        rst: el.querySelector('.badge-r'),
        hp: el.querySelector('.hp-num'),
        fat: el.querySelector('.fat-num'),
        ol: el.querySelector('.oneliner'),
        loc: el.querySelector('.loc'),
        sparkHP: el.querySelector('canvas.hp'),
        sparkFat: el.querySelector('canvas.fat'),
        err: el.querySelector('.errtxt'),
      };
      return el;
    }

    function updateRow(el, h) {
      const r = el.__r;
      r.dot.className = 'dot s-' + h.status;
      setText(r.glyph, NS.activityGlyph(h));
      setText(r.name, h.name);
      r.anlz.hidden = !h.analysis;
      setText(r.rst, h.restarts ? '↻' + h.restarts : '');
      setText(r.hp, h.live ? `${h.hp || 0}/${h.max_hp}` : '—');
      setText(r.fat, typeof h.fatigue === 'number' ? String(h.fatigue) : '—');
      const line = h.current_line ? ' L' + h.current_line : '';
      const ol = [h.goal, h.current_routine ? h.current_routine + line : '']
        .filter(Boolean).join(' · ') || (h.status === 'running' ? '(idle)' : '');
      setText(r.ol, ol);
      r.ol.title = ol;
      setText(r.loc, h.live ? `(${h.x},${h.y}) · fl ${NS.floorOf(h.y)}` : h.status);
      setText(r.err, h.err || '');
      const v = store.vitals(h.name) || { hp: [], fat: [] };
      drawSpark(r.sparkHP, v.hp, h.max_hp || 10);
      drawSpark(r.sparkFat, v.fat, 100);
      // display-by-exception: nominal = quietly running; anything else is lit
      setClass(el, 'nominal', h.status === 'running' && !h.err);
      setClass(el, 'sel', h.name === store.selected);
    }

    function applyCursor(scroll) {
      const rows = list.children;
      for (let i = 0; i < rows.length; i++) {
        setClass(rows[i], 'cursor', i === fleetUI.cursor);
        if (i === fleetUI.cursor && scroll) {
          // keyboard nav is explicit operator action — but still move only the
          // list's OWN scrollTop, never an ancestor (§4.2)
          const br = list.getBoundingClientRect(), rr = rows[i].getBoundingClientRect();
          if (rr.top < br.top) list.scrollTop += rr.top - br.top;
          else if (rr.bottom > br.bottom) list.scrollTop += rr.bottom - br.bottom;
        }
      }
    }

    function render() {
      const hosts = store.get('roster') || [];
      const rows = fleetRows(hosts);
      lastOrder = rows.map(h => h.name);
      if (fleetUI.cursor >= rows.length) fleetUI.cursor = rows.length - 1;
      setText(root.querySelector('.f-count'), String(hosts.length));
      setText(root.querySelector('.f-group'), `▼ FLEET (${hosts.length})`);
      const nominal = hosts.filter(h => h.status === 'running' && !h.err).length;
      setText(root.querySelector('.f-group-roll'),
        `${nominal} nominal · ${hosts.length - nominal} other`);
      renderKeyed(list, rows, h => h.name, createRow, updateRow);
      applyCursor(false);
    }

    const unsub = store.subscribe('roster', render);

    elSort.addEventListener('change', () => { fleetUI.sort = elSort.value; localStorage.setItem('cradle2.sort', fleetUI.sort); render(); });
    elFilter.addEventListener('change', () => { fleetUI.filter = elFilter.value; fleetUI.cursor = -1; render(); });
    elQ.addEventListener('input', () => { fleetUI.q = elQ.value.trim(); fleetUI.cursor = -1; render(); });
    elDense.addEventListener('click', () => {
      fleetUI.dense = !fleetUI.dense;
      localStorage.setItem('cradle2.dense', fleetUI.dense ? '1' : '0');
      setClass(list, 'dense', fleetUI.dense);
      setClass(elDense, 'on', fleetUI.dense);
    });

    list.addEventListener('click', e => {
      const btn = e.target.closest('button[data-act]');
      if (btn) {
        e.stopPropagation();
        const row = btn.closest('.row');
        if (row) store.lifecycle(row.dataset.name, btn.dataset.act);
        return;
      }
      const row = e.target.closest('.row');
      if (row) location.hash = '#/host/' + encodeURIComponent(row.dataset.name);
    });

    function onKey(e) {
      if (e.target && /^(INPUT|TEXTAREA|SELECT)$/.test(e.target.tagName)) return;
      if (e.key === 'j' || e.key === 'k') {
        e.preventDefault();
        if (!lastOrder.length) return;
        fleetUI.cursor = fleetUI.cursor < 0
          ? 0
          : Math.min(lastOrder.length - 1, Math.max(0, fleetUI.cursor + (e.key === 'j' ? 1 : -1)));
        applyCursor(true);
      } else if (e.key === 'Enter') {
        const name = lastOrder[fleetUI.cursor] || store.selected;
        if (name) location.hash = '#/host/' + encodeURIComponent(name);
      }
    }
    document.addEventListener('keydown', onKey);

    return () => {
      unsub();
      document.removeEventListener('keydown', onKey);
    };
  }

  // ---- boot ----------------------------------------------------------------------
  window.addEventListener('hashchange', route);
  // initial route waits for DOMContentLoaded so tablet.js (the next script tag)
  // has executed — a direct #/host/{n} load then mounts the real tablet.
  window.addEventListener('DOMContentLoaded', () => {
    wireChrome();
    route();
  });
})();
