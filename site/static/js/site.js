/* reqmd site — search, theme, copy buttons, mobile nav, focus trap */
(function () {
  "use strict";

  // ---- Theme toggle --------------------------------------------------------
  const themeBtn = document.querySelector("[data-theme-toggle]");
  if (themeBtn) {
    themeBtn.addEventListener("click", () => {
      // Read the user's *intent* from localStorage, not the resolved
      // value on the document (which may have been overwritten by the
      // "system" -> light/dark resolution in head.html).
      let cur;
      try { cur = localStorage.getItem("reqmd-theme") || "system"; } catch (e) { cur = "system"; }
      const next = cur === "system" ? "light" : cur === "light" ? "dark" : "system";
      if (next === "system") {
        const sys = window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
        document.documentElement.dataset.theme = sys;
      } else {
        document.documentElement.dataset.theme = next;
      }
      try { localStorage.setItem("reqmd-theme", next); } catch (e) {}
    });
  }

  // ---- Mobile nav drawer ---------------------------------------------------
  const navToggle = document.querySelector("[data-nav-toggle]");
  const primaryNav = document.querySelector(".primary-nav");
  if (navToggle && primaryNav) {
    navToggle.addEventListener("click", () => {
      const open = primaryNav.classList.toggle("is-open");
      navToggle.setAttribute("aria-expanded", open ? "true" : "false");
    });
    // close on link click
    primaryNav.addEventListener("click", (e) => {
      if (e.target.closest("a")) {
        primaryNav.classList.remove("is-open");
        navToggle.setAttribute("aria-expanded", "false");
      }
    });
  }

  // ---- Copy buttons --------------------------------------------------------
  document.addEventListener("click", (e) => {
    const btn = e.target.closest("[data-copy-target]");
    if (!btn) return;
    const id = btn.getAttribute("data-copy-target");
    const pre = document.getElementById(id);
    if (!pre) return;
    const text = pre.innerText;
    const done = () => {
      btn.classList.add("is-copied");
      const span = btn.querySelector("span");
      if (span) {
        const orig = span.textContent;
        span.textContent = "Copied";
        setTimeout(() => { span.textContent = orig; btn.classList.remove("is-copied"); }, 1200);
      }
    };
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(text).then(done);
    } else {
      const ta = document.createElement("textarea");
      ta.value = text; document.body.appendChild(ta);
      ta.select();
      try { document.execCommand("copy"); done(); } catch (e) {}
      document.body.removeChild(ta);
    }
  });

  // ---- Mermaid auto-init ---------------------------------------------------
  function initMermaid() {
    if (!window.mermaid) return;
    const dark = document.documentElement.dataset.theme === "dark"
      || (document.documentElement.dataset.theme === "system" && window.matchMedia("(prefers-color-scheme: dark)").matches);
    window.mermaid.initialize({
      startOnLoad: true,
      theme: dark ? "dark" : "neutral",
      securityLevel: "strict",
      themeVariables: {
        fontFamily: 'ui-monospace, "JetBrains Mono", "SF Mono", monospace',
        fontSize: "14px",
      },
    });
  }
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", initMermaid);
  } else {
    initMermaid();
  }
  setTimeout(initMermaid, 250);

  // ---- TOC active-section highlight via IntersectionObserver --------------
  if ("IntersectionObserver" in window) {
    const tocLinks = document.querySelectorAll("#TableOfContents a[href^='#']");
    if (tocLinks.length > 0) {
      const ids = Array.from(tocLinks).map(a => a.getAttribute("href").slice(1));
      const headings = ids.map(id => document.getElementById(id)).filter(Boolean);
      if (headings.length > 0) {
        const obs = new IntersectionObserver((entries) => {
          for (const entry of entries) {
            if (entry.isIntersecting) {
              tocLinks.forEach((a) => a.classList.remove("active"));
              const active = document.querySelector("#TableOfContents a[href='#" + entry.target.id + "']");
              if (active) active.classList.add("active");
            }
          }
        }, { rootMargin: "-30% 0% -60% 0%" });
        headings.forEach(h => obs.observe(h));
      }
    }
  }

  // ---- Search --------------------------------------------------------------
  const modal = document.querySelector("[data-search-modal]");
  if (!modal) return;

  const backdrop = modal.querySelector("[data-search-backdrop]");
  const openBtn = document.querySelector("[data-search-open]");
  const input = modal.querySelector("[data-search-input]");
  const results = modal.querySelector("[data-search-results]");
  const panel = modal.querySelector(".search-panel");

  let fuse = null;
  let indexLoaded = false;
  let focusIdx = -1;
  let currentResults = [];
  let lastFocused = null;

  function loadIndex() {
    if (indexLoaded) return Promise.resolve();
    const inline = document.getElementById("search-index");
    if (inline && inline.textContent.trim()) {
      try {
        // Hugo's jsonify wraps a JSON string with quotes; parse twice to unwrap.
        const data = JSON.parse(JSON.parse(inline.textContent));
        indexLoaded = true;
        return Promise.resolve(data);
      } catch (e) { /* fall through to fetch */ }
    }
    return fetch("/index.json", { credentials: "same-origin" })
      .then(r => r.ok ? r.json() : Promise.reject(new Error("index.json " + r.status)))
      .catch(() => []);
  }

  const fuseOpts = {
    keys: [
      { name: "title", weight: 0.5 },
      { name: "summary", weight: 0.3 },
      { name: "content", weight: 0.2 },
    ],
    threshold: 0.32,
    ignoreLocation: true,
    minMatchCharLength: 2,
    includeMatches: true,
    includeScore: true,
  };

  function getFuse(data) {
    if (fuse) return fuse;
    const Fuse = window.Fuse;
    if (!Fuse) return null;
    fuse = new Fuse(data, fuseOpts);
    return fuse;
  }

  function escapeHtml(s) {
    return String(s).replace(/[&<>"']/g, c => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", "\"": "&quot;", "'": "&#39;" }[c]));
  }

  function highlight(text, matches) {
    if (!matches || !matches.length) return escapeHtml(text);
    const indices = matches
      .filter(m => m.key === "title" || m.key === "summary" || m.key === "content")
      .flatMap(m => m.indices)
      .filter(([s, e]) => e - s >= 1)
      .sort((a, b) => a[0] - b[0]);
    if (!indices.length) return escapeHtml(text);
    let out = "";
    let last = 0;
    for (const [s, e] of indices) {
      if (s < last) continue;
      out += escapeHtml(text.slice(last, s));
      out += "<mark>" + escapeHtml(text.slice(s, e + 1)) + "</mark>";
      last = e + 1;
    }
    out += escapeHtml(text.slice(last));
    return out;
  }

  function snippet(text, matches, maxLen) {
    maxLen = maxLen || 160;
    if (!matches || !matches.length) return escapeHtml(text.slice(0, maxLen)) + (text.length > maxLen ? "…" : "");
    const m = matches.find(m => m.key === "content" || m.key === "summary");
    if (!m || !m.indices.length) return escapeHtml(text.slice(0, maxLen)) + (text.length > maxLen ? "…" : "");
    const start = Math.max(0, (m.value || "").indexOf(m.indices[0][0]) - 40);
    const slice = (m.value || "").slice(start, start + maxLen);
    return highlight(slice, [{ key: m.key, indices: m.indices.map(([s, e]) => [s - start, e - start]).filter(([s, e]) => s >= 0 && e < maxLen) }]);
  }

  function renderResults(rs) {
    const total = rs.length;
    if (!total) {
      results.innerHTML = '<div class="search-empty">No results. Try a different query.</div>';
      focusIdx = -1;
      currentResults = [];
      return;
    }
    currentResults = rs.slice(0, 12);
    focusIdx = 0;
    const showing = currentResults.length;
    const countLine = total > showing
      ? '<div class="search-count">Showing ' + showing + ' of ' + total + ' results</div>'
      : '';
    results.innerHTML = countLine + currentResults.map((r, i) => {
      const item = r.item;
      const titleHtml = highlight(item.title, r.matches);
      const summaryHtml = snippet(item.summary || item.content || "", r.matches);
      return '<a class="search-result' + (i === 0 ? " is-focused" : "") + '" href="' + escapeHtml(item.url) + '" data-result-idx="' + i + '">' +
        '<div class="search-result-section">' + escapeHtml(item.section || "") + '</div>' +
        '<div class="search-result-title">' + titleHtml + '</div>' +
        '<div class="search-result-summary">' + summaryHtml + '</div>' +
        '</a>';
    }).join("");
  }

  function openSearch() {
    lastFocused = document.activeElement;
    modal.hidden = false;
    modal.setAttribute("aria-hidden", "false");
    document.body.style.overflow = "hidden";
    input.focus();
    if (!indexLoaded) {
      results.innerHTML = '<div class="search-empty">Loading index…</div>';
      loadIndex().then(data => {
        indexLoaded = true;
        getFuse(data);
        if (!fuse) {
          results.innerHTML = '<div class="search-empty">Search index unavailable (Fuse.js not loaded).</div>';
        } else if (!input.value.trim()) {
          results.innerHTML = '<div class="search-empty">Type to search the docs.</div>';
        } else {
          renderResults(fuse.search(input.value.trim()));
        }
      });
    } else if (input.value.trim() && fuse) {
      renderResults(fuse.search(input.value.trim()));
    } else {
      results.innerHTML = '<div class="search-empty">Type to search the docs.</div>';
    }
  }

  function closeSearch() {
    modal.hidden = true;
    modal.setAttribute("aria-hidden", "true");
    document.body.style.overflow = "";
    input.value = "";
    results.innerHTML = "";
    focusIdx = -1;
    if (lastFocused && typeof lastFocused.focus === "function") {
      lastFocused.focus();
    }
  }

  if (openBtn) openBtn.addEventListener("click", openSearch);
  if (backdrop) backdrop.addEventListener("click", closeSearch);

  document.addEventListener("keydown", (e) => {
    if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "k") {
      e.preventDefault();
      if (modal.hidden) openSearch(); else closeSearch();
      return;
    }
    if (e.key === "Escape" && !modal.hidden) {
      closeSearch();
      return;
    }
    if (!modal.hidden) {
      if (e.key === "ArrowDown") { e.preventDefault(); focusIdx = Math.min(focusIdx + 1, currentResults.length - 1); updateFocus(); }
      if (e.key === "ArrowUp") { e.preventDefault(); focusIdx = Math.max(focusIdx - 1, 0); updateFocus(); }
      if (e.key === "Enter" && focusIdx >= 0 && currentResults[focusIdx]) {
        e.preventDefault();
        window.location.href = currentResults[focusIdx].item.url;
      }
    }
  });

  // Focus trap inside the search modal
  document.addEventListener("focusin", (e) => {
    if (modal.hidden) return;
    if (!panel.contains(e.target)) {
      // pull focus back to input
      e.stopPropagation();
      input.focus();
    }
  });

  function updateFocus() {
    results.querySelectorAll(".search-result").forEach((el, i) => {
      el.classList.toggle("is-focused", i === focusIdx);
      if (i === focusIdx) el.scrollIntoView({ block: "nearest" });
    });
  }

if (input) {
  input.addEventListener("input", () => {
    if (window.Fuse && !fuse) {
      // Try to re-init fuse if it failed earlier (e.g. Fuse script loaded after our code)
      loadIndex().then(data => { fuse = new window.Fuse(data, fuseOpts); renderResults(fuse.search(input.value.trim())); });
      return;
    }
    if (!fuse) return;
    const q = input.value.trim();
    if (!q) { results.innerHTML = '<div class="search-empty">Type to search the docs.</div>'; currentResults = []; focusIdx = -1; return; }
    renderResults(fuse.search(q));
  });
}
})();
