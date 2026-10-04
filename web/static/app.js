// rag-xray UI. Plain JS, no build step. Hash routes: #/learn/<slug>, #/xray, #/chunks
"use strict";

const $ = (s, el = document) => el.querySelector(s);
const el = (tag, props = {}, ...kids) => {
  const n = document.createElement(tag);
  for (const [k, v] of Object.entries(props)) {
    if (k === "class") n.className = v;
    else if (k === "text") n.textContent = v;
    else if (k.startsWith("on")) n.addEventListener(k.slice(2), v);
    else if (k === "style") n.setAttribute("style", v);
    else n.setAttribute(k, v);
  }
  for (const kid of kids) if (kid != null) n.append(kid);
  return n;
};

// Browser storage can be unavailable (private mode); never let it break the app.
const store = {
  get(k, d) { try { const v = localStorage.getItem(k); return v == null ? d : JSON.parse(v); } catch { return d; } },
  set(k, v) { try { localStorage.setItem(k, JSON.stringify(v)); } catch {} },
};

async function api(path, opts = {}) {
  const res = await fetch(path, opts);
  const body = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(body.error || res.statusText);
  return body;
}

let toastTimer;
function toast(msg) {
  const t = $("#toast");
  t.textContent = msg;
  t.classList.add("show");
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => t.classList.remove("show"), 2600);
}

async function copy(text) {
  try { await navigator.clipboard.writeText(text); toast("Copied"); } catch { toast("Copy failed"); }
}

// One stable colour per document, used in the doc list, hits and chunk cards.
const palette = ["#0f9f84", "#3b63d6", "#c2410c", "#9333ea", "#0891b2", "#ca8a04", "#db2777", "#4d7c0f"];
const docColor = (name) => {
  const i = info ? info.docs.findIndex((d) => d.name === name) : -1;
  if (i >= 0) return palette[i % palette.length];
  let h = 0;
  for (const ch of name) h = (h * 31 + ch.charCodeAt(0)) >>> 0;
  return palette[h % palette.length];
};

// ---------------------------------------------------------------- theme
(function theme() {
  const saved = store.get("theme", null);
  if (saved) document.documentElement.dataset.theme = saved;
  $("#theme").addEventListener("click", () => {
    const dark = document.documentElement.dataset.theme
      ? document.documentElement.dataset.theme === "dark"
      : matchMedia("(prefers-color-scheme: dark)").matches;
    const next = dark ? "light" : "dark";
    document.documentElement.dataset.theme = next;
    store.set("theme", next);
  });
})();

// ---------------------------------------------------------------- router
function route() {
  const [path, query = ""] = location.hash.split("?");
  const [, tab = "learn", slug] = path.split("/");
  for (const v of ["learn", "xray", "chunks"]) $("#view-" + v).hidden = v !== tab;
  document.querySelectorAll(".tabs a").forEach((a) => a.classList.toggle("on", a.dataset.tab === tab));
  if (tab === "learn") showLesson(slug);
  if (tab === "chunks") loadChunks();
  // Shareable X-ray links: #/xray?q=your+question
  const q = new URLSearchParams(query).get("q");
  if (tab === "xray" && q) { $("#question").value = q; whenReady(ask); }
  window.scrollTo(0, 0);
}

// Run fn once the index is ready (the first build can take a few seconds).
function whenReady(fn) {
  if (info && info.ready) return fn();
  setTimeout(() => whenReady(fn), 300);
}
window.addEventListener("hashchange", route);

// ---------------------------------------------------------------- learn
let lessons = [];
const done = new Set(store.get("done", []));

function renderLessonList(current) {
  const list = $("#lesson-list");
  list.replaceChildren(...lessons.map((l, i) =>
    el("li", {},
      el("a", { href: "#/learn/" + l.slug, class: [l.slug === current ? "on" : "", done.has(l.slug) ? "done" : ""].join(" ") },
        el("span", { class: "n", text: done.has(l.slug) ? "✓" : String(l.order ?? i) }),
        el("span", {},
          el("div", { class: "t", text: l.title }),
          el("div", { class: "m", text: l.minutes ? l.minutes + " min" : "" })))))
  );
  const n = lessons.filter((l) => done.has(l.slug)).length;
  $("#progress-text").textContent = `${n} / ${lessons.length}`;
  $("#progress-bar").style.width = lessons.length ? (100 * n) / lessons.length + "%" : "0";
}

async function showLesson(slug) {
  if (!lessons.length) return;
  slug = slug || store.get("lastLesson", lessons[0].slug);
  if (!lessons.find((l) => l.slug === slug)) slug = lessons[0].slug;
  store.set("lastLesson", slug);
  if (location.hash !== "#/learn/" + slug) history.replaceState(null, "", "#/learn/" + slug);

  const l = await api("/api/lessons/" + encodeURIComponent(slug));
  const idx = lessons.findIndex((x) => x.slug === slug);
  $("#lesson-eyebrow").textContent = [`Lesson ${l.order ?? idx}`, l.minutes && `${l.minutes} min`].filter(Boolean).join("  ·  ");
  $("#lesson-title").textContent = l.title;
  $("#lesson-summary").textContent = l.summary || "";
  $("#runbox").hidden = !l.run;
  $("#run-cmd").textContent = l.run || "";
  // Lesson HTML is rendered server-side from trusted course Markdown.
  $("#lesson-body").innerHTML = l.html;
  document.title = `${l.title} · rag-xray`;

  const cb = $("#lesson-done");
  cb.checked = done.has(slug);
  cb.onchange = () => {
    cb.checked ? done.add(slug) : done.delete(slug);
    store.set("done", [...done]);
    renderLessonList(slug);
    if (cb.checked) toast("Nice. Lesson complete.");
  };

  const prev = lessons[idx - 1], next = lessons[idx + 1];
  $("#pager").replaceChildren(
    prev ? el("a", { href: "#/learn/" + prev.slug, class: "prev" }, el("small", { text: "← Previous" }), prev.title) : el("span"),
    next ? el("a", { href: "#/learn/" + next.slug, class: "next" }, el("small", { text: "Next →" }), next.title) : el("span"),
  );

  const heads = [...$("#lesson-body").querySelectorAll("h2[id]")];
  $("#toc").replaceChildren(...(heads.length ? [el("h4", { text: "On this page" })] : []),
    ...heads.map((h) => el("a", { href: "#", text: h.textContent, onclick: (e) => { e.preventDefault(); h.scrollIntoView({ behavior: "smooth" }); } })));
  renderLessonList(slug);
}
$("#copy-run").addEventListener("click", () => copy($("#run-cmd").textContent));

// ---------------------------------------------------------------- info & documents
let info = null, pollTimer = null;

async function refreshInfo(data) {
  info = data || await api("/api/info");
  $("#models").replaceChildren(
    el("span", { class: "pill" }, "embed ", el("b", { text: info.embedModel })),
    el("span", { class: "pill" }, "chat ", el("b", { text: info.chatModel })),
  );
  $("#doc-list").replaceChildren(...info.docs.map((d) =>
    el("li", {},
      el("span", { class: "dot", style: `background:${docColor(d.name)}` }),
      el("span", { class: "name", text: d.name, title: d.name }),
      el("span", { class: "meta", text: info.ready ? `${d.chunks} ch` : "" }),
      el("button", { title: "Remove " + d.name, "aria-label": "Remove " + d.name, text: "×", onclick: () => removeDoc(d.name) })))
  );
  const st = $("#index-status");
  st.classList.toggle("error", !!info.indexError);
  if (info.building || (!info.ready && !info.indexError)) {
    st.replaceChildren(el("span", { class: "spin" }), "Indexing: chunking and embedding...");
    clearTimeout(pollTimer);
    pollTimer = setTimeout(() => refreshInfo(), 800);
  } else if (info.indexError) {
    st.textContent = "Index failed: " + info.indexError;
  } else {
    st.textContent = `${info.docs.length} docs → ${info.chunks} chunks, indexed in ${info.buildMs} ms`;
  }
  const size = $("#chunk-size");
  if (document.activeElement !== size) { size.value = info.maxChars; $("#chunk-size-out").textContent = info.maxChars; }
  renderExamples();
}

async function uploadFiles(files) {
  const fd = new FormData();
  [...files].forEach((f) => fd.append("files", f));
  $("#index-status").replaceChildren(el("span", { class: "spin" }), "Uploading and indexing...");
  try { await refreshInfo(await api("/api/docs", { method: "POST", body: fd })); toast(`Indexed ${files.length} file(s)`); }
  catch (e) { toast(e.message); refreshInfo(); }
}

async function removeDoc(name) {
  try { await refreshInfo(await api("/api/docs/" + encodeURIComponent(name), { method: "DELETE" })); } catch (e) { toast(e.message); }
}

const drop = $("#drop");
$("#file-input").addEventListener("change", (e) => e.target.files.length && uploadFiles(e.target.files));
["dragenter", "dragover"].forEach((t) => drop.addEventListener(t, (e) => { e.preventDefault(); drop.classList.add("over"); }));
["dragleave", "drop"].forEach((t) => drop.addEventListener(t, (e) => { e.preventDefault(); drop.classList.remove("over"); }));
drop.addEventListener("drop", (e) => e.dataTransfer.files.length && uploadFiles(e.dataTransfer.files));
$("#reset-docs").addEventListener("click", async () => {
  try { await refreshInfo(await api("/api/docs/reset", { method: "POST" })); toast("Back to the sample documents"); } catch (e) { toast(e.message); }
});

// ---------------------------------------------------------------- settings
const settings = { k: store.get("k", 4), mode: store.get("mode", "hybrid") };
$("#topk").value = settings.k;
$("#topk-out").textContent = settings.k;
$("#topk").addEventListener("input", (e) => { settings.k = +e.target.value; $("#topk-out").textContent = settings.k; store.set("k", settings.k); });
document.querySelectorAll("#mode button").forEach((b) => {
  b.classList.toggle("on", b.dataset.mode === settings.mode);
  b.setAttribute("aria-checked", b.dataset.mode === settings.mode);
  b.addEventListener("click", () => {
    settings.mode = b.dataset.mode;
    store.set("mode", settings.mode);
    document.querySelectorAll("#mode button").forEach((x) => { x.classList.toggle("on", x === b); x.setAttribute("aria-checked", x === b); });
  });
});
let sizeTimer;
$("#chunk-size").addEventListener("input", (e) => {
  $("#chunk-size-out").textContent = e.target.value;
  clearTimeout(sizeTimer);
  sizeTimer = setTimeout(async () => {
    $("#index-status").replaceChildren(el("span", { class: "spin" }), "Re-chunking and indexing...");
    try { await refreshInfo(await api("/api/index", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ maxChars: +e.target.value }) })); }
    catch (err) { toast(err.message); }
  }, 450);
});

// ---------------------------------------------------------------- examples
const sampleNames = ["expenses.md", "kestrel-api.md", "leave-policy.md", "office.md", "oncall.md", "security.md"];
const sampleQuestions = [
  "How many vacation days do I get?",
  "What does KST-503 mean?",
  "Vaultbox",
  "I'm ill and can't come to work",
  "What is the food budget per day when travelling abroad?",
  "What is the CEO's name?",
];
function renderExamples() {
  const isSample = info && info.docs.length === sampleNames.length && info.docs.every((d) => sampleNames.includes(d.name));
  const qs = isSample ? sampleQuestions : (info?.docs || []).slice(0, 4).map((d) => `Summarise ${d.name.replace(/\.(md|markdown|txt)$/i, "")}`);
  $("#examples").replaceChildren(...qs.map((q) => el("button", { class: "chip", type: "button", text: q, onclick: () => { $("#question").value = q; ask(); } })));
}

// ---------------------------------------------------------------- ask / x-ray
$("#ask-form").addEventListener("submit", (e) => { e.preventDefault(); ask(); });
$("#question").addEventListener("keydown", (e) => {
  if (e.key === "Enter" && !e.shiftKey) { e.preventDefault(); ask(); }
});

let lastPrompt = "";
$("#copy-prompt").addEventListener("click", () => copy(lastPrompt));

async function ask() {
  const question = $("#question").value.trim();
  if (!question) return;
  const noRag = $("#no-rag").checked;
  const btn = $("#ask-btn");
  btn.disabled = true;
  $("#xray-empty").hidden = true;
  $("#result").hidden = false;
  $("#step-retrieval").hidden = noRag;
  $("#columns").replaceChildren();
  $("#sources").replaceChildren();
  $("#prompt").replaceChildren();
  $("#timings").replaceChildren();
  const answer = $("#answer");
  answer.replaceChildren();
  answer.classList.add("typing");
  let text = "", final = [], retrievalMs = 0;

  try {
    const res = await fetch("/api/ask", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ question, k: settings.k, mode: settings.mode, noRag }),
    });
    if (!res.ok) throw new Error((await res.json().catch(() => ({}))).error || res.statusText);

    // Parse server-sent events from the streamed body.
    const reader = res.body.getReader();
    const dec = new TextDecoder();
    let buf = "";
    for (;;) {
      const { value, done: end } = await reader.read();
      if (end) break;
      buf += dec.decode(value, { stream: true });
      let i;
      while ((i = buf.indexOf("\n\n")) >= 0) {
        const raw = buf.slice(0, i);
        buf = buf.slice(i + 2);
        const event = (raw.match(/^event: (.*)$/m) || [])[1];
        const data = JSON.parse((raw.match(/^data: (.*)$/m) || [])[1] || "null");
        if (event === "retrieval") { final = data.final; retrievalMs = data.vectorMs + data.keywordMs; renderRetrieval(data); }
        else if (event === "prompt") renderPrompt(data);
        else if (event === "token") { text += data; renderAnswer(text, final); }
        else if (event === "done") {
          $("#timings").replaceChildren(
            ...(noRag ? [] : [el("span", { class: "pill", text: `retrieval ${retrievalMs} ms` })]),
            el("span", { class: "pill", text: `first token ${data.firstTokenMs} ms` }),
            el("span", { class: "pill", text: `total ${(data.totalMs / 1000).toFixed(1)} s` }),
          );
        } else if (event === "error") throw new Error(data.message);
      }
    }
  } catch (e) {
    answer.append(el("span", { class: "err", text: (text ? "\n" : "") + "Error: " + e.message }));
  } finally {
    answer.classList.remove("typing");
    btn.disabled = false;
  }
}

function renderAnswer(text, final) {
  // Turn [1], [2] citations into clickable badges that highlight the source.
  const parts = text.split(/(\[\d+\])/g);
  $("#answer").replaceChildren(...parts.map((p) => {
    const m = p.match(/^\[(\d+)\]$/);
    if (!m || !final[+m[1] - 1]) return p;
    return el("span", { class: "cite", text: m[1], title: final[+m[1] - 1].source, onclick: () => flashSource(+m[1]) });
  }));
  $("#sources").replaceChildren(...final.map((c, i) =>
    el("div", { class: "source", id: "src-" + (i + 1) },
      el("b", { text: `[${i + 1}]` }),
      el("span", {}, el("span", { text: c.source }), el("span", { class: "where", text: c.heading ? "  ›  " + c.heading : "" })))));
}

function flashSource(n) {
  const s = $("#src-" + n);
  if (!s) return;
  s.classList.add("flash");
  s.scrollIntoView({ behavior: "smooth", block: "nearest" });
  setTimeout(() => s.classList.remove("flash"), 1400);
}

function renderRetrieval(d) {
  const used = new Set(d.final.map((c) => c.id));
  const col = (key, title, explain, hits) => {
    const top = Math.max(...hits.map((h) => h.score), 1e-9);
    return el("div", { class: "col" + (d.mode === key ? " active" : "") },
      el("h4", {}, title),
      el("p", { class: "explain", text: explain }),
      hits.length
        ? el("ol", { class: "hits" }, ...hits.map((h, i) => {
            const moves = key === "hybrid" ? `vector #${h.vectorRank || "none"} · keyword #${h.keywordRank || "none"}` : "";
            const li = el("li", { class: "hit" + (d.mode === key && used.has(h.id) ? " used" : ""), style: `border-left:3px solid ${docColor(h.source)}` },
              el("span", { class: "rank", text: i + 1 }),
              el("div", { class: "src", text: h.source }),
              el("div", { class: "head", text: h.heading }),
              el("div", { class: "score" }, el("i", { style: `width:${Math.max(4, (60 * h.score) / top)}px` }), h.score.toFixed(key === "hybrid" ? 4 : 3)),
              moves ? el("div", { class: "moves", text: moves }) : null,
              el("div", { class: "text", text: h.text }));
            li.addEventListener("click", () => li.classList.toggle("open"));
            return li;
          }))
        : el("p", { class: "explain", text: "No results: no query word appears in any chunk." }));
  };
  $("#columns").replaceChildren(
    col("vector", "Vector", `Cosine similarity of meaning. ${d.vectorMs} ms incl. embedding the question.`, d.vector),
    col("keyword", "Keyword (BM25)", `Exact word matches, rare words weigh more. ${d.keywordMs} ms.`, d.keyword),
    col("hybrid", "Hybrid (RRF)", "Rank fusion of both lists: 1/(60+rank) summed.", d.hybrid),
  );
}

function renderPrompt(msgs) {
  lastPrompt = msgs.map((m) => `[${m.role}]\n${m.content}`).join("\n\n");
  $("#prompt").replaceChildren(...msgs.flatMap((m) => [el("span", { class: "role", text: m.role.toUpperCase() + "\n" }), m.content + "\n\n"]));
}

// ---------------------------------------------------------------- chunks view
let chunkCache = [];
async function loadChunks() {
  try { chunkCache = await api("/api/chunks"); } catch (e) { toast(e.message); }
  renderChunks();
}
function renderChunks() {
  const q = $("#chunk-filter").value.toLowerCase();
  const list = chunkCache.filter((c) => !q || (c.source + c.heading + c.text).toLowerCase().includes(q));
  const docs = new Set(chunkCache.map((c) => c.source)).size;
  const avg = chunkCache.length ? Math.round(chunkCache.reduce((s, c) => s + c.text.length, 0) / chunkCache.length) : 0;
  $("#chunks-summary").textContent = chunkCache.length
    ? `${chunkCache.length} chunks from ${docs} documents, ${avg} characters on average (max ${info?.maxChars}). Change the size in the X-ray settings.`
    : "No chunks yet: the index may still be building.";
  $("#chunk-grid").replaceChildren(...list.map((c) =>
    el("div", { class: "chunk", style: `--c:${docColor(c.source)}` },
      el("div", { class: "top" }, el("span", { text: c.source }), el("span", { text: `#${c.id} · ${c.text.length} ch` })),
      el("div", { class: "h", text: c.heading || "(no heading)" }),
      el("p", { text: c.text }))));
}
$("#chunk-filter").addEventListener("input", renderChunks);

// ---------------------------------------------------------------- boot
(async function boot() {
  try {
    lessons = await api("/api/lessons");
    renderLessonList();
  } catch (e) { toast("Could not load lessons: " + e.message); }
  refreshInfo().catch((e) => toast(e.message));
  if (!location.hash) history.replaceState(null, "", "#/learn");
  route();
})();
