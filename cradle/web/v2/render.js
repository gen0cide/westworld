// cradle v2 — render.js: the tiny DOM toolkit that replaces lit-html (no build
// step, no CDN). Keyed list reconciliation, HTML escaping, targeted text
// updates, plus the §4 live-region scroll helpers every pane must use.
// Everything hangs off window.CradleV2 so app.js / tablet.js share one kit.
(function () {
  'use strict';
  const NS = (window.CradleV2 = window.CradleV2 || {});

  // esc HTML-escapes a value for innerHTML interpolation.
  NS.esc = s => String(s ?? '').replace(/[&<>"']/g, c =>
    ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));

  // setText writes textContent only when it actually changed, so steady-state
  // refreshes never disturb layout or an in-progress text selection (§4.1/§4.5).
  NS.setText = (el, text) => {
    text = String(text ?? '');
    if (el && el.textContent !== text) el.textContent = text;
  };

  // setClass toggles a class only on change (no churn for CSS transitions).
  NS.setClass = (el, cls, on) => {
    if (el && el.classList.contains(cls) !== !!on) el.classList.toggle(cls, !!on);
  };

  // renderKeyed reconciles container's children against items, by key.
  //   key(item) -> string        stable identity
  //   create(item) -> Element    skeleton, called ONCE per key
  //   update(el, item, i)        patch targeted fields (setText/setClass/canvas)
  // Existing nodes are MOVED, never recreated, so canvases, scroll positions and
  // text selections survive reorders (§4.1). Container must be exclusively
  // managed by this helper.
  NS.renderKeyed = function (container, items, key, create, update) {
    const byKey = container.__keyed || (container.__keyed = new Map());
    const seen = new Set();
    let cursor = container.firstChild;
    for (let i = 0; i < items.length; i++) {
      const it = items[i], k = String(key(it));
      seen.add(k);
      let el = byKey.get(k);
      if (!el) { el = create(it); el.__key = k; byKey.set(k, el); }
      if (el === cursor) cursor = cursor.nextSibling;
      else container.insertBefore(el, cursor);
      update(el, it, i);
    }
    for (const [k, el] of byKey) {
      if (!seen.has(k)) { byKey.delete(k); el.remove(); }
    }
  };

  // followPane implements §4.2/§4.3 for pointer-follow panes (e.g. the routine
  // line highlight): auto-follow may only move THIS pane's own scrollTop, and
  // any user-originated scroll (wheel, trackpad momentum, scrollbar drag,
  // keyboard — they all fire 'scroll') suspends following for quietMs.
  // Programmatic moves are time-flagged so they don't count as manual.
  NS.followPane = function (el, quietMs = 4000) {
    let userAt = 0, progAt = 0;
    el.addEventListener('scroll', () => {
      if (Date.now() - progAt < 120) return; // our own move echoing back
      userAt = Date.now();
    }, { passive: true });
    return {
      canFollow: () => Date.now() - userAt > quietMs,
      // follow centres a child row inside the pane when it drifts off-screen.
      follow(row, margin = 12) {
        if (!row || !this.canFollow()) return;
        const br = el.getBoundingClientRect(), rr = row.getBoundingClientRect();
        if (rr.top < br.top + margin || rr.bottom > br.bottom - margin) {
          progAt = Date.now();
          el.scrollTop += (rr.top - br.top) - br.height / 2;
        }
      },
      scrollTo(top) { progAt = Date.now(); el.scrollTop = top; },
    };
  };

  // feedPin implements §4.4 for append-only feeds: stick to the tail only when
  // already at the tail; otherwise append silently and light a '▼ N new' pill
  // that re-pins on click. Usage per batch:
  //   const was = pin.atBottom();  ...append rows...  pin.appended(n, was);
  NS.feedPin = function (el, pill) {
    let progAt = 0, newCount = 0;
    const atBottom = () => el.scrollHeight - el.scrollTop - el.clientHeight < 30;
    const show = () => {
      if (!pill) return;
      pill.textContent = '▼ ' + newCount + ' new';
      pill.hidden = newCount === 0;
    };
    el.addEventListener('scroll', () => {
      if (Date.now() - progAt < 120) return;
      if (atBottom()) { newCount = 0; show(); } // user re-pinned by hand
    }, { passive: true });
    if (pill) pill.addEventListener('click', () => {
      newCount = 0; show();
      progAt = Date.now(); el.scrollTop = el.scrollHeight;
    });
    return {
      atBottom,
      appended(n, wasAtBottom) {
        if (wasAtBottom) { progAt = Date.now(); el.scrollTop = el.scrollHeight; newCount = 0; }
        else newCount += (n || 1);
        show();
      },
    };
  };

  // drawSpark renders a fixed-size bar sparkline; null samples are gaps (no
  // data), so a missing series reads as silence, not zero. Color comes from the
  // canvas's CSS color so the palette stays in style.css.
  NS.drawSpark = function (canvas, samples, max) {
    const ctx = canvas.getContext('2d');
    const w = canvas.width, h = canvas.height;
    ctx.clearRect(0, 0, w, h);
    if (!samples || !samples.length || !(max > 0)) return;
    ctx.fillStyle = getComputedStyle(canvas).color;
    const bw = w / samples.length;
    for (let i = 0; i < samples.length; i++) {
      const v = samples[i];
      if (v == null) continue;
      const bh = Math.max(1, Math.round((Math.min(v, max) / max) * (h - 1)));
      ctx.fillRect(Math.round(i * bw), h - bh, Math.max(1, Math.floor(bw) - 1), bh);
    }
  };
})();
