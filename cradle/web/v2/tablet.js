// cradle v2 — tablet.js: THE HOST TABLET (#/host/{name}) — spec §1.2/§2.2.
// Mounted by app.js via window.CradleV2.renderTablet(el, name, store); the
// store contract (keys/cadences/shapes) is documented in app.js's header.
// Layout: header band / cognition tree + right feed column (+console) / bottom
// drawer. Every pane honors the live-region contract (§4): keyed in-place
// patches, auto-follow moves only a pane's OWN scrollTop (followPane/feedPin),
// manual scroll wins, fixed-height chrome bands. The accent is spent ONLY on
// the live execution path, exceptions (vetoes, crashed, fatigue≥90) and
// staleness. Optional endpoints (aspirations/fog/where/decisions/persona) may
// 404/501 while the backend lags — the store flags them hidden and the panes
// render "endpoint pending" (persona: the band is simply absent).
(function () {
  'use strict';
  const NS = (window.CradleV2 = window.CradleV2 || {});
  const { esc, setText, setClass, renderKeyed, followPane, feedPin } = NS;

  // ---- persistent UI state (§4.7: survives data refresh AND host switching) --
  const ui = {
    rightTab: localStorage.getItem('cradle2.tab.right') || 'thoughts',
    drawerTab: localStorage.getItem('cradle2.tab.drawer') || 'knowledge',
    drawerOpen: localStorage.getItem('cradle2.drawer.open') !== '0',
    consoleOpen: localStorage.getItem('cradle2.console.open') !== '0',
    consoleMode: 'eval',
    thoughts: { paused: false, kinds: { thought: true, note: true, veto: true, system: true, chat: true } },
    kn: { prov: { system: true, observed: true, deduced: true, hearsay: true }, kind: 'all', seen: 'all', q: '', sort: 'rank' },
    rel: { sort: 'closeness' },
    knExpanded: new Map(), // host -> Set(subject): expanded knowledge rows
    treeOpen: new Map(),   // host -> Set(parent id): expanded ✔-done chips
  };
  const perHostSet = (map, host) => {
    if (!map.has(host)) map.set(host, new Set());
    return map.get(host);
  };

  // ---- tiny formatters -------------------------------------------------------
  const BLOCKS = '▁▂▃▄▅▆▇█';
  const blk = v => BLOCKS[Math.max(0, Math.min(7, Math.round((v || 0) * 7)))];
  const pad2 = n => String(n).padStart(2, '0');
  const clock = at => {
    const d = new Date(at);
    return isNaN(d) ? '—' : pad2(d.getHours()) + ':' + pad2(d.getMinutes()) + ':' + pad2(d.getSeconds());
  };
  const when = at => { // clock today, MM-DD HH:MM when older — decisions span days
    const d = new Date(at);
    if (isNaN(d)) return '—';
    return (Date.now() - d.getTime() > 86400e3)
      ? pad2(d.getMonth() + 1) + '-' + pad2(d.getDate()) + ' ' + pad2(d.getHours()) + ':' + pad2(d.getMinutes())
      : clock(at);
  };
  // beliefs `at` / relationships `last_at` are unix seconds on the wire; accept
  // an RFC3339 string too (older servers) so timestamp math never goes NaN.
  const toUnix = v => (typeof v === 'string' ? (Date.parse(v) / 1000 || 0) : (v || 0));
  const age = at => { // unix seconds or RFC3339; 0/absent = unknown
    const unixSec = toUnix(at);
    if (!unixSec) return '—';
    const s = Math.max(0, Date.now() / 1000 - unixSec);
    if (s < 60) return Math.floor(s) + 's';
    if (s < 3600) return Math.floor(s / 60) + 'm';
    if (s < 86400) return Math.floor(s / 3600) + 'h';
    return Math.floor(s / 86400) + 'd';
  };
  const pct = v => ((v || 0) * 100).toFixed((v || 0) * 100 < 10 && v > 0 ? 1 : 0) + '%';
  const num1 = v => (Math.round((v || 0) * 10) / 10).toString();

  // setHTML mirrors setText for composed fragments: steady-state polls must not
  // blow away DOM the operator may be selecting text in (§4.1).
  const setHTML = (el, html) => {
    if (el.__html !== html) { el.__html = html; el.innerHTML = html; }
  };

  const GLYPH = { done: '✔', active: '●', open: '○', blocked: '▣', abandoned: '✕' };
  const statusGlyph = n => (n.kind === 'open_question' ? '?' : (GLYPH[n.status] || '○'));

  // forage-state chips parsed from open-question tags (spec §1.2 row 4)
  function forageChips(tags) {
    if (!tags || !tags.length) return '';
    let tried = 0, spent = 0, inflight = false, spinning = false;
    for (const t of tags) {
      if (t.startsWith('source-tried:')) tried++;
      else if (t.startsWith('source-spent:')) spent++;
      else if (t === 'forage-inflight') inflight = true;
      else if (t === 'spinning') spinning = true;
    }
    const out = [];
    if (tried) out.push('tried:' + tried);
    if (spent) out.push('spent:' + spent);
    if (inflight) out.push('foraging…');
    if (spinning) out.push('spinning');
    return out.join(' · ');
  }

  // group an event ring kind into a THOUGHTS-feed chip family
  const THOUGHT_GROUP = {
    agent_thought: 'thought', routine_note: 'note', policy_veto: 'veto', whisper: 'thought',
    system_message: 'system', chat_received: 'chat', other_player_chat: 'chat',
    private_message: 'chat', npc_chat: 'chat',
  };

  // ============================================================================
  // mount
  // ============================================================================
  NS.renderTablet = function (el, name, store) {
    const enc = encodeURIComponent(name);
    const dbg = p => `/api/hosts/${enc}/debug/${p}`;
    const unsubs = [];
    const sub = (key, fn) => unsubs.push(store.subscribe(key, fn));
    const knExpanded = perHostSet(ui.knExpanded, name);
    const treeOpen = perHostSet(ui.treeOpen, name);

    el.innerHTML = SKELETON;
    const $ = s => el.querySelector(s);
    const root = $('#tablet');

    // ---- header band ---------------------------------------------------------
    const hd = {
      dot: $('.t-dot'), name: $('.t-name'), status: $('.t-status'),
      hpBar: $('.bar.hp i'), hp: $('.t-hp'), fatBar: $('.bar.fat i'), fat: $('.t-fat'),
      cmb: $('.t-cmb'), stale: $('.t-hstale'),
      floor: $('.t-floor'), where: $('.t-where'), restarts: $('.t-restarts'),
      anlzBadge: $('.t-anlzbadge'), explored: $('.t-explored'),
      btns: $('.t-btns'), anlzBtn: $('.t-anlz'),
      starRow: $('.th-star'), star: $('.t-star'), personaRow: $('.th-persona'),
    };
    setText(hd.name, name);

    let host = null;
    sub('host', (h, m) => {
      hd.stale.hidden = !m.stale;
      if (!h) { setText(hd.status, m.stale ? 'unreachable' : '…'); return; }
      host = h;
      hd.dot.className = 't-dot dot s-' + h.status;
      setText(hd.status, h.status + (h.autonomous ? '·autonomous' : ''));
      if (h.live && h.max_hp) {
        const hp = h.hp || 0; // hp is omitempty on the wire: absent at exactly 0
        const f = hp / h.max_hp;
        hd.hpBar.style.width = (f * 100).toFixed(0) + '%';
        setClass(hd.hpBar, 'hot', f < 0.3); // low hp = exception
        setText(hd.hp, hp + '/' + h.max_hp);
      } else setText(hd.hp, '—');
      const fl = NS.floorOf(h.y);
      setText(hd.floor, h.live ? `floor ${fl}${fl > 0 ? ' (upstairs)' : ''} · (${h.x},${h.y})` : '');
      setText(hd.restarts, 'restarts ' + (h.restarts || 0));
      hd.anlzBadge.hidden = !h.analysis;
      setClass(hd.anlzBtn, 'on', !!h.analysis);
      renderCog(); // routine name/source live on HostStatus
    });
    sub('state', (s, m) => {
      if (!s) return;
      const fat = s.fatigue || 0;
      hd.fatBar.style.width = Math.min(100, fat) + '%';
      setClass(hd.fatBar, 'hot', fat >= 90); // the boxed-upstairs deadlock class
      setText(hd.fat, String(fat));
      setText(hd.cmb, 'cmb ' + (s.combat_level || 0));
      renderBody(s, m);
    });
    sub('where', (w, m) => {
      if (m.hidden) { setText(hd.where, ''); return; }
      if (!w) return;
      setText(hd.where, w.text ? '“' + w.text + '”' : '');
      hd.where.title = w.surroundings || '';
    });
    sub('fog', (f, m) => {
      if (m.hidden) { setText(hd.explored, 'explored —'); renderFog(null, m); return; }
      if (!f) return;
      const dir = f.frontiers && f.frontiers.length ? ' ➤' + f.frontiers[0].dir : '';
      setText(hd.explored, 'explored ' + pct(f.coverage && f.coverage.frac) + dir);
      renderFog(f, m);
    });
    sub('persona', (p, m) => {
      // pane absent (not "pending") when the proxy 404/501s — spec default.
      // A 200 with "persona":null (mesa has no document) or a cornerstone-less
      // document hides too: rendering it would fabricate an all-default band.
      const doc = p && (p.persona || p);
      if (m.hidden || !doc || !doc.cornerstone) { hd.starRow.hidden = true; hd.personaRow.hidden = true; return; }
      renderPersona(p);
    });

    hd.btns.addEventListener('click', e => {
      const b = e.target.closest('button');
      if (!b) return;
      if (b.dataset.act) { store.lifecycle(name, b.dataset.act); return; }
      if (b.classList.contains('t-anlz')) toggleAnalysis();
    });
    async function toggleAnalysis() {
      const action = host && host.analysis ? 'exit' : 'enter';
      try {
        const r = await store.fetchJSON(`/api/hosts/${enc}/analysis/${action}`, { method: 'POST' });
        conLine(`analysis ${action}: ${r && r.text ? r.text : (r && r.active ? 'active' : 'off')}`);
      } catch (e) { conLine('✕ analysis ' + action + ': ' + e.message, true); }
      store.refreshRoster();
    }

    // persona bar-matrix: HEXACO mu tick inside the band-width track, curiosity
    // flavors, decision dials, mood. Built once per fetch (persona is
    // fetched once per selection); innerHTML here is not a live pane.
    function renderPersona(resp) {
      const p = resp.persona || resp;
      const c = p.cornerstone || {};
      const star = c.identity && c.identity.north_star && c.identity.north_star.statement;
      hd.starRow.hidden = !star;
      if (star) setText(hd.star, star);

      const ORD = { very_low: 0, low: 1, mid: 2, mid_high: 3, high: 4, very_high: 5 };
      const dial = t => (t && t.mu > 0 && t.mu <= 1) ? t.mu
        : (t && ORD[t.band] != null ? (ORD[t.band] + 0.5) / 6 : 0.5);
      const trait = (letter, t) => {
        const o = t && ORD[t.band];
        const seg = o == null ? '' :
          `<span class="bandseg" style="left:${(o / 6 * 100).toFixed(1)}%;width:${(100 / 6).toFixed(1)}%"></span>`;
        const mu = `<span class="mu" style="left:${(dial(t) * 100).toFixed(1)}%"></span>`;
        return `<span class="trait" title="${esc(letter)} mu=${esc(t && t.mu)} band=${esc(t && t.band)}"><b>${esc(letter)}</b><span class="track">${seg}${mu}</span></span>`;
      };
      const hx = c.hexaco || {};
      let html = ['H', 'E', 'X', 'A', 'C', 'O'].map(l => trait(l, hx[l])).join('');
      const pr = c.prefs || {};
      const cu = pr.curiosity || {};
      html += `<span class="pgroup">curio soc${blk(cu.social)} sp${blk(cu.spatial)} sk${blk(cu.skill)} ec${blk(cu.economic)} rk${blk(cu.risk)}</span>`;
      const lam = pr.loss_aversion && pr.loss_aversion.mu;
      html += `<span class="pgroup">λ${lam ? num1(lam) : '—'} · patience ${blk(dial(pr.patience))} · aggr ${blk(dial(pr.aggression))} · decis ${blk(dial(pr.decisiveness))}</span>`;
      const mood = p.trajectory && p.trajectory.mood;
      if (mood && (mood.valence || mood.arousal || mood.stress || mood.confidence)) {
        html += `<span class="pgroup">mood v${mood.valence >= 0 ? '+' : ''}${num1(mood.valence)} a${num1(mood.arousal)} s${num1(mood.stress)} c${num1(mood.confidence)}</span>`;
      }
      setHTML(hd.personaRow, html);
      hd.personaRow.hidden = false;
    }

    // ---- COGNITION TREE (§2.2 left pane) --------------------------------------
    const cogPane = $('.t-cog');
    const cogBody = $('.cog-body');
    const cogEmpty = $('.cog-empty');
    const cogStale = $('.t-cstale');
    let mind = null, asps = null;
    sub('mind', (v, m) => {
      cogStale.hidden = !m.stale;
      setClass(cogPane, 'stale', !!m.stale);
      if (v) mind = v;
      renderCog();
      renderKnowledge();
      renderRelationships();
      renderQuestions();
    });
    sub('aspirations', (v, m) => { asps = m.hidden ? null : v; renderCog(); });

    // paced line pointer (store-replayed from line_trace/line_seq) — §4.2/4.3:
    // the highlight moves ONLY the routine box's own scrollTop via followPane.
    let curLine = 0;
    sub('routine_line', line => {
      if (typeof line !== 'number') return;
      curLine = line;
      pointRoutine();
    });

    function pointRoutine() {
      const keyed = cogBody.__keyed;
      const row = keyed && keyed.get('routine');
      if (!row) return;
      const box = row.querySelector('.crbox');
      setText(row.querySelector('.crpos'), `L${curLine}/${box.__total || '?'}`);
      const prev = box.querySelector('.ln.cur');
      if (prev && +prev.dataset.line === curLine) return;
      if (prev) prev.classList.remove('cur');
      const ln = box.querySelector(`.ln[data-line="${curLine}"]`);
      if (ln) {
        ln.classList.add('cur');
        if (box.__follow) box.__follow.follow(ln);
      }
    }

    // tree layout from goal_edges: serves/requires/enables/produces hang nodes,
    // with the HIGHER-kind side as parent (handles both authoring directions:
    // "smith requires bars" and the director's "bars requires smith" both put
    // the prerequisite under the goal); blocked_by renders ▣ + cause link.
    const KRANK = { aspiration: 5, goal: 4, open_goal: 3, subgoal: 2, open_question: 1, state: 1 };
    const HIER = { serves: 4, requires: 3, enables: 2, produces: 1 };

    function buildTree() {
      const nodes = new Map();
      for (const n of (mind.goal_nodes || [])) nodes.set(n.id, n);
      for (const n of (mind.open_questions || [])) if (!nodes.has(n.id)) nodes.set(n.id, n);
      const edges = mind.goal_edges || [];
      const parent = new Map(), children = new Map(), blocked = new Map();
      const hier = edges.filter(e => HIER[e.rel]).sort((a, b) => HIER[b.rel] - HIER[a.rel]);
      for (const e of hier) {
        let p = e.to, c = e.from;
        const rf = KRANK[(nodes.get(e.from) || {}).kind] || 0;
        const rt = KRANK[(nodes.get(e.to) || {}).kind] || 0;
        if (rf > rt) { p = e.from; c = e.to; }
        if (!nodes.has(p) || !nodes.has(c) || p === c) continue;
        if (nodes.get(c).kind === 'aspiration') continue; // aspirations are always roots
        if (parent.has(c)) continue;
        let anc = p, hop = 0, cyc = false;
        while (anc && hop++ < 64) { if (anc === c) { cyc = true; break; } anc = parent.get(anc); }
        if (cyc) continue;
        parent.set(c, p);
        if (!children.has(p)) children.set(p, []);
        children.get(p).push(c);
      }
      for (const e of edges) if (e.rel === 'blocked_by' && nodes.has(e.from)) blocked.set(e.from, e.to);
      return { nodes, parent, children, blocked, edges };
    }

    // requires/enables order siblings: a prerequisite sorts before its dependent
    function orderSiblings(ids, edges) {
      const before = (a, b) => edges.some(e =>
        (e.rel === 'enables' && e.from === a && e.to === b) ||
        (e.rel === 'requires' && e.from === b && e.to === a));
      const out = ids.slice();
      for (let pass = 0; pass < out.length; pass++) {
        let moved = false;
        for (let i = 1; i < out.length; i++) {
          if (before(out[i], out[i - 1])) { const t = out[i]; out[i] = out[i - 1]; out[i - 1] = t; moved = true; }
        }
        if (!moved) break;
      }
      return out;
    }

    function computeSpine(t) {
      const want = (host && host.goal || '').trim();
      let goalId = null;
      if (want) for (const n of t.nodes.values()) { if (n.id === want || n.label === want) { goalId = n.id; break; } }
      if (!goalId) for (const n of t.nodes.values()) { if (n.kind === 'goal' && n.status === 'active') { goalId = n.id; break; } }
      const spine = new Set();
      if (!goalId) return { spine, routineParent: null };
      let cur = goalId;
      while (cur && !spine.has(cur)) { spine.add(cur); cur = t.parent.get(cur); }
      let leaf = goalId;
      for (;;) { // descend the active chain to the deepest active node
        const kids = t.children.get(leaf) || [];
        const next = kids.find(k => { const n = t.nodes.get(k); return n && n.status === 'active' && !spine.has(k); });
        if (!next) break;
        spine.add(next);
        leaf = next;
      }
      return { spine, routineParent: leaf };
    }

    function aspStatusFor(node) {
      if (!asps) return null;
      const bare = node.id.replace(/^aspiration:/, '');
      return asps.find(a => a.id === node.id || a.id === bare || 'aspiration:' + a.id === node.id || a.label === node.label) || null;
    }

    function renderCog() {
      if (!mind) return;
      const t = buildTree();
      const { spine, routineParent } = computeSpine(t);
      const hasRoutine = !!(host && host.current_routine_source &&
        (host.status === 'running' || host.status === 'paused'));
      const rows = [];
      const seen = new Set();

      const emit = (id, depth) => {
        if (seen.has(id)) return;
        seen.add(id);
        const n = t.nodes.get(id);
        if (!n) return;
        rows.push(nodeRow(n, depth, t, spine));
        const kids = orderSiblings(t.children.get(id) || [], t.edges);
        const live = [], term = [];
        for (const k of kids) {
          const kn = t.nodes.get(k);
          ((kn && (kn.status === 'done' || kn.status === 'abandoned')) ? term : live).push(k);
        }
        for (const k of live) emit(k, depth + 1);
        if (term.length) { // §6.7 default: done/abandoned collapse to a chip
          const done = term.filter(k => (t.nodes.get(k) || {}).status === 'done').length;
          const open = treeOpen.has(id);
          rows.push({
            key: 'chip:' + id, type: 'chip', depth: depth + 1, parent: id,
            label: (open ? '▾ ' : '') + (done ? '✔ ' + done + ' done' : '') +
              (done && term.length - done ? ' · ' : '') +
              (term.length - done ? '✕ ' + (term.length - done) + ' abandoned' : ''),
          });
          if (open) for (const k of term) emit(k, depth + 1);
        }
        if (hasRoutine && id === routineParent) rows.push(routineRow(depth + 1, true));
      };

      // roots: aspirations first, then unparented goal-ish nodes (active first)
      const roots = [];
      for (const n of t.nodes.values()) if (n.kind === 'aspiration') roots.push(n.id);
      const rest = [];
      for (const n of t.nodes.values()) {
        if (n.kind !== 'aspiration' && !t.parent.has(n.id)) rest.push(n.id);
      }
      rest.sort((a, b) => {
        const sa = (t.nodes.get(a) || {}).status === 'active' ? 0 : 1;
        const sb = (t.nodes.get(b) || {}).status === 'active' ? 0 : 1;
        return sa - sb;
      });
      for (const id of roots.concat(rest)) emit(id, 0);
      if (hasRoutine && (!routineParent || !seen.has(routineParent))) rows.push(routineRow(0, false));

      cogEmpty.hidden = rows.length > 0;
      renderKeyed(cogBody, rows, r => r.key, createCogRow, updateCogRow);
      pointRoutine();
    }

    function nodeRow(n, depth, t, spine) {
      const isAsp = n.kind === 'aspiration';
      const row = {
        key: 'n:' + n.id, type: isAsp ? 'asp' : 'node', depth, node: n,
        spine: spine.has(n.id), blockedBy: t.blocked.get(n.id) || (n.status === 'blocked' ? '' : null),
        chips: n.kind === 'open_question' ? forageChips(n.tags) : '',
      };
      if (isAsp) {
        const st = aspStatusFor(n);
        let done = 0, active = 0, open = 0;
        if (st) { done = st.goals_done; active = st.goals_active; open = st.goals_open; }
        else { // derive the rollup from serving children until G-1 data arrives
          for (const k of (t.children.get(n.id) || [])) {
            const kn = t.nodes.get(k);
            if (!kn || (kn.kind !== 'goal' && kn.kind !== 'open_goal' && kn.kind !== 'subgoal')) continue;
            if (kn.status === 'done') done++;
            else if (kn.status === 'active' || kn.status === 'blocked') active++;
            else if (kn.status !== 'abandoned') open++;
          }
        }
        row.rollup = { done, active, open, neglected: !!(st && st.neglected) };
      }
      return row;
    }

    function routineRow(depth, attached) {
      return {
        key: 'routine', type: 'routine', depth, attached,
        name: (host && host.current_routine) || '',
        src: (host && host.current_routine_source) || '',
        spine: true,
      };
    }

    function createCogRow(r) {
      const div = document.createElement('div');
      if (r.type === 'chip') {
        div.className = 'cchip';
        div.addEventListener('click', () => {
          if (treeOpen.has(r.parent)) treeOpen.delete(r.parent);
          else treeOpen.add(r.parent);
          renderCog();
        });
        return div;
      }
      if (r.type === 'routine') {
        div.className = 'croutine';
        div.innerHTML = `
          <div class="crhead"><span class="crmark">└═</span>ROUTINE <span class="crname"></span> <span class="crpos"></span></div>
          <div class="crbox"></div>`;
        const box = div.querySelector('.crbox');
        box.__follow = followPane(box); // §4.3: own scrollTop, 4s manual yield
        return div;
      }
      div.className = r.type === 'asp' ? 'cnode casp' : 'cnode';
      div.innerHTML = `
        <span class="cglyph"></span><span class="clabel"></span>
        <span class="croll" hidden><span class="cbar"><i></i></span><span class="crolltxt"></span><span class="cneg" hidden>⚑NEGLECTED</span></span>
        <span class="cprog" hidden><span class="cbar"><i></i></span><span class="cpct"></span></span>
        <span class="cblk" hidden>▣</span><span class="cchips"></span>`;
      div.querySelector('.cblk').addEventListener('click', () => {
        const cause = div.dataset.cause;
        const target = cause && cogBody.__keyed && cogBody.__keyed.get('n:' + cause);
        if (!target) return; // explicit operator action — own-pane scroll only
        cogPane.scrollTop = Math.max(0, target.offsetTop - cogPane.clientHeight / 3);
        target.classList.remove('flash');
        void target.offsetWidth; // restart the flash animation
        target.classList.add('flash');
      });
      return div;
    }

    function updateCogRow(el, r) {
      el.style.paddingLeft = (r.depth * 16) + 'px';
      if (r.type === 'chip') { setText(el, r.label); return; }
      if (r.type === 'routine') {
        setClass(el, 'spine', true);
        setText(el.querySelector('.crname'), r.name);
        const box = el.querySelector('.crbox');
        if (box.__src !== r.src) { // rebuild source ONLY when the routine changes
          box.__src = r.src;
          const lines = r.src.replace(/\n$/, '').split('\n');
          box.__total = lines.length;
          box.innerHTML = '<pre>' + lines.map((ln, i) =>
            `<div class="ln" data-line="${i + 1}"><span class="num">${i + 1}</span><span class="lsrc">${esc(ln) || ' '}</span></div>`
          ).join('') + '</pre>';
          pointRoutine();
        }
        return;
      }
      const n = r.node;
      setClass(el, 'spine', !!r.spine);
      setClass(el, 'terminal', n.status === 'done' || n.status === 'abandoned');
      const g = el.querySelector('.cglyph');
      setText(g, r.type === 'asp' ? '◇' : statusGlyph(n));
      const lab = el.querySelector('.clabel');
      const text = r.type === 'asp' ? n.label.replace(/^aspiration:/, '').replace(/-/g, ' ') : (n.label || n.id);
      setText(lab, text);
      lab.title = n.id + ' [' + n.kind + ' · ' + n.status + ']';
      if (r.type === 'asp') {
        const roll = el.querySelector('.croll');
        roll.hidden = false;
        const { done, active, open, neglected } = r.rollup;
        const total = done + active + open;
        roll.querySelector('.cbar i').style.width = total ? (done / total * 100).toFixed(0) + '%' : '0%';
        setText(roll.querySelector('.crolltxt'), `${done}✔ ${active}● ${open}○`);
        roll.querySelector('.cneg').hidden = !neglected;
      } else {
        const prog = el.querySelector('.cprog');
        const show = typeof n.progress === 'number' && n.progress > 0;
        prog.hidden = !show;
        if (show) {
          prog.querySelector('.cbar i').style.width = (n.progress * 100).toFixed(0) + '%';
          setText(prog.querySelector('.cpct'), (n.progress * 100).toFixed(0) + '%');
        }
      }
      const blkEl = el.querySelector('.cblk');
      blkEl.hidden = r.blockedBy == null;
      if (r.blockedBy != null) {
        el.dataset.cause = r.blockedBy;
        blkEl.title = r.blockedBy ? 'blocked by: ' + r.blockedBy + ' (click to locate)' : 'blocked';
        setText(blkEl, r.blockedBy ? '▣ blocked → ' + r.blockedBy.slice(0, 24) : '▣ blocked');
      }
      setText(el.querySelector('.cchips'), r.chips || '');
    }

    // ---- right column: THOUGHTS / DECISIONS tabs ------------------------------
    const rtTabs = $('.rt-tabs');
    const thPane = $('.th-pane'), decPane = $('.dec-pane');
    const feedEl = $('.th-feed'), pill = $('.th-pill');
    const thTools = $('.th-tools');
    const pin = feedPin(feedEl, pill);
    const tRows = [];
    let lastSeq = 0, dropped = false;

    function applyRightTab() {
      for (const b of rtTabs.querySelectorAll('button[data-tab]')) setClass(b, 'on', b.dataset.tab === ui.rightTab);
      thPane.hidden = ui.rightTab !== 'thoughts';
      decPane.hidden = ui.rightTab !== 'decisions';
      thTools.hidden = ui.rightTab !== 'thoughts';
    }
    rtTabs.addEventListener('click', e => {
      const b = e.target.closest('button[data-tab]');
      if (!b) return;
      ui.rightTab = b.dataset.tab;
      localStorage.setItem('cradle2.tab.right', ui.rightTab);
      applyRightTab();
    });

    // kind-filter chips + pause
    const pauseBtn = $('.th-pause');
    thTools.addEventListener('click', e => {
      const b = e.target.closest('button');
      if (!b) return;
      if (b === pauseBtn) {
        ui.thoughts.paused = !ui.thoughts.paused;
        setClass(pauseBtn, 'on', ui.thoughts.paused);
        if (!ui.thoughts.paused) renderThoughts();
        return;
      }
      if (b.dataset.k) {
        ui.thoughts.kinds[b.dataset.k] = !ui.thoughts.kinds[b.dataset.k];
        setClass(b, 'on', ui.thoughts.kinds[b.dataset.k]);
        renderThoughts();
      }
    });
    for (const b of thTools.querySelectorAll('button[data-k]')) setClass(b, 'on', !!ui.thoughts.kinds[b.dataset.k]);
    setClass(pauseBtn, 'on', ui.thoughts.paused);

    sub('events', (evs, m) => {
      setClass(thPane, 'stale', !!m.stale);
      if (!evs) return;
      let added = 0;
      for (const e of evs) {
        if (e.seq <= lastSeq) continue;
        lastSeq = e.seq;
        const g = THOUGHT_GROUP[e.kind];
        if (!g) continue;
        tRows.push({ seq: e.seq, at: e.at, group: g, kind: e.kind, data: e.data || {} });
        added++;
      }
      if (tRows.length > 200) { tRows.splice(0, tRows.length - 200); dropped = true; }
      if (added || feedEl.childElementCount === 0) renderThoughts();
    });

    let renderedSeq = 0; // newest row seq the feed has actually painted
    function renderThoughts() {
      if (ui.thoughts.paused) return;
      const vis = tRows.filter(r => ui.thoughts.kinds[r.group]);
      const items = dropped ? [{ seq: '_cap' }].concat(vis) : vis;
      $('.th-empty').hidden = vis.length > 0;
      const was = pin.atBottom();
      renderKeyed(feedEl, items, r => r.seq, createThought, () => {});
      // §4.4: count appends from DATA, not a DOM-size delta — at the 200-row
      // cap every batch evicts as many rows as it appends, so childElementCount
      // never grows and the tail-pin + '▼ N new' pill would silently die.
      let grew = 0;
      for (let i = vis.length - 1; i >= 0 && vis[i].seq > renderedSeq; i--) grew++;
      if (vis.length) renderedSeq = Math.max(renderedSeq, vis[vis.length - 1].seq);
      if (grew > 0) pin.appended(grew, was);
    }

    function createThought(r) {
      const div = document.createElement('div');
      if (r.seq === '_cap') {
        div.className = 'tcap';
        div.textContent = '— 200-row cap · older → Archive —';
        return div;
      }
      div.className = 'trow g-' + r.group;
      const d = r.data, t = clock(r.at);
      let html = '';
      if (r.kind === 'agent_thought') {
        html = `<div class="t1"><span class="tt">${esc(t)}</span> turn ${esc(d.turn)} · ${esc(d.trigger || 'tick')}</div>` +
          `<div class="t2">${esc(d.reasoning || '')}</div>` +
          (d.move_kind || d.dsl ? `<div class="t3">move: ${esc(d.move_kind || '—')}${d.dsl ? '  dsl: ' + esc(d.dsl) : ''}</div>` : '');
      } else if (r.kind === 'routine_note') {
        html = `<div class="t1"><span class="tt">${esc(t)}</span> ⚑ <span class="t2i">${esc(d.Text || d.text || '')}</span></div>`;
      } else if (r.kind === 'policy_veto') {
        html = `<div class="t1"><span class="tt">${esc(t)}</span> ⚐ veto ${esc(d.action || '')} — rule ${esc(d.rule || '')}</div>` +
          (d.reason ? `<div class="t2">${esc(d.reason)}</div>` : '');
      } else if (r.kind === 'system_message') {
        html = `<div class="t1"><span class="tt">${esc(t)}</span> ⚙ <span class="t2i">${esc(d.Message || d.message || '')}</span></div>`;
      } else { // chat family
        const who = d.Speaker || d.Sender || '';
        const msg = d.Message || d.MessageText || d.message || '';
        html = `<div class="t1"><span class="tt">${esc(t)}</span> ♪ ${who ? esc(who) + ': ' : ''}<span class="t2i">${esc(msg)}</span></div>`;
      }
      div.innerHTML = html;
      return div;
    }

    // DECISIONS: durable tail (debug/decisions, gap G-4), newest first
    const decList = $('.dec-list');
    sub('decisions', (v, m) => {
      setClass(decPane, 'stale', !!m.stale);
      $('.dec-pending').hidden = !m.hidden;
      if (m.hidden) { $('.dec-empty').hidden = true; return; }
      const recs = Array.isArray(v) ? v.slice().reverse() : [];
      $('.dec-empty').hidden = recs.length > 0;
      // identical decisions in the same second (the retry-spin this pane
      // exists to expose) must not collapse into one row: disambiguate
      // repeats with an occurrence counter, keeping distinct rows' keys
      // stable across polls (an index-based key would rebuild the whole list).
      const occ = new Map();
      const decKey = r => {
        const base = r.at + '|' + (r.kind || '') + '|' + (r.reasoning || '').slice(0, 24);
        const n = occ.get(base) || 0;
        occ.set(base, n + 1);
        return n ? base + '|' + n : base;
      };
      renderKeyed(decList, recs, decKey, () => {
        const div = document.createElement('div');
        div.className = 'drow';
        div.innerHTML = `<div class="d1"><span class="tt"></span><span class="dkind"></span><span class="dtrig"></span></div>
          <div class="dgoal"></div><div class="dreason"></div>`;
        return div;
      }, (el, r) => {
        setText(el.querySelector('.tt'), when(r.at));
        setText(el.querySelector('.dkind'), r.kind || '');
        setText(el.querySelector('.dtrig'), r.trigger || '');
        const goal = el.querySelector('.dgoal');
        setText(goal, r.goal ? 'goal: ' + r.goal : '');
        goal.hidden = !r.goal;
        setText(el.querySelector('.dreason'), r.reasoning || '');
      });
    });

    // ---- CONSOLE drawer (eval / script / analysis directive / whisper) --------
    const conOut = $('.con-out'), conIn = $('.con-in'), conBody = $('.con-body');
    const conUrg = $('.con-urg'); // whisper urgency select — fixed chrome, whisper mode only
    const PLACEHOLDER = {
      eval: 'eval one DSL line, e.g. say("hi") — Enter to run',
      script: 'paste a .routine — Ctrl+Enter to run',
      analysis: 'operator directive — e.g. "go to the bank", "?where are you" — Enter to send',
      whisper: 'whisper a thought into the host’s head, e.g. "the bank is north of you" — Enter to send',
    };
    function applyConsole() {
      conBody.hidden = !ui.consoleOpen;
      setText($('.con-toggle'), ui.consoleOpen ? 'CONSOLE ▾' : 'CONSOLE ▸');
      for (const b of el.querySelectorAll('.con-h button.cm')) setClass(b, 'on', b.dataset.m === ui.consoleMode);
      conUrg.hidden = ui.consoleMode !== 'whisper';
      conIn.placeholder = PLACEHOLDER[ui.consoleMode];
      conIn.rows = ui.consoleMode === 'script' ? 5 : 1;
    }
    $('.con-toggle').addEventListener('click', () => {
      ui.consoleOpen = !ui.consoleOpen;
      localStorage.setItem('cradle2.console.open', ui.consoleOpen ? '1' : '0');
      applyConsole();
    });
    $('.con-h').addEventListener('click', e => {
      const b = e.target.closest('button.cm');
      if (!b) return;
      ui.consoleMode = b.dataset.m;
      applyConsole();
    });
    conIn.addEventListener('keydown', e => {
      const submit = ui.consoleMode === 'script'
        ? (e.key === 'Enter' && (e.ctrlKey || e.metaKey))
        : (e.key === 'Enter' && !e.shiftKey);
      if (submit) { e.preventDefault(); runConsole(); }
    });
    $('.con-run').addEventListener('click', runConsole);

    function conLine(text, isErr) {
      const div = document.createElement('div');
      div.className = 'con-line' + (isErr ? ' err' : '');
      div.textContent = text;
      conOut.appendChild(div);
      while (conOut.childElementCount > 60) conOut.firstElementChild.remove();
      conOut.scrollTop = conOut.scrollHeight; // own pane, response to user action
    }
    async function runConsole() {
      const text = conIn.value.trim();
      if (!text) return;
      conLine('> ' + text);
      conIn.value = '';
      try {
        let r;
        if (ui.consoleMode === 'eval') r = await store.fetchJSON(dbg('eval'), { method: 'POST', body: text });
        else if (ui.consoleMode === 'script') r = await store.fetchJSON(dbg('script'), { method: 'POST', body: text });
        else if (ui.consoleMode === 'whisper') r = await store.fetchJSON(dbg('whisper'), { method: 'POST', body: JSON.stringify({ text, urgency: conUrg.value }) });
        else r = await store.fetchJSON(`/api/hosts/${enc}/analysis/directive`, { method: 'POST', body: text });
        if (ui.consoleMode === 'whisper') {
          if (r && r.ok) conLine(`💭 whispered (${r.urgency || conUrg.value}) · ${r.queued} queued — surfaces next director turn`);
          else conLine('✕ ' + ((r && r.error) || 'whisper failed'), true);
        } else if (ui.consoleMode === 'analysis') {
          conLine(`[${r.kind || 'verdict'}${r.executed ? ' · executed' : ''}${r.active ? ' · analysis ON' : ''}] ${r.text || ''}${r.dsl ? '  dsl: ' + r.dsl : ''}`, !!r.error);
          if (r.error) conLine('✕ ' + r.error, true);
        } else if (r.ok === false || r.error) {
          conLine('✕ ' + (r.error || 'failed'), true);
        } else {
          conLine('= ' + (r.value === undefined || r.value === null ? 'ok' : JSON.stringify(r.value)) + (r.kind ? '  (' + r.kind + ')' : ''));
        }
      } catch (e) { conLine('✕ ' + e.message, true); }
    }
    applyConsole();
    applyRightTab();

    // ---- bottom drawer --------------------------------------------------------
    const drTabs = $('.dr-tabs'), drBody = $('.dr-body');
    function applyDrawer() {
      for (const b of drTabs.querySelectorAll('button[data-tab]')) setClass(b, 'on', b.dataset.tab === ui.drawerTab);
      drBody.hidden = !ui.drawerOpen;
      setClass($('#t-drawer'), 'closed', !ui.drawerOpen);
      for (const p of drBody.querySelectorAll('.dr-pane')) p.hidden = !p.classList.contains('p-' + ui.drawerTab);
    }
    drTabs.addEventListener('click', e => {
      const b = e.target.closest('button[data-tab]');
      if (!b) return;
      if (b.dataset.tab === ui.drawerTab) ui.drawerOpen = !ui.drawerOpen; // re-click toggles
      else { ui.drawerTab = b.dataset.tab; ui.drawerOpen = true; }
      localStorage.setItem('cradle2.tab.drawer', ui.drawerTab);
      localStorage.setItem('cradle2.drawer.open', ui.drawerOpen ? '1' : '0');
      applyDrawer();
    });
    applyDrawer();

    // -- KNOWLEDGE table (filters + search + sort + row-expand) -----------------
    const knToolbar = $('.kn-toolbar'), knList = $('.kn-list');
    const knKind = $('.kn-kind'), knSeen = $('.kn-seen'), knQ = $('.kn-q'), knSort = $('.kn-sort');
    knKind.value = ui.kn.kind; knSeen.value = ui.kn.seen; knQ.value = ui.kn.q; knSort.value = ui.kn.sort;
    for (const b of knToolbar.querySelectorAll('button[data-p]')) setClass(b, 'on', !!ui.kn.prov[b.dataset.p]);
    knToolbar.addEventListener('click', e => {
      const b = e.target.closest('button[data-p]');
      if (!b) return;
      ui.kn.prov[b.dataset.p] = !ui.kn.prov[b.dataset.p];
      setClass(b, 'on', ui.kn.prov[b.dataset.p]);
      renderKnowledge();
    });
    knKind.addEventListener('change', () => { ui.kn.kind = knKind.value; renderKnowledge(); });
    knSeen.addEventListener('change', () => { ui.kn.seen = knSeen.value; renderKnowledge(); });
    knQ.addEventListener('input', () => { ui.kn.q = knQ.value.trim().toLowerCase(); renderKnowledge(); });
    knSort.addEventListener('change', () => { ui.kn.sort = knSort.value; renderKnowledge(); });

    const factUpdated = f => Array.isArray(f.beliefs) ? f.beliefs.reduce((m, b) => Math.max(m, toUnix(b.at)), 0) : 0;
    const beliefCount = f => Array.isArray(f.beliefs) ? f.beliefs.length : (f.beliefs || 0);
    const SEEN_S = { '10m': 600, '1h': 3600, '24h': 86400 };

    function renderKnowledge() {
      if (!mind) return;
      const all = mind.knowledge || [];
      setText($('.cnt-k'), String(all.length));
      let rows = all.map((f, i) => ({ f, i }));
      rows = rows.filter(({ f }) => {
        if (f.provenance && ui.kn.prov[f.provenance] === false) return false;
        if (ui.kn.kind !== 'all' && f.kind !== ui.kn.kind) return false;
        if (ui.kn.seen !== 'all') { // recency needs belief timestamps (?full=1)
          const u = factUpdated(f);
          if (u && Date.now() / 1000 - u > SEEN_S[ui.kn.seen]) return false;
        }
        if (ui.kn.q && !((f.subject + ' ' + (f.top_claim || '')).toLowerCase().includes(ui.kn.q))) return false;
        return true;
      });
      const cmp = {
        rank: (a, b) => a.i - b.i,
        confidence: (a, b) => (b.f.confidence || 0) - (a.f.confidence || 0),
        subject: (a, b) => a.f.subject.localeCompare(b.f.subject),
        familiar: (a, b) => (b.f.familiar || 0) - (a.f.familiar || 0),
        beliefs: (a, b) => beliefCount(b.f) - beliefCount(a.f),
        updated: (a, b) => factUpdated(b.f) - factUpdated(a.f),
      }[ui.kn.sort] || ((a, b) => a.i - b.i);
      rows.sort(cmp);

      // keep the kind <select> options synced to the kinds actually present
      const kinds = [...new Set(all.map(f => f.kind).filter(Boolean))].sort();
      const sig = kinds.join(',');
      if (knKind.__sig !== sig) {
        knKind.__sig = sig;
        knKind.innerHTML = '<option value="all">all</option>' + kinds.map(k => `<option>${esc(k)}</option>`).join('');
        knKind.value = kinds.includes(ui.kn.kind) ? ui.kn.kind : (ui.kn.kind = 'all');
      }
      setText($('.kn-n'), rows.length + ' / ' + all.length + ' subjects');
      $('.kn-empty').hidden = rows.length > 0;

      renderKeyed(knList, rows, r => r.f.subject, () => {
        const div = document.createElement('div');
        div.className = 'krow';
        div.innerHTML = `
          <div class="khead"><span class="kx"></span><span class="ksub"></span><span class="kkind"></span>
            <span class="kclaim"></span><span class="kconf"></span><span class="kprov"></span>
            <span class="kab"></span><span class="kblf"></span><span class="kupd"></span></div>
          <div class="kdetails" hidden></div>`;
        div.querySelector('.khead').addEventListener('click', () => {
          const s = div.__subject;
          if (knExpanded.has(s)) knExpanded.delete(s); else knExpanded.add(s);
          renderKnowledge();
        });
        return div;
      }, (elr, { f }) => {
        elr.__subject = f.subject;
        const open = knExpanded.has(f.subject);
        setText(elr.querySelector('.kx'), open ? '▾' : '▸');
        setText(elr.querySelector('.ksub'), f.subject);
        setText(elr.querySelector('.kkind'), f.kind || '');
        const claim = elr.querySelector('.kclaim');
        setText(claim, f.top_claim || '—');
        claim.title = f.top_claim || '';
        setText(elr.querySelector('.kconf'), (f.confidence || 0).toFixed(2));
        setText(elr.querySelector('.kprov'), f.provenance || '—');
        const top = Array.isArray(f.beliefs) && f.beliefs[0];
        setText(elr.querySelector('.kab'), top ? num1(top.alpha) + '/' + num1(top.beta) : '—');
        setText(elr.querySelector('.kblf'), String(beliefCount(f)));
        setText(elr.querySelector('.kupd'), age(factUpdated(f)));
        const det = elr.querySelector('.kdetails');
        det.hidden = !open;
        if (open) {
          setHTML(det, Array.isArray(f.beliefs)
            ? f.beliefs.map(b => {
                const conf = (b.alpha + b.beta) > 0 ? (b.alpha / (b.alpha + b.beta)).toFixed(2) : '—';
                return `<div class="kbel">• “${esc(b.claim)}”  <span class="kprov">${esc(b.provenance)}</span>  α${esc(num1(b.alpha))} β${esc(num1(b.beta))}  ${esc(conf)}  ${esc(age(b.at))} ago</div>`;
              }).join('') || '<div class="kbel dim">no graded beliefs (familiarity only)</div>'
            : '<div class="kbel dim">full belief detail needs /mind?full=1 (G-3) — endpoint pending</div>');
        }
      });
    }

    // -- RELATIONSHIPS ledger ---------------------------------------------------
    const relSort = $('.rel-sort'), relList = $('.rel-list');
    relSort.value = ui.rel.sort;
    relSort.addEventListener('change', () => { ui.rel.sort = relSort.value; renderRelationships(); });

    function renderRelationships() {
      if (!mind) return;
      const all = (mind.relationships || []).slice();
      setText($('.cnt-r'), String(all.length));
      $('.rel-empty').hidden = all.length > 0;
      const closeness = r => (r.trust || 0) + Math.max(0, r.affinity || 0) - (r.grievance || 0);
      const cmp = {
        closeness: (a, b) => closeness(b) - closeness(a),
        trust: (a, b) => (b.trust || 0) - (a.trust || 0),
        grievance: (a, b) => (b.grievance || 0) - (a.grievance || 0),
        recent: (a, b) => toUnix(b.last_at) - toUnix(a.last_at),
        familiar: (a, b) => (b.familiar || 0) - (a.familiar || 0),
      }[ui.rel.sort];
      all.sort(cmp);
      renderKeyed(relList, all, r => r.name, () => {
        const div = document.createElement('div');
        div.className = 'rrow';
        div.innerHTML = `
          <span class="rname"></span><span class="rgrade"></span>
          <span class="rbarw">t <span class="bar t"><i></i></span> <span class="rtv"></span></span>
          <span class="rbarw">a <span class="bar a"><i></i></span> <span class="rav"></span></span>
          <span class="rbarw">g <span class="bar g"><i></i></span> <span class="rgv"></span></span>
          <span class="rint"></span><span class="rlast"></span><span class="rfam"></span><span class="rtags"></span>`;
        return div;
      }, (elr, r) => {
        setText(elr.querySelector('.rname'), r.name);
        setText(elr.querySelector('.rgrade'), r.grade || '');
        elr.querySelector('.bar.t i').style.width = (Math.max(0, Math.min(1, r.trust || 0)) * 100).toFixed(0) + '%';
        setText(elr.querySelector('.rtv'), (r.trust || 0).toFixed(2));
        const aBar = elr.querySelector('.bar.a i');
        aBar.style.width = (Math.min(1, Math.abs(r.affinity || 0)) * 100).toFixed(0) + '%';
        setClass(aBar, 'neg', (r.affinity || 0) < 0);
        setText(elr.querySelector('.rav'), ((r.affinity || 0) >= 0 ? '+' : '') + (r.affinity || 0).toFixed(2));
        const gBar = elr.querySelector('.bar.g i');
        gBar.style.width = (Math.max(0, Math.min(1, r.grievance || 0)) * 100).toFixed(0) + '%';
        setClass(gBar, 'hot', (r.grievance || 0) > 0.4);
        setText(elr.querySelector('.rgv'), (r.grievance || 0).toFixed(2));
        setText(elr.querySelector('.rint'), 'int ' + (r.interactions != null ? r.interactions : '—'));
        setText(elr.querySelector('.rlast'), r.last_at ? age(r.last_at) : '—');
        setText(elr.querySelector('.rfam'), 'fam ' + (r.familiar || 0));
        setText(elr.querySelector('.rtags'), (r.tags || []).join(' '));
      });
    }

    // -- EXPLORATION (numeric/textual — spec §2.2, no canvas) -------------------
    function renderFog(f, m) {
      const pane = $('.p-exploration');
      setClass(pane, 'stale', !!(m && m.stale));
      $('.fog-pending').hidden = !(m && m.hidden);
      $('.fog-data').hidden = !!(m && m.hidden) || !f;
      if (!f) return;
      const cov = f.coverage || {};
      // seen/total count SECTORS (48×48-tile units), not cells; frac is the
      // cell-weighted fraction, so the two legitimately disagree.
      setText($('.fog-cov'), `coverage ${pct(cov.frac)}  (saw ${(cov.seen || 0).toLocaleString()} / ${(cov.total || 0).toLocaleString()} sectors) · known POIs ${f.known_pois || 0}`);
      const here = f.here || {};
      setText($('.fog-here'), `here: terrain ${pct(here.terrain)} · contents ${pct(here.contents)}`);
      const fr = f.frontiers || [];
      renderKeyed($('.fog-frontiers'), fr, x => x.dir, () => {
        const d = document.createElement('div');
        d.className = 'fog-fr';
        return d;
      }, (elr, x) => setText(elr, `➤ ${x.dir}  ${x.dist} tiles  (${x.x},${x.y})`));
      $('.fog-fr-h').hidden = !fr.length;
      const adj = (f.adjacent || []).map(a => `${a.dir} ${pct(a.terrain)}/${pct(a.contents)}`).join(' · ');
      setText($('.fog-adj'), adj ? 'adjacent: ' + adj : '');
    }

    // -- BODY: skills / inventory / equipment / ground / dialog / server msgs ---
    function renderBody(s, m) {
      const pane = $('.p-body');
      setClass(pane, 'stale', !!(m && m.stale));
      renderKeyed($('.bd-skills'), s.skills || [], x => x.name, () => {
        const d = document.createElement('div');
        d.className = 'bd-skill';
        d.innerHTML = '<b></b><span class="lv"></span><span class="xp"></span>';
        return d;
      }, (elr, x) => {
        setText(elr.querySelector('b'), x.name);
        setText(elr.querySelector('.lv'), x.cur + '/' + x.level);
        setClass(elr.querySelector('.lv'), 'boost', x.cur !== x.level);
        setText(elr.querySelector('.xp'), x.xp + 'xp');
      });

      setText($('.bd-inv-h'), `INVENTORY ${s.inventory_used || 0}/30 (${s.inventory_free || 0} free)`);
      const worn = new Set((s.equipment || []).map(q => q.name));
      renderKeyed($('.bd-inv'), s.inventory || [], x => x.slot, () => {
        const d = document.createElement('div');
        d.className = 'bd-item';
        return d;
      }, (elr, x) => {
        setText(elr, x.name + (x.amount > 1 ? ' ×' + x.amount : '') + (worn.has(x.name) ? ' ⛨' : ''));
        setClass(elr, 'worn', worn.has(x.name));
      });
      $('.bd-inv-empty').hidden = (s.inventory || []).length > 0;

      const eqWrap = $('.bd-eq');
      if (s.equipment === undefined) { // pre-G-8 binary: key absent entirely
        setHTML(eqWrap, '<div class="dim">equipment — endpoint pending (G-8)</div>');
      } else if (!s.equipment || !s.equipment.length) {
        setHTML(eqWrap, '<div class="dim">nothing worn</div>');
      } else {
        setHTML(eqWrap, s.equipment.map(q =>
          `<div class="bd-item worn">${esc(q.slot_name)}: ${esc(q.name)}</div>`).join(''));
      }
      const b = s.bonuses;
      setText($('.bd-bonus'), b ? `armour ${b.armour} · aim ${b.weapon_aim} · power ${b.weapon_power} · magic ${b.magic} · prayer ${b.prayer}` : '');

      const gr = $('.bd-ground');
      if (s.ground_items === undefined) setHTML(gr, '<div class="dim">ground items — endpoint pending (G-10)</div>');
      else if (!s.ground_items || !s.ground_items.length) setHTML(gr, '<div class="dim">nothing on the ground in view</div>');
      else setHTML(gr, s.ground_items.map(g =>
        `<div class="bd-item">${esc(g.name)} (${esc(g.x)},${esc(g.y)})</div>`).join(''));

      const dlg = $('.bd-dialog');
      // dialog_open is true only when an OPTION MENU is up; a plain NPC line
      // populates dialog_text alone — show either, options only when present.
      if (s.dialog_open || s.dialog_text) {
        setHTML(dlg, `<div class="bd-dtext">“${esc(s.dialog_text || '')}”</div>` +
          (s.dialog_options || []).map((o, i) => `<div class="bd-dopt">${i + 1}. ${esc(o)}</div>`).join(''));
      } else setHTML(dlg, '<div class="dim">no dialog open</div>');

      const msgs = (s.recent_server_messages || []).slice(-8);
      setHTML($('.bd-msgs'), msgs.length
        ? msgs.map(x => `<div class="bd-msg">${esc(x)}</div>`).join('')
        : '<div class="dim">no recent server messages</div>');
    }

    // -- QUESTIONS: every open question + forage chips --------------------------
    function renderQuestions() {
      if (!mind) return;
      const t = buildTree();
      const qs = [...t.nodes.values()].filter(n =>
        n.kind === 'open_question' && n.status !== 'done' && n.status !== 'abandoned');
      setText($('.cnt-q'), String(qs.length));
      $('.q-empty').hidden = qs.length > 0;
      renderKeyed($('.q-list'), qs, n => n.id, () => {
        const d = document.createElement('div');
        d.className = 'qrow';
        d.innerHTML = '<span class="qglyph">?</span><span class="qlabel"></span><span class="qchips"></span><span class="qunder"></span>';
        return d;
      }, (elr, n) => {
        setText(elr.querySelector('.qlabel'), n.label || n.id);
        setText(elr.querySelector('.qchips'), forageChips(n.tags));
        const p = t.parent.get(n.id);
        setText(elr.querySelector('.qunder'), p ? '↳ ' + p : '');
      });
    }

    // first paint for already-cached values happens via the immediate-fire
    // subscriptions above; cleanup detaches them all.
    return () => { for (const u of unsubs) u(); };
  };

  // ============================================================================
  // skeleton + styles
  // ============================================================================
  const SKELETON = `
<div id="tablet">
  <div id="t-head">
    <div class="th-row th-1">
      <span class="t-dot dot"></span><span class="t-name"></span><span class="t-status"></span>
      <span class="t-vital">hp <span class="bar hp"><i></i></span> <span class="t-hp"></span></span>
      <span class="t-vital">fatigue <span class="bar fat"><i></i></span> <span class="t-fat"></span></span>
      <span class="t-cmb"></span>
      <span class="t-hstale badge" hidden>STALE</span>
    </div>
    <div class="th-row th-2">
      <span class="t-floor"></span><span class="t-where"></span>
      <span class="t-restarts"></span><span class="t-anlzbadge badge" hidden>ANALYSIS</span>
      <span class="t-explored"></span>
      <span class="t-btns">
        <button data-act="pause" title="pause">⏸</button><button data-act="resume" title="resume">▶</button>
        <button data-act="restart" title="restart">↻</button><button data-act="stop" title="stop" class="danger">⏹</button>
        <button class="t-anlz" title="toggle analysis mode">ANLZ</button>
      </span>
    </div>
    <div class="th-row th-star" hidden>★ <span class="t-star"></span></div>
    <div class="th-row th-persona" hidden></div>
  </div>
  <div id="t-main">
    <div class="t-cog">
      <div class="pane-h">COGNITION <span class="dim">live path ══ · ⚑ neglected</span> <span class="t-cstale badge" hidden>STALE</span></div>
      <div class="cog-body"></div>
      <div class="cog-empty pane-empty" hidden>no goals yet</div>
    </div>
    <div class="t-right">
      <div class="t-tabs rt-tabs">
        <button data-tab="thoughts">THOUGHTS</button><button data-tab="decisions">DECISIONS</button>
        <span class="th-tools">
          <button data-k="thought">thought</button><button data-k="note">note</button>
          <button data-k="veto">veto</button><button data-k="system">sys</button><button data-k="chat">chat</button>
          <button class="th-pause" title="pause feed">⏸</button>
        </span>
      </div>
      <div class="rt-pane th-pane">
        <div class="rt-scroll th-feed"></div>
        <div class="th-empty pane-empty" hidden>no thoughts yet</div>
        <button class="pin-pill th-pill" hidden></button>
      </div>
      <div class="rt-pane dec-pane" hidden>
        <div class="rt-scroll dec-list"></div>
        <div class="dec-pending pane-empty" hidden>decisions — endpoint pending (G-4)</div>
        <div class="dec-empty pane-empty" hidden>no decisions yet</div>
      </div>
      <div id="t-console">
        <div class="con-h">
          <button class="con-toggle">CONSOLE ▾</button>
          <button class="cm" data-m="eval">eval</button><button class="cm" data-m="script">script</button><button class="cm" data-m="analysis">analysis</button><button class="cm" data-m="whisper">whisper</button>
          <select class="con-urg" title="whisper urgency — high interrupts the current turn" hidden><option value="low">low</option><option value="normal" selected>normal</option><option value="high">high</option></select>
        </div>
        <div class="con-body">
          <div class="con-out"></div>
          <div class="con-in-row"><textarea class="con-in" rows="1"></textarea><button class="con-run">⏎</button></div>
        </div>
      </div>
    </div>
  </div>
  <div id="t-drawer">
    <div class="t-tabs dr-tabs">
      <button data-tab="knowledge">KNOWLEDGE <span class="cnt cnt-k">0</span></button>
      <button data-tab="relationships">RELATIONSHIPS <span class="cnt cnt-r">0</span></button>
      <button data-tab="exploration">EXPLORATION</button>
      <button data-tab="body">BODY</button>
      <button data-tab="questions">QUESTIONS <span class="cnt cnt-q">0</span></button>
    </div>
    <div class="dr-body">
      <div class="dr-pane p-knowledge">
        <div class="dr-toolbar kn-toolbar">
          prov <button data-p="system">sys</button><button data-p="observed">obs</button><button data-p="deduced">ded</button><button data-p="hearsay">hearsay</button>
          kind <select class="kn-kind"><option value="all">all</option></select>
          seen <select class="kn-seen"><option value="all">all</option><option value="10m">10m</option><option value="1h">1h</option><option value="24h">24h</option></select>
          <input type="search" class="kn-q" placeholder="search subject / claim">
          sort <select class="kn-sort"><option value="rank">rank</option><option value="confidence">confidence</option><option value="subject">subject</option><option value="familiar">familiar</option><option value="beliefs">beliefs</option><option value="updated">updated</option></select>
          <span class="kn-n"></span>
        </div>
        <div class="khead khdr"><span class="kx"></span><span class="ksub">SUBJECT</span><span class="kkind">KIND</span><span class="kclaim">TOP CLAIM</span><span class="kconf">CONF</span><span class="kprov">PROV</span><span class="kab">α/β</span><span class="kblf">BLF</span><span class="kupd">UPD</span></div>
        <div class="kn-list"></div>
        <div class="kn-empty pane-empty" hidden>no knowledge yet</div>
      </div>
      <div class="dr-pane p-relationships">
        <div class="dr-toolbar">
          sort <select class="rel-sort"><option value="closeness">closeness</option><option value="trust">trust</option><option value="grievance">grievance</option><option value="recent">recent</option><option value="familiar">familiar</option></select>
          <span class="dim">t trust · a affinity · g grievance</span>
        </div>
        <div class="rel-list"></div>
        <div class="rel-empty pane-empty" hidden>no one known yet</div>
      </div>
      <div class="dr-pane p-exploration">
        <div class="fog-pending pane-empty" hidden>exploration — endpoint pending (G-2n)</div>
        <div class="fog-data">
          <div class="fog-cov"></div>
          <div class="fog-here"></div>
          <div class="fog-fr-h dim">frontiers (best first — RSC compass: east = smaller x):</div>
          <div class="fog-frontiers"></div>
          <div class="fog-adj dim"></div>
        </div>
      </div>
      <div class="dr-pane p-body">
        <div class="bd-h">SKILLS</div><div class="bd-skills"></div>
        <div class="bd-h bd-inv-h">INVENTORY</div><div class="bd-inv"></div>
        <div class="bd-inv-empty dim" hidden>inventory empty</div>
        <div class="bd-h">EQUIPMENT</div><div class="bd-eq"></div><div class="bd-bonus dim"></div>
        <div class="bd-h">GROUND</div><div class="bd-ground"></div>
        <div class="bd-h">DIALOG</div><div class="bd-dialog"></div>
        <div class="bd-h">SERVER MESSAGES</div><div class="bd-msgs"></div>
      </div>
      <div class="dr-pane p-questions">
        <div class="q-list"></div>
        <div class="q-empty pane-empty" hidden>no open questions</div>
      </div>
    </div>
  </div>
</div>`;

  // Tablet styles are injected here (style.css is the shell's file; the tablet
  // module stays self-contained). Idempotent across remounts.
  (function injectCSS() {
    if (document.getElementById('tablet-css')) return;
    const st = document.createElement('style');
    st.id = 'tablet-css';
    st.textContent = `
#tablet { flex:1; min-height:0; display:flex; flex-direction:column; }
#tablet .pane-empty { color:var(--faint); padding:10px 14px; font-size:12px; }
#tablet .dim { color:var(--dim); }
#tablet .badge { color:var(--accent); border:1px solid var(--accent); border-radius:4px; padding:0 5px; font-size:9px; letter-spacing:1px; }
#tablet .stale { opacity:.6; }

/* header band — fixed-height rows (§4.5) */
#t-head { flex:none; border-bottom:1px solid var(--line); background:var(--panel); padding:3px 14px; }
.th-row { height:22px; display:flex; align-items:center; gap:12px; overflow:hidden; white-space:nowrap; }
.th-row .t-dot { width:9px; height:9px; }
.t-name { font-weight:600; letter-spacing:1px; }
.t-status { color:var(--dim); }
.t-vital { display:inline-flex; align-items:center; gap:6px; color:var(--dim); font-size:12px; }
#tablet .bar { display:inline-block; width:84px; height:9px; border:1px solid var(--line); background:var(--bg); }
#tablet .bar i { display:block; height:100%; width:0; background:var(--spark); }
#tablet .bar i.hot { background:var(--accent); }
#tablet .bar i.neg { background:var(--faint); }
.t-cmb { color:var(--dim); }
.th-2 { color:var(--dim); font-size:12px; }
.t-where { overflow:hidden; text-overflow:ellipsis; }
.t-btns { margin-left:auto; display:inline-flex; gap:4px; }
.t-anlz.on { color:var(--accent); border-color:var(--accent); }
.t-explored { color:var(--dim); }
.th-star { color:var(--dim); font-style:italic; }
.th-persona { height:24px; color:var(--dim); font-size:11px; gap:14px; }
/* §4.5: the persona rows are RESERVED — hidden keeps them in flow at fixed
   height (like the fleet .badge-a[hidden] pattern) so the once-per-selection
   persona fetch landing (or a 501 host) never shifts the panes below. */
.th-star[hidden], .th-persona[hidden] { display:flex; visibility:hidden; }
.trait { display:inline-flex; align-items:center; gap:3px; }
.trait b { color:var(--fg); font-weight:600; }
.trait .track { position:relative; display:inline-block; width:36px; height:9px; border:1px solid var(--line); background:var(--bg); }
.trait .bandseg { position:absolute; top:0; bottom:0; background:var(--panel2); }
.trait .mu { position:absolute; top:-1px; bottom:-1px; width:2px; background:var(--fg); }
.pgroup { white-space:nowrap; }

/* main split */
#t-main { flex:1; min-height:0; display:flex; }
.t-cog { flex:1.15; min-width:0; overflow-y:auto; border-right:1px solid var(--line); padding:4px 10px 14px; }
.pane-h { height:24px; display:flex; align-items:center; gap:10px; color:var(--dim); font-size:11px; letter-spacing:1px; position:sticky; top:0; background:var(--bg); }

/* cognition tree */
.cnode { display:flex; align-items:center; gap:7px; min-height:20px; overflow:hidden; white-space:nowrap; border-left:2px solid transparent; }
.cnode .cglyph { flex:none; width:2ch; text-align:center; color:var(--dim); }
.cnode .clabel { min-width:0; overflow:hidden; text-overflow:ellipsis; }
.cnode.casp .clabel { color:var(--fg); letter-spacing:.5px; }
.cnode.terminal { color:var(--faint); }
.cnode.terminal .clabel { text-decoration:line-through; }
.cnode.spine { border-left-color:var(--accent); }
.cnode.spine .cglyph { color:var(--accent); }
.cnode.flash { animation:cogflash 1.2s ease-out 1; }
@keyframes cogflash { 0% { background:var(--panel2); } 100% { background:transparent; } }
.cbar { display:inline-block; width:56px; height:7px; border:1px solid var(--line); background:var(--bg); }
.cbar i { display:block; height:100%; width:0; background:var(--spark); }
.croll, .cprog { display:inline-flex; align-items:center; gap:6px; flex:none; color:var(--dim); font-size:11px; }
.cneg { color:var(--fg); font-size:10px; letter-spacing:1px; }
.cblk { flex:none; color:var(--fg); border:1px solid var(--line); border-radius:4px; padding:0 5px; font-size:10px; cursor:pointer; }
.cchips { flex:none; color:var(--dim); font-size:10px; }
.cchip { color:var(--faint); font-size:11px; cursor:pointer; min-height:18px; }
.cchip:hover { color:var(--dim); }
.croutine { border-left:2px solid var(--accent); margin:2px 0 6px; }
.crhead { height:20px; display:flex; align-items:center; gap:8px; color:var(--dim); font-size:11px; letter-spacing:1px; }
.crhead .crmark { color:var(--accent); }
.crhead .crname { color:var(--fg); }
.crbox { max-height:300px; overflow-y:auto; border:1px solid var(--line); background:var(--panel); margin-left:14px; }
.crbox pre { margin:0; font-size:11px; line-height:1.5; }
.crbox .ln { display:flex; }
.crbox .ln .num { flex:none; width:4ch; text-align:right; padding-right:8px; color:var(--faint); user-select:none; }
.crbox .ln .lsrc { white-space:pre; }
.crbox .ln.cur { background:var(--panel2); }
.crbox .ln.cur .num { color:var(--accent); }
.crbox .ln.cur .lsrc { color:var(--fg); }

/* right column */
.t-right { flex:1; min-width:0; display:flex; flex-direction:column; }
.t-tabs { height:26px; flex:none; display:flex; align-items:center; gap:4px; padding:0 10px; border-bottom:1px solid var(--line); background:var(--panel); overflow:hidden; }
.t-tabs > button { border:none; background:none; color:var(--dim); letter-spacing:1px; font-size:11px; padding:5px 8px; }
.t-tabs > button.on { color:var(--fg); border-bottom:1px solid var(--fg); border-radius:0; }
.th-tools { margin-left:auto; display:inline-flex; gap:4px; }
.th-tools button { font-size:10px; padding:0 6px; color:var(--faint); }
.th-tools button.on { color:var(--fg); background:var(--line); }
.rt-pane { flex:1; min-height:0; position:relative; display:flex; flex-direction:column; }
.rt-scroll { flex:1; min-height:0; overflow-y:auto; padding:4px 10px; }
.tcap { color:var(--faint); font-size:10px; text-align:center; padding:2px 0; }
.trow { padding:3px 0; border-bottom:1px solid var(--panel2); font-size:12px; }
.trow .tt { color:var(--faint); font-size:11px; }
.trow .t1 { color:var(--dim); }
.trow .t2 { color:var(--fg); white-space:pre-wrap; }
.trow .t2i { color:var(--fg); }
.trow .t3 { color:var(--dim); font-size:11px; }
.trow.g-veto .t1 { color:var(--accent); }
.trow.g-chat .t1 { color:var(--dim); }
.drow { padding:3px 0; border-bottom:1px solid var(--panel2); font-size:12px; }
.drow .d1 { display:flex; gap:10px; color:var(--dim); }
.drow .dkind { color:var(--fg); }
.drow .dtrig { overflow:hidden; text-overflow:ellipsis; white-space:nowrap; min-width:0; }
.drow .dgoal { color:var(--faint); font-size:11px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.drow .dreason { color:var(--fg); white-space:pre-wrap; }

/* console */
#t-console { flex:none; border-top:1px solid var(--line); background:var(--panel); }
.con-h { height:24px; display:flex; align-items:center; gap:4px; padding:0 10px; }
.con-h .con-toggle { border:none; background:none; color:var(--dim); letter-spacing:1px; font-size:11px; }
.con-h .cm { font-size:10px; padding:0 6px; color:var(--faint); }
.con-h .cm.on { color:var(--fg); background:var(--line); }
.con-h .con-urg { font-size:10px; background:var(--bg); color:var(--dim); border:1px solid var(--line); border-radius:4px; }
.con-body { padding:0 10px 8px; }
.con-out { max-height:110px; overflow-y:auto; font-size:11px; padding:2px 0; }
.con-line { color:var(--dim); white-space:pre-wrap; }
.con-line.err { color:var(--accent); }
.con-in-row { display:flex; gap:6px; align-items:flex-end; }
.con-in { flex:1; resize:vertical; background:var(--bg); color:var(--fg); border:1px solid var(--line); border-radius:4px; font:inherit; font-size:12px; padding:3px 8px; }

/* drawer */
#t-drawer { flex:none; height:36%; min-height:0; display:flex; flex-direction:column; border-top:1px solid var(--line); }
#t-drawer.closed { height:auto; }
.dr-body { flex:1; min-height:0; display:flex; flex-direction:column; }
.dr-pane { flex:1; min-height:0; overflow-y:auto; padding:0 14px 10px; }
.dr-toolbar { position:sticky; top:0; background:var(--bg); display:flex; align-items:center; gap:8px; min-height:30px; color:var(--dim); font-size:11px; flex-wrap:wrap; z-index:1; }
.dr-toolbar button { font-size:10px; padding:0 6px; color:var(--faint); }
.dr-toolbar button.on { color:var(--fg); background:var(--line); }
.dr-toolbar input[type="search"] { width:170px; }
.cnt { color:var(--faint); }
.t-tabs button.on .cnt { color:var(--dim); }

/* knowledge table */
.khead { display:grid; grid-template-columns:2ch 18ch 7ch 1fr 5ch 9ch 8ch 4ch 5ch; gap:8px; align-items:center; min-height:20px; font-size:12px; cursor:pointer; }
.khead > span { overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.khead .kconf, .khead .kab, .khead .kblf, .khead .kupd { text-align:right; }
.khdr { color:var(--faint); font-size:10px; letter-spacing:1px; cursor:default; position:sticky; top:30px; background:var(--bg); z-index:1; }
.krow { border-bottom:1px solid var(--panel2); }
.krow .kkind, .krow .kprov, .krow .kupd { color:var(--dim); }
.krow .kclaim { color:var(--fg); }
.kdetails { padding:2px 0 4px 3ch; }
.kbel { color:var(--dim); font-size:11px; }

/* relationships */
.rrow { display:grid; grid-template-columns:12ch 9ch 16ch 17ch 16ch 7ch 5ch 8ch 1fr; gap:8px; align-items:center; min-height:22px; font-size:12px; border-bottom:1px solid var(--panel2); }
.rrow > span { overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.rrow .rgrade { color:var(--dim); }
.rrow .rbarw { display:inline-flex; align-items:center; gap:4px; color:var(--faint); font-size:11px; }
.rrow .rbarw .bar { width:56px; height:8px; }
.rrow .rtv, .rrow .rav, .rrow .rgv { color:var(--dim); }
.rrow .rint, .rrow .rlast, .rrow .rfam { color:var(--dim); font-size:11px; }
.rrow .rtags { color:var(--faint); font-size:11px; }

/* exploration */
.p-exploration .fog-cov { color:var(--fg); padding-top:8px; }
.p-exploration .fog-here { color:var(--dim); }
.p-exploration .fog-fr { color:var(--fg); padding-left:2ch; }
.p-exploration .fog-adj { padding-top:4px; }

/* body */
.bd-h { color:var(--faint); font-size:10px; letter-spacing:1px; padding:8px 0 2px; }
.bd-skills { display:grid; grid-template-columns:repeat(6, 1fr); gap:2px 14px; }
.bd-skill { display:flex; gap:6px; font-size:11px; align-items:baseline; }
.bd-skill b { color:var(--dim); font-weight:400; width:9ch; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.bd-skill .lv { color:var(--fg); }
.bd-skill .lv.boost { color:var(--accent); }
.bd-skill .xp { color:var(--faint); font-size:10px; }
.bd-inv { display:grid; grid-template-columns:repeat(3, 1fr); gap:1px 14px; }
.bd-item { font-size:11px; color:var(--fg); overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.bd-item.worn { color:var(--dim); }
.bd-dtext { color:var(--fg); }
.bd-dopt { color:var(--dim); padding-left:2ch; }
.bd-msg { color:var(--dim); font-size:11px; }

/* questions */
.qrow { display:flex; gap:8px; align-items:center; min-height:20px; font-size:12px; border-bottom:1px solid var(--panel2); overflow:hidden; white-space:nowrap; }
.qrow .qglyph { flex:none; color:var(--dim); }
.qrow .qlabel { min-width:0; overflow:hidden; text-overflow:ellipsis; }
.qrow .qchips { flex:none; color:var(--dim); font-size:10px; }
.qrow .qunder { flex:none; color:var(--faint); font-size:10px; max-width:30ch; overflow:hidden; text-overflow:ellipsis; }
`;
    document.head.appendChild(st);
  })();
})();
