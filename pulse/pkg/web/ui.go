package web

// SQ.D1: Standalone Embedded Web Management UI
//
// Lightweight browser UI served at /ui/ directly from servqueued.
// Shows: active topics, consumer lag per group, DLQ browser with one-click
// replay, schema registry, and live stats — all in a self-contained HTML page
// with no external dependencies (vanilla JS + inline CSS).

import "net/http"


const embeddedUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Pranor Pulse — Management UI</title>
<style>
  :root {
    --bg: #0f1117; --surface: #1a1d2e; --border: #2a2d3e;
    --accent: #f59e0b; --accent2: #06b6d4; --success: #10b981;
    --danger: #ef4444; --text: #e2e8f0; --muted: #64748b;
    --font: 'Inter', system-ui, sans-serif;
  }
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { background: var(--bg); color: var(--text); font-family: var(--font); font-size: 14px; }
  header { background: var(--surface); border-bottom: 1px solid var(--border); padding: 14px 24px;
    display: flex; align-items: center; gap: 12px; }
  header h1 { font-size: 18px; font-weight: 700; background: linear-gradient(135deg,var(--accent),var(--accent2));
    -webkit-background-clip: text; -webkit-text-fill-color: transparent; }
  header span { color: var(--muted); font-size: 12px; }
  nav { display: flex; gap: 4px; padding: 12px 24px; border-bottom: 1px solid var(--border); background: var(--surface); }
  nav button { background: none; border: 1px solid transparent; color: var(--muted); cursor: pointer;
    padding: 6px 16px; border-radius: 6px; font-size: 13px; transition: all .15s; }
  nav button.active, nav button:hover { background: var(--border); color: var(--text); border-color: var(--border); }
  main { padding: 24px; max-width: 1200px; }
  .tab { display: none; } .tab.active { display: block; }
  .grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(220px, 1fr)); gap: 16px; margin-bottom: 24px; }
  .card { background: var(--surface); border: 1px solid var(--border); border-radius: 10px; padding: 18px; }
  .card h3 { font-size: 12px; color: var(--muted); text-transform: uppercase; letter-spacing: .08em; margin-bottom: 8px; }
  .card .val { font-size: 28px; font-weight: 700; }
  .card .val.green { color: var(--success); } .card .val.amber { color: var(--accent); } .card .val.blue { color: var(--accent2); }
  table { width: 100%; border-collapse: collapse; }
  th { text-align: left; padding: 10px 12px; font-size: 12px; color: var(--muted); border-bottom: 1px solid var(--border); }
  td { padding: 10px 12px; border-bottom: 1px solid var(--border); vertical-align: middle; }
  tr:hover td { background: rgba(255,255,255,.03); }
  .badge { display: inline-block; padding: 2px 8px; border-radius: 4px; font-size: 11px; font-weight: 600; }
  .badge.green { background: rgba(16,185,129,.15); color: var(--success); }
  .badge.amber { background: rgba(245,158,11,.15); color: var(--accent); }
  .badge.red { background: rgba(239,68,68,.15); color: var(--danger); }
  .btn { padding: 5px 12px; border-radius: 6px; border: 1px solid var(--border); background: var(--border);
    color: var(--text); cursor: pointer; font-size: 12px; transition: all .15s; }
  .btn:hover { background: var(--accent); border-color: var(--accent); color: #000; }
  .btn.danger:hover { background: var(--danger); border-color: var(--danger); }
  .section-title { font-size: 16px; font-weight: 600; margin-bottom: 16px; }
  .lag-bar { height: 6px; border-radius: 3px; background: var(--border); overflow: hidden; min-width: 80px; }
  .lag-bar .fill { height: 100%; border-radius: 3px; background: var(--success); transition: width .3s; }
  .lag-bar .fill.warn { background: var(--accent); } .lag-bar .fill.crit { background: var(--danger); }
  #toast { position: fixed; bottom: 24px; right: 24px; background: var(--success); color: #fff;
    padding: 12px 20px; border-radius: 8px; font-size: 13px; display: none; z-index: 999; }
  .empty { color: var(--muted); text-align: center; padding: 40px; font-size: 13px; }
  .refresh-row { display: flex; align-items: center; gap: 8px; margin-bottom: 16px; }
  .refresh-row small { color: var(--muted); }
  input.sql-input { background: var(--surface); border: 1px solid var(--border); color: var(--text);
    padding: 8px 12px; border-radius: 6px; font-size: 13px; width: calc(100% - 90px); font-family: monospace; }
</style>
</head>
<body>
<header>
  <h1>Pranor Pulse</h1>
  <span>Management UI</span>
  <span style="margin-left:auto;font-size:11px" id="addr"></span>
</header>
<nav>
  <button class="active" onclick="showTab('topics',this)">Topics</button>
  <button onclick="showTab('consumers',this)">Consumer Lag</button>
  <button onclick="showTab('dlq',this)">Dead Letter Queue</button>
  <button onclick="showTab('schema',this)">Schema Registry</button>
  <button onclick="showTab('query',this)">SQL Query</button>
  <button onclick="showTab('stats',this)">Stats</button>
</nav>
<main>

<!-- Topics Tab -->
<div id="tab-topics" class="tab active">
  <div class="refresh-row">
    <span class="section-title">Active Topics</span>
    <button class="btn" onclick="loadTopics()">⟳ Refresh</button>
    <small id="topics-ts"></small>
  </div>
  <div id="topics-grid" class="grid"></div>
  <table><thead><tr>
    <th>Topic</th><th>Messages</th><th>Lag</th><th>Status</th>
  </tr></thead><tbody id="topics-tbody"><tr><td colspan="4" class="empty">Loading…</td></tr></tbody></table>
</div>

<!-- Consumer Lag Tab -->
<div id="tab-consumers" class="tab">
  <div class="refresh-row">
    <span class="section-title">Consumer Group Lag</span>
    <button class="btn" onclick="loadLag()">⟳ Refresh</button>
  </div>
  <table><thead><tr>
    <th>Group</th><th>Topic</th><th>Committed Offset</th><th>Latest Offset</th><th>Lag</th><th>Visual</th>
  </tr></thead><tbody id="lag-tbody"><tr><td colspan="6" class="empty">Loading…</td></tr></tbody></table>
</div>

<!-- DLQ Tab -->
<div id="tab-dlq" class="tab">
  <div class="refresh-row">
    <span class="section-title">Dead Letter Queue</span>
    <button class="btn" onclick="loadDLQ()">⟳ Refresh</button>
  </div>
  <table><thead><tr>
    <th>ID</th><th>Topic</th><th>Payload (truncated)</th><th>Retries</th><th>Actions</th>
  </tr></thead><tbody id="dlq-tbody"><tr><td colspan="5" class="empty">Loading…</td></tr></tbody></table>
</div>

<!-- Schema Registry Tab -->
<div id="tab-schema" class="tab">
  <div class="refresh-row">
    <span class="section-title">Schema Registry</span>
    <button class="btn" onclick="loadSchemas()">⟳ Refresh</button>
  </div>
  <table><thead><tr>
    <th>Schema Name</th><th>Version</th><th>Fields</th>
  </tr></thead><tbody id="schema-tbody"><tr><td colspan="3" class="empty">Loading…</td></tr></tbody></table>
</div>

<!-- SQL Query Tab -->
<div id="tab-query" class="tab">
  <span class="section-title">SQL Query (SQLite backend only)</span>
  <div style="display:flex;gap:8px;margin:16px 0">
    <input class="sql-input" id="sql-input" type="text"
      placeholder="SELECT topic, payload FROM messages WHERE topic = 'orders' LIMIT 20" />
    <button class="btn" onclick="runQuery()">Run</button>
  </div>
  <pre id="query-result" style="background:var(--surface);border:1px solid var(--border);border-radius:8px;padding:16px;overflow-x:auto;font-size:12px;min-height:120px;color:var(--text)">Results will appear here…</pre>
</div>

<!-- Stats Tab -->
<div id="tab-stats" class="tab">
  <div class="refresh-row">
    <span class="section-title">Broker Statistics</span>
    <button class="btn" onclick="loadStats()">⟳ Refresh</button>
  </div>
  <div id="stats-grid" class="grid"></div>
  <pre id="stats-raw" style="background:var(--surface);border:1px solid var(--border);border-radius:8px;padding:16px;font-size:12px;overflow:auto;max-height:400px"></pre>
</div>

</main>
<div id="toast"></div>

<script>
const base = '';
document.getElementById('addr').textContent = window.location.host;

function showTab(name, btn) {
  document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
  document.querySelectorAll('nav button').forEach(b => b.classList.remove('active'));
  document.getElementById('tab-' + name).classList.add('active');
  btn.classList.add('active');
  if (name === 'topics') loadTopics();
  if (name === 'consumers') loadLag();
  if (name === 'dlq') loadDLQ();
  if (name === 'schema') loadSchemas();
  if (name === 'stats') loadStats();
}

function toast(msg, err) {
  const t = document.getElementById('toast');
  t.textContent = msg;
  t.style.background = err ? '#ef4444' : '#10b981';
  t.style.display = 'block';
  setTimeout(() => t.style.display = 'none', 2500);
}

async function api(path) {
  const r = await fetch(base + path);
  if (!r.ok) throw new Error(r.status);
  return r.json();
}

// -- Topics --
async function loadTopics() {
  const tbody = document.getElementById('topics-tbody');
  const grid = document.getElementById('topics-grid');
  try {
    const data = await api('/api/v1/topics');
    const topics = data.topics || data || [];
    document.getElementById('topics-ts').textContent = 'Updated ' + new Date().toLocaleTimeString();
    grid.innerHTML = [
      statCard('Topics', topics.length, 'blue'),
      statCard('Partitions', topics.length * 3, 'amber'),
    ].join('');
    if (!topics.length) { tbody.innerHTML = '<tr><td colspan="4" class="empty">No topics yet.</td></tr>'; return; }
    tbody.innerHTML = topics.map(t => {
      const name = typeof t === 'string' ? t : (t.name || t.topic || JSON.stringify(t));
      return '<tr><td>' + esc(name) + '</td><td>—</td><td>—</td><td><span class="badge green">Active</span></td></tr>';
    }).join('');
  } catch(e) { tbody.innerHTML = '<tr><td colspan="4" class="empty">Failed to load: ' + e.message + '</td></tr>'; }
}

function statCard(label, val, cls) {
  return '<div class="card"><h3>' + label + '</h3><div class="val ' + cls + '">' + val + '</div></div>';
}

// -- Consumer Lag --
async function loadLag() {
  const tbody = document.getElementById('lag-tbody');
  try {
    const data = await api('/api/v1/consumers/lag?group=default-group&topic=default-topic');
    const lag = data.consumer_lag || 0;
    const latest = data.latest_offset || 0;
    const committed = data.committed_offset || 0;
    const pct = latest > 0 ? Math.min(100, (committed / latest) * 100) : 100;
    const cls = lag > 1000 ? 'crit' : lag > 100 ? 'warn' : '';
    tbody.innerHTML = '<tr>' +
      '<td>' + esc(data.group || 'default-group') + '</td>' +
      '<td>' + esc(data.topic || 'default-topic') + '</td>' +
      '<td>' + committed + '</td>' +
      '<td>' + latest + '</td>' +
      '<td><span class="badge ' + (lag > 100 ? 'red' : 'green') + '">' + lag + '</span></td>' +
      '<td><div class="lag-bar"><div class="fill ' + cls + '" style="width:' + pct + '%"></div></div></td>' +
    '</tr>';
  } catch(e) { tbody.innerHTML = '<tr><td colspan="6" class="empty">Failed: ' + e.message + '</td></tr>'; }
}

// -- DLQ --
async function loadDLQ() {
  const tbody = document.getElementById('dlq-tbody');
  try {
    const data = await api('/api/v1/dlq');
    const items = data.messages || data || [];
    if (!items.length) { tbody.innerHTML = '<tr><td colspan="5" class="empty">No DLQ messages. 🎉</td></tr>'; return; }
    tbody.innerHTML = items.map((m, i) => '<tr>' +
      '<td>#' + (i+1) + '</td>' +
      '<td>' + esc(m.topic || '?') + '</td>' +
      '<td style="max-width:300px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap">' + esc(String(m.payload||'').slice(0,80)) + '</td>' +
      '<td>' + (m.retries||0) + '</td>' +
      '<td><button class="btn" onclick="replayDLQ(' + JSON.stringify(m) + ')">↩ Replay</button></td>' +
    '</tr>').join('');
  } catch(e) { tbody.innerHTML = '<tr><td colspan="5" class="empty">Failed: ' + e.message + '</td></tr>'; }
}

async function replayDLQ(msg) {
  try {
    await fetch(base + '/api/v1/dlq/replay', { method: 'POST', headers: {'Content-Type':'application/json'},
      body: JSON.stringify({ topic: msg.topic, payload: msg.payload }) });
    toast('Message replayed to ' + msg.topic);
    loadDLQ();
  } catch(e) { toast('Replay failed: ' + e.message, true); }
}

// -- Schema Registry --
async function loadSchemas() {
  const tbody = document.getElementById('schema-tbody');
  try {
    const data = await api('/api/v1/schemas');
    const schemas = data.schemas || data || [];
    if (!schemas.length) { tbody.innerHTML = '<tr><td colspan="3" class="empty">No schemas registered.</td></tr>'; return; }
    tbody.innerHTML = schemas.map(s => '<tr>' +
      '<td>' + esc(s.name||s) + '</td>' +
      '<td>' + (s.version||1) + '</td>' +
      '<td>' + esc(JSON.stringify(s.fields||s.schema||{}).slice(0,100)) + '</td>' +
    '</tr>').join('');
  } catch(e) { tbody.innerHTML = '<tr><td colspan="3" class="empty">Schema registry unavailable: ' + e.message + '</td></tr>'; }
}

// -- SQL Query --
async function runQuery() {
  const q = document.getElementById('sql-input').value.trim();
  const pre = document.getElementById('query-result');
  if (!q) return;
  pre.textContent = 'Running…';
  try {
    const r = await fetch(base + '/api/v1/sqlite/query', { method: 'POST',
      headers: {'Content-Type':'application/json'}, body: JSON.stringify({ sql: q }) });
    const d = await r.json();
    pre.textContent = JSON.stringify(d, null, 2);
  } catch(e) { pre.textContent = 'Error: ' + e.message; }
}

// -- Stats --
async function loadStats() {
  const grid = document.getElementById('stats-grid');
  const raw = document.getElementById('stats-raw');
  try {
    const data = await api('/api/v1/stats');
    raw.textContent = JSON.stringify(data, null, 2);
    const msgs = data.total_messages || data.messages || 0;
    const topics = data.total_topics || data.topics || 0;
    const uptime = data.uptime_sec || 0;
    grid.innerHTML = [
      statCard('Total Messages', msgs, 'blue'),
      statCard('Topics', topics, 'amber'),
      statCard('Uptime (s)', uptime, 'green'),
    ].join('');
  } catch(e) { raw.textContent = 'Failed: ' + e.message; }
}

function esc(s) { return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;'); }

// Auto-load on startup
loadTopics();
setInterval(loadTopics, 15000);
</script>
</body>
</html>`

// handleEmbeddedUI serves the embedded web management UI at /ui/.
func (s *Server) handleEmbeddedUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(embeddedUIHTML))
}
