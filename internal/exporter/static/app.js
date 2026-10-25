// reqmd HTML export client.
//
// The Document view is server-rendered and remains fully readable when this
// script fails or the page is opened offline. This controller adds: one
// filtering pipeline over the shared page model, three analytical views that
// render on demand (table, trace, coverage), authoritative URL state, and the
// palette/shortcut/dialog behaviour.
(function () {
  'use strict';

  var COVERAGE_LABELS = {
    'covered': 'covered',
    'missing-coverage': 'missing coverage',
    'unresolved': 'unresolved source',
    'not-required': 'not required'
  };
  var VERDICT_ORDER = ['fail', 'inconclusive', 'skipped', 'pass'];

  function readJSON(id) {
    var el = document.getElementById(id);
    if (!el) return null;
    try { return JSON.parse(el.textContent); } catch (e) { return null; }
  }

  function el(tag, props, children) {
    var node = document.createElement(tag);
    if (props) {
      Object.keys(props).forEach(function (k) {
        var v = props[k];
        if (v === null || v === undefined || v === false) return;
        if (k === 'text') { node.textContent = v; return; }
        if (k === 'className') { node.className = v; return; }
        node.setAttribute(k, v === true ? '' : v);
      });
    }
    (children || []).forEach(function (c) {
      if (c === null || c === undefined) return;
      node.appendChild(typeof c === 'string' ? document.createTextNode(c) : c);
    });
    return node;
  }

  function clear(node) { while (node.firstChild) node.removeChild(node.firstChild); }

  function attrValues(node) {
    if (!node.attrs) return [];
    var out = [];
    Object.keys(node.attrs).forEach(function (k) {
      var v = node.attrs[k];
      if (Array.isArray(v)) {
        v.forEach(function (item) {
          if (item !== null && typeof item !== 'object') out.push(String(item));
        });
      } else if (v !== null && typeof v === 'object') {
        // nested objects are not searchable text
      } else if (v !== undefined && v !== null) {
        out.push(String(v));
      }
    });
    return out;
  }

  // ── shareable URL state (pure; exercised by url_state.test.js) ──
  // Filters are repeated `f=attr:value` parameters. Only the first colon is
  // structural, so values may contain colons, commas, and any Unicode text.
  function decodeFilters(pairs) {
    var filters = {};
    (pairs || []).forEach(function (pair) {
      var i = pair.indexOf(':');
      if (i <= 0) return;
      var attr = pair.slice(0, i), value = pair.slice(i + 1);
      if (!value) return;
      if (!filters[attr]) filters[attr] = [];
      if (filters[attr].indexOf(value) === -1) filters[attr].push(value);
    });
    return filters;
  }

  function encodeFilters(filters) {
    var pairs = [];
    Object.keys(filters).sort().forEach(function (attr) {
      (filters[attr] || []).forEach(function (value) { pairs.push(attr + ':' + value); });
    });
    return pairs;
  }

  // An explicitly cleared filter set must round-trip as `f=`; otherwise the
  // document default (usually status=approved) would come back on reload.
  function encodeUrlState(s) {
    var p = new URLSearchParams();
    if (s.query) p.set('q', s.query);
    var pairs = encodeFilters(s.filters);
    if (pairs.length) pairs.forEach(function (pair) { p.append('f', pair); });
    else if (s.explicitFilters) p.set('f', '');
    if (s.view && s.view !== 'document') p.set('view', s.view);
    if (s.density === 'compact') p.set('density', 'compact');
    if (s.view === 'trace' && s.selected) p.set('req', s.selected);
    return p.toString();
  }

  function decodeUrlState(search) {
    var p = new URLSearchParams(search || '');
    var out = {
      query: p.get('q') || '',
      filters: {},
      explicitFilters: false,
      view: 'document',
      density: 'comfortable',
      selected: p.get('req') || null
    };
    if (p.has('f')) {
      out.explicitFilters = true;
      out.filters = decodeFilters(p.getAll('f'));
    }
    var view = p.get('view');
    if (['document', 'table', 'trace', 'coverage'].indexOf(view) !== -1) out.view = view;
    if (p.get('density') === 'compact') out.density = 'compact';
    return out;
  }

  // ── platform hint ──
  // The palette is ⌘K on macOS and Ctrl+K elsewhere; label the button the
  // way the reader's keyboard actually spells it.
  var IS_MAC = /Mac|iPhone|iPad|iPod/.test(navigator.platform || '');
  var MOD_LABEL = IS_MAC ? '\u2318K' : 'Ctrl K';

  function setupModHints() {
    Array.prototype.forEach.call(document.querySelectorAll('[data-mod-hint]'), function (b) {
      b.textContent = MOD_LABEL;
      if (b.getAttribute('data-mod-hint') === 'palette') {
        b.title = 'Command palette (' + MOD_LABEL + ')';
      }
    });
    // The shortcuts dialog spells the modifier the way the platform does too.
    Array.prototype.forEach.call(document.querySelectorAll('[data-mod-keys]'), function (dt) {
      dt.textContent = '';
      (IS_MAC ? ['\u2318'] : ['Ctrl']).concat('K').forEach(function (k) {
        var kb = document.createElement('kbd');
        kb.textContent = k;
        dt.appendChild(kb);
      });
    });
  }

  // ── theme ──
  function setupTheme() {
    var buttons = document.querySelectorAll('[data-action="theme"]');
    function apply() {
      var dark = document.documentElement.classList.contains('dark');
      Array.prototype.forEach.call(buttons, function (b) {
        b.setAttribute('aria-pressed', dark ? 'true' : 'false');
        b.textContent = dark ? '☀' : '☾';
      });
    }
    Array.prototype.forEach.call(buttons, function (b) {
      b.addEventListener('click', function () {
        var dark = !document.documentElement.classList.contains('dark');
        document.documentElement.classList.toggle('dark', dark);
        try { localStorage.setItem('reqmd-theme', dark ? 'dark' : 'light'); } catch (e) { /* private mode */ }
        apply();
      });
    });
    apply();
  }

  // ── focus trap ──
  var FOCUSABLE = 'a[href],button:not([disabled]),input,select,textarea,[tabindex]:not([tabindex="-1"])';

  function trapFocus(panel, e) {
    if (e.key !== 'Tab') return;
    var items = panel.querySelectorAll(FOCUSABLE);
    if (!items.length) return;
    var first = items[0], last = items[items.length - 1];
    if (e.shiftKey && document.activeElement === first) { e.preventDefault(); last.focus(); }
    else if (!e.shiftKey && document.activeElement === last) { e.preventDefault(); first.focus(); }
  }

  // ── document page ──
  function initDocumentPage() {
    var model = readJSON('req-model');
    if (!model) return;

    var nodes = model.nodes || [];
    var byId = {};
    var reqNodes = [];
    nodes.forEach(function (n) {
      byId[n.id] = n;
      if (!n.type) reqNodes.push(n);
    });

    var state = {
      view: 'document',
      density: 'comfortable',
      query: '',
      filters: {},
      explicitFilters: false,
      selected: null,
      trail: [],
      matrix: false,
      tableSort: { key: 'id', dir: 'asc' },
      tableCols: null,
      openRows: {}
    };
    var visibleSet = new Set();
    var bodyTextCache = {};

    var views = document.querySelectorAll('[data-view-panel]');
    function panel(name) {
      for (var i = 0; i < views.length; i++) {
        if (views[i].getAttribute('data-view-panel') === name) return views[i];
      }
      return null;
    }
    var documentPanel = panel('document');
    var tablePanel = panel('table');
    var tracePanel = panel('trace');
    var coveragePanel = panel('coverage');
    var countEl = document.querySelector('[data-count]');
    var tocCountEl = document.querySelector('[data-toc-count]');
    var searchEl = document.querySelector('[data-search]');
    var chipsEl = document.querySelector('[data-active-filters]');
    var filtersEl = document.querySelector('[data-filters]');
    var overlays = {
      palette: document.getElementById('palette'),
      shortcuts: document.getElementById('shortcuts')
    };
    var paletteInput = document.querySelector('[data-palette-input]');
    var paletteList = document.querySelector('[data-palette-list]');
    var paletteEmpty = document.querySelector('[data-palette-empty]');
    var paletteIndex = 0;
    var lastFocus = null;

    if (filtersEl && filtersEl.getAttribute('data-default-status')) {
      state.defaultStatus = filtersEl.getAttribute('data-default-status');
    } else {
      state.defaultStatus = '';
    }

    // ── filtering ──
    function bodyText(id) {
      if (bodyTextCache[id] !== undefined) return bodyTextCache[id];
      var card = document.getElementById(id);
      var text = card ? (card.textContent || '').toLowerCase() : '';
      bodyTextCache[id] = text;
      return text;
    }

    function matches(node) {
      var q = state.query.trim().toLowerCase();
      if (q) {
        var hay = node.id + ' ' + (node.title || '') + ' ' + attrValues(node).join(' ');
        if (hay.toLowerCase().indexOf(q) === -1 && bodyText(node.id).indexOf(q) === -1) return false;
      }
      var names = Object.keys(state.filters);
      for (var i = 0; i < names.length; i++) {
        var attr = names[i];
        var selected = state.filters[attr];
        if (!selected || !selected.length) continue;
        if (attr === 'state.coverage') {
          if (selected.indexOf(node.coverage || '') === -1) return false;
          continue;
        }
        if (attr === 'state.verdict') {
          if (selected.indexOf(node.verdict || 'none') === -1) return false;
          continue;
        }
        var value = node.attrs ? node.attrs[attr] : undefined;
        var hit = false;
        if (value !== undefined && value !== null) {
          if (Array.isArray(value)) {
            hit = value.some(function (v) { return selected.indexOf(String(v)) !== -1; });
          } else {
            hit = selected.indexOf(String(value)) !== -1;
          }
        }
        if (!hit) return false;
      }
      return true;
    }

    function computeVisible() {
      visibleSet = new Set();
      reqNodes.forEach(function (n) { if (matches(n)) visibleSet.add(n.id); });
    }

    function visibleRequirements() {
      return reqNodes.filter(function (n) { return visibleSet.has(n.id); });
    }

    // ── filter controls ──
    function syncFilterInputs() {
      var inputs = document.querySelectorAll('[data-filter-attr]');
      Array.prototype.forEach.call(inputs, function (input) {
        var attr = input.getAttribute('data-filter-attr');
        var values = state.filters[attr] || [];
        input.checked = values.indexOf(input.value) !== -1;
      });
      var counts = document.querySelectorAll('[data-filter-count]');
      Array.prototype.forEach.call(counts, function (span) {
        var attr = span.getAttribute('data-filter-count');
        var n = (state.filters[attr] || []).length;
        span.textContent = n ? '(' + n + ')' : '';
      });
    }

    function toggleFilter(attr, value, on) {
      var values = state.filters[attr] || [];
      var i = values.indexOf(value);
      if (on && i === -1) values.push(value);
      if (!on && i !== -1) values.splice(i, 1);
      if (values.length) state.filters[attr] = values;
      else delete state.filters[attr];
      state.explicitFilters = true;
      apply();
    }

    function filterPairs() {
      return encodeFilters(state.filters);
    }

    function setState(attr, value) {
      // Keep the existing filters: the exception lists are computed over the
      // visible set, so applying a state filter must not widen it.
      state.filters[attr] = [value];
      state.explicitFilters = true;
      state.query = '';
      if (searchEl) searchEl.value = '';
      setView('document');
      apply();
    }

    // ── URL state ──
    function writeState() {
      var qs = encodeUrlState(state);
      try {
        history.replaceState(null, '', location.pathname + (qs ? '?' + qs : '') + location.hash);
      } catch (e) { /* file:// or sandboxed */ }
    }

    function restoreState() {
      var decoded;
      try {
        decoded = decodeUrlState(location.search);
      } catch (e) {
        return;
      }
      state.query = decoded.query;
      state.filters = decoded.filters;
      state.explicitFilters = decoded.explicitFilters;
      state.view = decoded.view;
      state.density = decoded.density;
      if (decoded.selected) state.selected = decoded.selected;
      if (searchEl) searchEl.value = state.query;
      if (!state.explicitFilters && state.defaultStatus && model.filters &&
          model.filters.status && model.filters.status.indexOf(state.defaultStatus) !== -1) {
        state.filters.status = [state.defaultStatus];
      }
    }

    // ── view switching ──
    function setView(name) {
      state.view = name;
      var buttons = document.querySelectorAll('.viewswitch [data-view]');
      Array.prototype.forEach.call(buttons, function (b) {
        var on = b.getAttribute('data-view') === name;
        b.classList.toggle('is-active', on);
        b.setAttribute('aria-pressed', on ? 'true' : 'false');
      });
      Array.prototype.forEach.call(views, function (v) {
        v.hidden = v.getAttribute('data-view-panel') !== name;
      });
      document.body.setAttribute('data-view', name);
      if (name === 'table') renderTable();
      if (name === 'trace') renderTrace();
      if (name === 'coverage') renderCoverage();
      if (name !== 'document') clearScrollSpy();
      syncOutlineSelection();
      if (name === 'document') revealSelectedInDocument();
      writeState();
    }

    function apply() {
      computeVisible();
      syncFilterInputs();
      var cardEls = documentPanel.querySelectorAll('.req');
      Array.prototype.forEach.call(cardEls, function (card) {
        var id = card.getAttribute('data-id');
        card.classList.toggle('is-filtered-out', !visibleSet.has(id));
      });
      syncTOC();
      var visible = visibleSet.size;
      var total = reqNodes.length;
      if (countEl) countEl.textContent = countLabel(visible, total);
      if (tocCountEl) tocCountEl.textContent = visible === total ? String(total) : visible + '/' + total;
      renderChips();
      if (state.view === 'table') renderTable();
      if (state.view === 'trace') renderTrace();
      if (state.view === 'coverage') renderCoverage();
      writeState();
    }

    // countLabel states both the visible and total requirements, and names the
    // default filter when it is the reason some are hidden.
    function countLabel(visible, total) {
      var hidden = total - visible;
      if (!hidden) return visible + ' shown · ' + total + ' total';
      var reason = '';
      if (state.filters.status && state.filters.status.length) {
        reason = 'status: ' + state.filters.status.join(', ');
      }
      return visible + ' shown · ' + total + ' total (' + hidden + ' hidden' +
        (reason ? ' by ' + reason : '') + ')';
    }

    function renderChips() {
      if (!chipsEl) return;
      clear(chipsEl);
      var pairs = filterPairs();
      if (!pairs.length) { chipsEl.hidden = true; return; }
      chipsEl.hidden = false;
      pairs.forEach(function (pair) {
        var i = pair.indexOf(':');
        var attr = pair.slice(0, i), value = pair.slice(i + 1);
        var chip = el('span', { className: 'chip chip--filter' }, [
          el('span', { text: labelFor(attr) + ': ' + value }),
          el('button', { type: 'button', className: 'chip__remove', 'aria-label': 'Remove filter ' + labelFor(attr) + ' ' + value, text: '×' })
        ]);
        chip.querySelector('button').addEventListener('click', function () {
          toggleFilter(attr, value, false);
        });
        chipsEl.appendChild(chip);
      });
      var clearBtn = el('button', { type: 'button', className: 'linkbtn', text: 'Clear all' });
      clearBtn.addEventListener('click', function () {
        state.filters = {};
        state.query = '';
        if (searchEl) searchEl.value = '';
        state.explicitFilters = true;
        apply();
      });
      chipsEl.appendChild(clearBtn);
    }

    function labelFor(attr) {
      if (attr === 'state.coverage') return 'Coverage';
      if (attr === 'state.verdict') return 'Verdict';
      return attr.replace(/[-_]/g, ' ').replace(/^./, function (c) { return c.toUpperCase(); });
    }

    // ── TOC filtering + scroll spy ──
    function syncTOC() {
      var items = document.querySelectorAll('[data-toc-for]');
      function visibleFor(item) {
        var id = item.getAttribute('data-toc-for');
        if (byId[id] && !byId[id].type) return visibleSet.has(id);
        var nested = item.querySelectorAll('[data-toc-for]');
        var hasRequirements = false;
        for (var i = 0; i < nested.length; i++) {
          var nid = nested[i].getAttribute('data-toc-for');
          if (byId[nid] && !byId[nid].type) {
            hasRequirements = true;
            if (visibleSet.has(nid)) return true;
          }
        }
        // Leaf items (prose blocks) are structural: always visible.
        return !hasRequirements;
      }
      Array.prototype.forEach.call(items, function (item) {
        item.classList.toggle('is-filtered-out', !visibleFor(item));
      });
    }

    // The outline is the requirement list for every view: in the Trace and
    // Coverage views a click selects that requirement, in the reading views it
    // scrolls to it.
    function setupOutlineClicks() {
      document.querySelectorAll('.toc__link').forEach(function (a) {
        a.addEventListener('click', function (e) {
          var id = a.getAttribute('href').slice(1);
          if (!byId[id] || byId[id].type) return; // items keep plain anchors
          if (state.view === 'document') {
            // Plain anchor navigation, but record the selection so it carries
            // over when the reader switches to the analytical views.
            selectQuietly(id);
            return;
          }
          if (state.view === 'trace') {
            e.preventDefault();
            selectNode(id);
            return;
          }
          if (state.view === 'table') {
            // The table already contains the requirement: select it and bring
            // the row into view instead of switching to the reading view.
            e.preventDefault();
            revealInTable(id);
            return;
          }
          if (state.view === 'coverage') {
            e.preventDefault();
            jumpTo(id, true);
          }
        });
      });
    }

    // Mark the outline entry that the active view is centred on.
    // syncOutlineSelection marks the requirement the user selected. It is the
    // same requirement in every view, so switching views keeps the highlight.
    function syncOutlineSelection() {
      var selected = state.selected;
      document.querySelectorAll('.toc__item').forEach(function (item) {
        var id = item.getAttribute('data-toc-for');
        item.classList.toggle('is-selected', !!selected && id === selected && !!(byId[id] && !byId[id].type));
      });
    }

    // The reading views follow the scroll position; the Trace view is driven
    // by the selected requirement instead, so scrolling must not steal the
    // highlight there.
    function scrollSpyActive() {
      return state.view === 'document';
    }

    function setupScrollSpy() {
      if (!window.IntersectionObserver) return;
      var links = {};
      document.querySelectorAll('.toc__link').forEach(function (a) {
        var id = a.getAttribute('href').slice(1);
        links[id] = a;
      });
      var observer = new IntersectionObserver(function (entries) {
        if (!scrollSpyActive() || scrollSpySuspended) return;
        entries.forEach(function (entry) {
          if (!entry.isIntersecting) return;
          Object.keys(links).forEach(function (id) { links[id].classList.remove('is-active'); });
          var link = links[entry.target.id];
          if (!link) return;
          link.classList.add('is-active');
          var item = link.closest('[data-toc-for]');
          while (item) {
            var parent = item.parentElement ? item.parentElement.closest('[data-toc-for]') : null;
            if (parent && parent !== item) parent.querySelector('.toc__link').classList.add('is-active');
            item = parent;
          }
        });
      }, { rootMargin: '-72px 0px -70% 0px' });
      document.querySelectorAll('.req, .item').forEach(function (n) { observer.observe(n); });
    }

    // scrollSpySuspended is set while a programmatic reveal is running, so the
    // observer cannot overwrite the highlight of the requirement the reader
    // just selected.
    var scrollSpySuspended = false;

    // clearScrollSpy removes the position highlight when a view stops
    // following the scroll position.
    function clearScrollSpy() {
      document.querySelectorAll('.toc__link.is-active').forEach(function (a) {
        a.classList.remove('is-active');
      });
    }

    // ── table view ──
    function attributeColumns() {
      return model.attributeColumns || [];
    }

    function defaultColumns() {
      var cols = { status: true, coverage: true, verdict: true, up: true, down: true };
      attributeColumns().slice(0, 3).forEach(function (a) { cols['attr:' + a] = true; });
      return cols;
    }

    if (!state.tableCols) state.tableCols = defaultColumns();

    function columnDefs() {
      var defs = [
        { key: 'id', label: 'ID', locked: true },
        { key: 'title', label: 'Title', locked: true, wide: true }
      ];
      if (model.statusEnabled) defs.push({ key: 'status', label: 'Status' });
      defs = defs.concat([
        { key: 'coverage', label: 'Coverage' },
        { key: 'verdict', label: 'Verdict' },
        { key: 'up', label: 'Up', title: 'Upstream trace links', numeric: true },
        { key: 'down', label: 'Down', title: 'Downstream trace links', numeric: true }
      ]);
      attributeColumns().forEach(function (a) {
        defs.push({ key: 'attr:' + a, label: labelFor(a), attr: a });
      });
      return defs;
    }

    function cellValue(node, key) {
      switch (key) {
        case 'id': return node.id;
        case 'title': return node.title || '';
        case 'status': return node.attrs && node.attrs.status ? String(node.attrs.status) : 'approved';
        case 'coverage': return node.coverage || '';
        case 'verdict': return node.verdict || '';
        case 'up': return (node.up || []).length;
        case 'down': return (node.down || []).length;
        default: {
          var name = key.slice(5);
          var v = node.attrs ? node.attrs[name] : undefined;
          if (Array.isArray(v)) return v.join(', ');
          if (v === null || v === undefined) return '';
          if (typeof v === 'object') return '';
          return String(v);
        }
      }
    }

    // withEmptyColumnsHidden drops optional columns that carry a value in
    // fewer than half the visible rows (for example Verdict when no results
    // were exported, or an attribute set on one requirement out of twenty).
    // A mostly-blank column looks like a bug and costs width; the column
    // chooser still lists such columns, so selecting one brings it back.
    function withEmptyColumnsHidden(defs, nodes) {
      if (!nodes.length) return defs;
      return defs.filter(function (d) {
        if (d.locked) return true;
        var filled = 0;
        nodes.forEach(function (n) {
          var v = cellValue(n, d.key);
          if (v !== '' && v !== null && v !== undefined) filled++;
        });
        return filled * 2 >= nodes.length;
      });
    }

    function renderTable() {
      if (!tablePanel) return;
      clear(tablePanel);
      var visible = visibleRequirements();
      var defs = withEmptyColumnsHidden(columnDefs().filter(function (d) {
        return d.locked || state.tableCols[d.key];
      }), visible);
      var sort = state.tableSort;

      var head = el('div', { className: 'view-head' }, [
        el('h2', { text: 'Table' }),
        el('p', { className: 'view-sub', text: 'Sort by any column, choose columns, expand a row for body and evidence.' + (model.hasResults ? '' : ' No verification results are loaded, so every verdict shows as a dash.') })
      ]);
      var chooser = buildColumnChooser(defs);
      var toolbar = el('div', { className: 'view-tools' }, [chooser]);
      tablePanel.appendChild(head);
      tablePanel.appendChild(toolbar);

      if (!visible.length) {
        tablePanel.appendChild(el('div', { className: 'empty' }, [
          el('p', { className: 'empty__title', text: 'No requirements match the current filters' }),
          el('p', { text: 'Clear the search or filters to see rows.' })
        ]));
        return;
      }

      var table = el('table', { className: 'grid', id: 'req-table' });
      var thead = el('thead', {}, [el('tr', {})]);
      var headRow = thead.firstChild;
      headRow.appendChild(el('th', { scope: 'col', className: 'cell-detail' }, [
        el('span', { className: 'visually-hidden', text: 'Details' })
      ]));
      defs.forEach(function (d) {
        var active = sort.key === d.key;
        var th = el('th', {
          scope: 'col',
          className: 'sortable' + (d.numeric ? ' num' : '') + (active ? ' is-sorted' : ''),
          title: d.title || '',
          'aria-sort': active ? (sort.dir === 'asc' ? 'ascending' : 'descending') : 'none'
        }, [el('span', { text: d.label }), el('span', { className: 'sort-ind', text: active ? (sort.dir === 'asc' ? ' ↑' : ' ↓') : '' })]);
        th.addEventListener('click', function () {
          if (sort.key === d.key) sort.dir = sort.dir === 'asc' ? 'desc' : 'asc';
          else { sort.key = d.key; sort.dir = 'asc'; }
          renderTable();
        });
        headRow.appendChild(th);
      });
      table.appendChild(thead);

      var rows = visible.slice().sort(function (a, b) { return compareRows(a, b, sort); });
      var tbody = el('tbody', {});
      rows.forEach(function (node) {
        var tr = el('tr', { className: 'grid__row', 'data-id': node.id }, [
          el('td', { className: 'cell-detail' }, [el('button', {
            type: 'button', className: 'row-toggle', 'aria-expanded': state.openRows[node.id] ? 'true' : 'false',
            'aria-label': 'Toggle details for ' + node.id, text: state.openRows[node.id] ? '▾ Details' : '▸ Details'
          })])
        ]);
        tr.addEventListener('click', function (e) {
          if (e.target.closest('a')) return;
          revealInTable(node.id);
        });
        tr.firstChild.firstChild.addEventListener('click', function () {
          state.openRows[node.id] = !state.openRows[node.id];
          renderTable();
        });
        defs.forEach(function (d) {
          var value = cellValue(node, d.key);
          var td;
          if (d.key === 'id') {
            td = el('td', { className: 'cell-id' }, [el('a', { href: '#' + node.id, 'data-open-req': node.id }, [el('code', { text: node.id })])]);
          } else if (d.key === 'status' || d.key === 'coverage' || d.key === 'verdict') {
            td = el('td', {}, [stateBadge(d.key, value)]);
          } else {
            td = el('td', { className: d.numeric ? 'num' : '', text: value });
          }
          tr.appendChild(td);
        });
        tbody.appendChild(tr);

        if (state.openRows[node.id]) {
          var detail = detailRow(node, defs.length + 1);
          tbody.appendChild(detail);
        }
      });
      table.appendChild(tbody);
      tablePanel.appendChild(el('div', { className: 'table-scroll' }, [table]));
    }

    function stateBadge(kind, value) {
      if (!value) return el('span', { className: 'muted', text: '—' });
      var label = kind === 'coverage' ? (COVERAGE_LABELS[value] || value) : value;
      return el('span', { className: 'chip chip--' + value, text: label });
    }

    function compareRows(a, b, sort) {
      var va = cellValue(a, sort.key), vb = cellValue(b, sort.key);
      var emptyA = va === '' || va === null, emptyB = vb === '' || vb === null;
      if (emptyA && emptyB) return 0;
      if (emptyA) return 1;
      if (emptyB) return -1;
      var numA = Number(va), numB = Number(vb);
      var cmp;
      if (va !== '' && vb !== '' && !isNaN(numA) && !isNaN(numB)) {
        cmp = numA - numB;
      } else {
        cmp = String(va).localeCompare(String(vb), undefined, { numeric: true, sensitivity: 'base' });
      }
      return sort.dir === 'asc' ? cmp : -cmp;
    }

    function buildColumnChooser(visibleDefs) {
      var wrap = el('details', { className: 'columns' }, [el('summary', { className: 'columns__toggle', text: 'Columns' })]);
      var menu = el('div', { className: 'columns__menu' });
      columnDefs().forEach(function (d) {
        if (d.locked) return;
        var on = !!state.tableCols[d.key];
        var box = el('input', { type: 'checkbox' });
        box.checked = on;
        box.addEventListener('change', function () {
          state.tableCols[d.key] = box.checked;
          renderTable();
        });
        menu.appendChild(el('label', { className: 'columns__option' }, [box, el('span', { text: d.label })]));
      });
      wrap.appendChild(menu);
      return wrap;
    }

    function detailRow(node, colspan) {
      var tr = el('tr', { className: 'grid__detail' });
      var td = el('td', { colspan: String(colspan) });
      var card = document.getElementById(node.id);
      if (card) {
        var body = card.querySelector('.req__body');
        if (body) td.appendChild(body.cloneNode(true));
        var detail = card.querySelector('.req__detail');
        if (detail) td.appendChild(detail.cloneNode(true));
        var flags = card.querySelector('.req__flags');
        if (flags) td.appendChild(flags.cloneNode(true));
      } else {
        td.appendChild(el('p', { className: 'muted', text: 'No server-rendered content for this requirement.' }));
      }
      tr.appendChild(td);
      return tr;
    }

    // revealInTable selects a requirement and scrolls its row into view
    // without leaving the Table view.
    function revealInTable(id) {
      if (!byId[id]) return;
      state.selected = id;
      syncOutlineSelection();
      writeState();
      var row = document.querySelector('#req-table tr.grid__row[data-id="' + id + '"]');
      if (row) {
        row.scrollIntoView({ block: 'nearest' });
        row.classList.add('is-flash');
        setTimeout(function () { row.classList.remove('is-flash'); }, 1400);
      }
    }

    // ── trace view ──
    // ensureSelection returns the requirement the Trace view is centred on.
    // The selection is kept across view switches; it is only replaced when the
    // selected requirement disappears from the filtered set entirely.
    function ensureSelection() {
      var visible = visibleRequirements();
      var selected = state.selected && byId[state.selected];
      if (selected && visibleSet.has(state.selected)) return selected;
      if (!visible.length) {
        state.selected = null;
        return null;
      }
      if (!selected) state.selected = visible[0].id;
      return byId[state.selected];
    }

    function renderTrace() {
      if (!tracePanel) return;
      clear(tracePanel);
      var node = ensureSelection();
      tracePanel.appendChild(el('div', { className: 'view-head' }, [
        el('h2', { text: 'Trace explorer' }),
        el('p', { className: 'view-sub', text: 'Choose a requirement in the outline (or search), then click any neighbour to recenter. Requirements in other documents open there.' })
      ]));

      if (!node) {
        tracePanel.appendChild(el('div', { className: 'empty' }, [
          el('p', { className: 'empty__title', text: 'No requirements match the current filters' }),
          el('p', { text: 'Clear the search or filters to explore traces.' })
        ]));
        return;
      }
      if (!visibleSet.has(node.id)) {
        tracePanel.appendChild(el('p', { className: 'filter-note' }, [
          'This requirement is hidden by the current filters: ',
          el('button', { type: 'button', className: 'linkbtn', text: 'Clear filters' })
        ]));
        tracePanel.lastChild.lastChild.addEventListener('click', function () {
          state.query = '';
          state.filters = {};
          state.explicitFilters = true;
          if (searchEl) searchEl.value = '';
          apply();
        });
      }

      var nav = el('div', { className: 'view-tools' });
      if (state.trail.length) {
        var back = el('button', { type: 'button', className: 'btn', text: '← Back' });
        back.addEventListener('click', function () {
          state.selected = state.trail.pop();
          renderTrace();
          writeState();
        });
        nav.appendChild(back);
      }
      var openDoc = el('a', { className: 'btn', href: '#' + node.id, text: 'Open in document' });
      openDoc.addEventListener('click', function (e) {
        e.preventDefault();
        jumpTo(node.id);
      });
      var showTable = el('button', { type: 'button', className: 'btn', text: 'Show in table' });
      showTable.addEventListener('click', function () { setView('table'); });
      var matrixBtn = el('button', { type: 'button', className: 'btn' + (state.matrix ? ' is-active' : ''), text: state.matrix ? 'Hide matrix' : 'Full matrix' });
      matrixBtn.addEventListener('click', function () { state.matrix = !state.matrix; renderTrace(); });
      nav.appendChild(openDoc);
      nav.appendChild(showTable);
      nav.appendChild(matrixBtn);
      tracePanel.appendChild(nav);

      var grid = el('div', { className: 'trace' }, [
        traceColumn('Upstream', node.up || [], node, 'up'),
        traceCenter(node),
        traceColumn('Downstream', node.down || [], node, 'down')
      ]);
      tracePanel.appendChild(grid);

      if (state.matrix) tracePanel.appendChild(buildMatrix());
    }

    // selectQuietly records the selection without re-rendering a view; used by
    // the reading view, where the outline follows the scroll position.
    function selectQuietly(id) {
      if (!byId[id]) return;
      state.selected = id;
      state.trail = [];
      syncOutlineSelection();
      writeState();
    }

    function selectNode(id, href) {
      if (!byId[id]) {
        if (href) location.href = href;
        return;
      }
      if (state.selected && state.selected !== id) state.trail.push(state.selected);
      state.selected = id;
      state.matrix = false;
      renderTrace();
      syncOutlineSelection();
      writeState();
    }

    function traceColumn(title, links, node, dir) {
      var col = el('div', { className: 'trace__col' }, [
        el('div', { className: 'trace__head' }, [
          el('span', { text: title }),
          el('span', { className: 'trace__count', text: String(links.length) })
        ])
      ]);
      if (!links.length) {
        col.appendChild(el('p', { className: 'muted', text: 'None declared.' }));
        return col;
      }
      links.forEach(function (link) {
        if (!link.resolved) {
          col.appendChild(el('div', { className: 'trace__node trace__node--broken' }, [
            el('code', { text: link.ref }),
            el('span', { className: 'trace__meta', text: 'unresolved reference' })
          ]));
          return;
        }
        var inDocument = !!byId[link.id];
        var filteredOut = inDocument && !visibleSet.has(link.id);
        var tag = inDocument ? 'button' : 'a';
        var attrs = {
          className: 'trace__node trace__node--external' + (filteredOut ? ' is-muted' : ''),
          'data-link-id': link.id,
          title: (dir === 'up' ? node.id + ' → ' + link.id : link.id + ' → ' + node.id)
        };
        if (inDocument) attrs.type = 'button';
        else if (link.href) attrs.href = link.href;
        var button = el(tag, attrs, [
          el('code', { text: link.id }),
          el('span', { className: 'trace__title', text: link.title || '' }),
          traceMeta(link, dir, inDocument, filteredOut)
        ]);
        button.addEventListener('click', function () { selectNode(link.id); });
        col.appendChild(button);
      });
      return col;
    }

    function traceMeta(link, dir, inDocument, filteredOut) {
      var parts = [];
      parts.push(dir === 'up' ? 'this → ' + link.id : link.id + ' → this');
      if (link.pinned) {
        parts.push('pin ~' + link.pin + (link.version ? ' (v' + link.version + ')' : ''));
      }
      if (link.direction === 'outdated') parts.push('outdated pin');
      if (link.direction === 'predated') parts.push('predated pin');
      if (!inDocument) parts.push('other document');
      else if (filteredOut) parts.push('filtered out');
      return el('span', { className: 'trace__meta', text: parts.join(' · ') });
    }

    function traceCenter(node) {
      var center = el('div', { className: 'trace__center' }, [
        el('div', { className: 'trace__head', text: 'Selected requirement' }),
        el('div', { className: 'trace__card' }, [
          el('code', { className: 'trace__id', text: node.id }),
          el('div', { className: 'trace__title', text: node.title || '' }),
          el('div', { className: 'trace__badges' }, traceBadges(node))
        ])
      ]);
      return center;
    }

    function traceBadges(node) {
      var badges = [
        el('span', { className: 'chip chip--' + (node.coverage || 'none'), text: COVERAGE_LABELS[node.coverage] || 'no coverage data' })
      ];
      if (node.verdict) {
        badges.push(el('span', { className: 'chip chip--' + node.verdict, text: node.verdict }));
      } else if (node.eligible && model.hasResults) {
        badges.push(el('span', { className: 'chip chip--no-verdict', text: 'no verdict' }));
      }
      return badges;
    }

    function buildMatrix() {
      var visible = visibleRequirements();
      var wrap = el('section', { className: 'matrix' });
      var blocks = 0;
      wrap.appendChild(el('h3', { className: 'matrix__title', text: 'Full matrix' }));
      wrap.appendChild(el('p', { className: 'view-sub', text: 'Rows are requirements in this document; columns are the requirements they link to. Cell tooltips name the link direction.' }));
      [['up', 'Traces to'], ['down', 'Traced from']].forEach(function (pair) {
        var dir = pair[0];
        var cols = [];
        var seen = {};
        visible.forEach(function (n) {
          (n[dir] || []).forEach(function (l) {
            var key = l.resolved ? l.id : 'ref:' + l.ref;
            if (!seen[key]) { seen[key] = true; cols.push(l); }
          });
        });
        cols.sort(function (a, b) { return (a.id || a.ref).localeCompare(b.id || b.ref); });
        if (!cols.length) return;

        blocks++;
        var block = el('div', { className: 'matrix__block' }, [
          el('h4', { text: pair[1] + ' ' + cols.length })
        ]);
        var cells = cols.length * visible.length;
        if (cells > 20000) {
          block.appendChild(el('div', { className: 'empty' }, [
            el('p', { className: 'empty__title', text: 'Matrix too large to draw' }),
            el('p', { text: cells + ' cells would exceed the ' + '20000-cell limit. Narrow the filters, or use the trace neighbourhood above.' })
          ]));
          wrap.appendChild(block);
          return;
        }

        var table = el('table', { className: 'matrix__table' }, [
          el('thead', {}, [el('tr', {}, [el('th', { scope: 'col', className: 'matrix__corner' })].concat(cols.map(function (col) {
            var head = el('code', { text: col.id || col.ref });
          if (col.resolved && col.href) return el('th', { scope: 'col', className: 'matrix__col', title: col.title || col.ref }, [el('a', { href: col.href }, [head])]);
          if (col.resolved && byId[col.id]) return el('th', { scope: 'col', className: 'matrix__col', title: col.title || col.ref }, [el('a', { href: '#' + col.id, 'data-open-req': col.id }, [head])]);
          return el('th', { scope: 'col', className: 'matrix__col', title: col.title || col.ref }, [head]);
          })))])
        ]);
        var tbody = el('tbody');
        visible.forEach(function (n) {
          var row = el('tr', {}, [el('th', { scope: 'row', className: 'matrix__row' }, [el('a', { href: '#' + n.id, 'data-open-req': n.id }, [el('code', { text: n.id })])])]);
          cols.forEach(function (col) {
            var on = (n[dir] || []).some(function (l) {
              return l.resolved && col.resolved ? l.id === col.id : l.ref === col.ref;
            });
            if (!on) { row.appendChild(el('td', { className: 'matrix__cell' })); return; }
            // Direction-specific tooltip: in the downstream matrix the link
            // runs column → row.
            var label = dir === 'up' ? n.id + ' → ' + (col.id || col.ref) : (col.id || col.ref) + ' → ' + n.id;
            row.appendChild(el('td', { className: 'matrix__cell matrix__cell--on', title: label, 'data-open-req': dir === 'up' ? (col.id || '') : n.id, tabindex: '0' }, [
              el('span', { className: 'matrix__dot', 'aria-hidden': 'true', text: '●' }),
              el('span', { className: 'visually-hidden', text: label })
            ]));
          });
          tbody.appendChild(row);
        });
        table.appendChild(tbody);
        block.appendChild(el('div', { className: 'table-scroll' }, [table]));
        wrap.appendChild(block);
      });
      if (blocks === 0) {
        wrap.appendChild(el('p', { className: 'muted', text: 'No trace links among the filtered requirements.' }));
      }
      return wrap;
    }

    // ── coverage view ──
    function summarizeCoverage(list) {
      var s = { expected: 0, covered: 0, missing: 0, unresolved: 0, notRequired: 0, drafts: 0 };
      list.forEach(function (n) {
        s.drafts += n.drafts || 0;
        switch (n.coverage) {
          case 'covered': s.covered++; s.expected++; break;
          case 'missing-coverage': s.missing++; s.expected++; break;
          case 'unresolved': s.unresolved++; s.expected++; break;
          default: s.notRequired++;
        }
      });
      return s;
    }

    function summarizeVerdicts(list) {
      var s = { eligible: 0, pass: 0, fail: 0, skipped: 0, inconclusive: 0, none: 0 };
      list.forEach(function (n) {
        // Eligibility comes from the graph's missing-verdict gate, so the
        // denominator includes unverified measures and excludes draft,
        // deferred, and rejected ones.
        if (!n.eligible) return;
        s.eligible++;
        if (n.verdict === 'pass') s.pass++;
        else if (n.verdict === 'fail') s.fail++;
        else if (n.verdict === 'skipped') s.skipped++;
        else if (n.verdict === 'inconclusive') s.inconclusive++;
        else s.none++;
      });
      return s;
    }

    function summarizeLinks(list) {
      var s = { links: 0, unresolved: 0, outdated: 0, predated: 0 };
      list.forEach(function (n) {
        (n.up || []).forEach(function (l) {
          s.links++;
          if (!l.resolved) s.unresolved++;
          if (l.direction === 'outdated') s.outdated++;
          if (l.direction === 'predated') s.predated++;
        });
      });
      return s;
    }

    function renderCoverage() {
      if (!coveragePanel) return;
      clear(coveragePanel);
      var visible = visibleRequirements();
      var cov = summarizeCoverage(visible);
      var ver = summarizeVerdicts(visible);
      var links = summarizeLinks(visible);

      coveragePanel.appendChild(el('div', { className: 'view-head' }, [
        el('h2', { text: 'Coverage' }),
        el('p', { className: 'view-sub', text: 'Coverage, verification, and link health agree with reqmd check. Every number below covers the filtered requirement set.' })
      ]));
      if (visible.length !== reqNodes.length) {
        coveragePanel.appendChild(el('p', { className: 'filter-note', text: 'Filtered view: ' + visible.length + ' of ' + reqNodes.length + ' requirements.' }));
      }

      var tiles = el('div', { className: 'tiles' });
      tiles.appendChild(tile('Expected coverage', cov.expected ? cov.covered + ' of ' + cov.expected : 'not required',
        cov.missing + ' missing · ' + cov.unresolved + ' unresolved · ' + cov.drafts + ' draft downstreams ignored',
        cov.missing > 0 || cov.unresolved > 0));
      var verDetail = ver.eligible
        ? ver.fail + ' fail · ' + ver.none + ' no verdict · ' + ver.inconclusive + ' inconclusive · ' + ver.skipped + ' skipped'
        : 'No requirement here declares itself a measure for verification';
      tiles.appendChild(tile('Verification', ver.eligible ? ver.pass + ' of ' + ver.eligible + ' pass' : 'not applicable',
        verDetail, ver.fail > 0 || ver.none > 0));
      tiles.appendChild(tile('Trace links', String(links.links),
        links.unresolved + ' unresolved · ' + (links.outdated + links.predated) + ' stale pins', links.unresolved > 0 || links.outdated + links.predated > 0));
      coveragePanel.appendChild(tiles);

      var broken = visible.filter(function (n) {
        return (n.up || []).some(function (l) { return !l.resolved || l.direction === 'outdated' || l.direction === 'predated'; });
      });
      var queueSpecs = [
        {
          title: 'Missing coverage',
          body: 'Requirements whose expected downstream trace is missing.',
          filterAttr: 'state.coverage', filterValue: 'missing-coverage',
          list: visible.filter(function (n) { return n.coverage === 'missing-coverage'; })
        },
        {
          title: 'Unresolved coverage sources',
          body: 'Requirements expecting coverage from an unknown document-id or level.',
          filterAttr: 'state.coverage', filterValue: 'unresolved',
          list: visible.filter(function (n) { return n.coverage === 'unresolved'; })
        },
        {
          title: 'Missing verdicts',
          body: 'Eligible measures with no verification result.',
          filterAttr: 'state.verdict', filterValue: 'none',
          list: visible.filter(function (n) { return n.eligible && !n.verdict; })
        },
        {
          title: 'Failing verification',
          body: 'Measures whose rolled-up verdict is fail.',
          filterAttr: 'state.verdict', filterValue: 'fail',
          list: visible.filter(function (n) { return n.verdict === 'fail'; })
        },
        {
          title: 'Broken and stale traces',
          body: 'Requirements with unresolved references or stale version pins.',
          list: broken,
          noteFor: function (n) {
            return (n.up || []).filter(function (l) { return !l.resolved || l.direction === 'outdated' || l.direction === 'predated'; })
              .map(function (l) {
                if (!l.resolved) return 'unresolved: ' + l.ref;
                return (l.direction === 'outdated' ? 'outdated pin ' : 'predated pin ') + l.id + '~' + l.pin + ' (v' + l.version + ')';
              }).join('; ');
          }
        }
      ];
      var pending = queueSpecs.filter(function (spec) { return spec.list.length > 0; });
      if (pending.length) {
        var queues = el('div', { className: 'queues' });
        pending.forEach(function (spec) {
          queues.appendChild(queue(spec.title, spec.body, spec.filterAttr || null, spec.filterValue || null, spec.list, spec.noteFor));
        });
        coveragePanel.appendChild(queues);
      } else {
        coveragePanel.appendChild(el('div', { className: 'empty' }, [
          el('p', { className: 'empty__title', text: 'No coverage or verification exceptions in this view' }),
          el('p', { text: 'Every expected requirement is covered, every eligible measure has a verdict, and every declared trace link resolves.' })
        ]));
      }

      // Verification breakdown in worst-first order, with no-verdict in the
      // denominator.
      if (ver.eligible) {
        var bars = el('div', { className: 'dash-bars' });
        var order = VERDICT_ORDER.concat(['none']);
        order.forEach(function (outcome) {
          var n = outcome === 'none' ? ver.none : ver[outcome];
          if (!n) return;
          var label = outcome === 'none' ? 'no verdict' : outcome;
          var pct = n * 100 / ver.eligible;
          bars.appendChild(el('div', { className: 'dash-bar dash-bar--' + label.replace(' ', '-') }, [
            el('span', { className: 'dash-bar__label', text: label }),
            el('span', { className: 'dash-bar__track' }, [el('span', { className: 'dash-bar__fill', style: 'width:' + pct.toFixed(1) + '%' })]),
            el('span', { className: 'dash-bar__val', text: String(n) })
          ]));
        });
        coveragePanel.appendChild(el('div', { className: 'dash' }, [el('h3', { text: 'Verification breakdown' }), bars]));
      }
    }

    function tile(label, value, detail, attention) {
      return el('div', { className: 'tile' + (attention ? ' tile--attention' : '') }, [
        el('span', { className: 'tile__label', text: label }),
        el('span', { className: 'tile__value', text: value }),
        el('span', { className: 'tile__detail', text: detail })
      ]);
    }

    function queue(title, body, filterAttr, filterValue, list, noteFor) {
      var wrap = el('div', { className: 'queue' });
      wrap.appendChild(el('h3', { className: 'queue__title' }, [el('span', { text: title }), el('span', { className: 'queue__count', text: String(list.length) })]));
      wrap.appendChild(el('p', { className: 'queue__body', text: body }));
      if (!list.length) {
        wrap.appendChild(el('p', { className: 'muted', text: 'None.' }));
        return wrap;
      }
      var ul = el('ul', { className: 'queue__list' });
      list.slice(0, 100).forEach(function (n) {
        var note = noteFor ? noteFor(n) : '';
        var link = el('a', { href: '#' + n.id }, [el('code', { text: n.id }), el('span', { className: 'queue__req-title', text: n.title || '' })]);
        link.addEventListener('click', function (e) {
          e.preventDefault();
          jumpTo(n.id, true);
        });
        ul.appendChild(el('li', {}, [link, note ? el('span', { className: 'queue__note', text: note }) : null]));
      });
      wrap.appendChild(ul);
      if (list.length > 100) {
        wrap.appendChild(el('p', { className: 'queue__more', text: 'Showing the first 100 of ' + list.length + '.' }));
      }
      if (filterAttr) {
        var btn = el('button', { type: 'button', className: 'linkbtn', text: 'Show only these requirements' });
        btn.addEventListener('click', function () { setState(filterAttr, filterValue); });
        wrap.appendChild(btn);
      }
      return wrap;
    }

    // stickyChromeHeight returns the height of the sticky app bar plus, when it
    // is currently pinned, the sticky toolbar. Jumps must clear both or the
    // requirement heading ends up hidden underneath them.
    function stickyChromeHeight() {
      var bar = document.querySelector('.app-bar');
      var height = bar ? bar.offsetHeight : 0;
      var toolbar = document.querySelector('.toolbar');
      if (toolbar) height += toolbar.offsetHeight;
      return height;
    }

    // revealSelectedInDocument scrolls the reading view to the selected
    // requirement, so switching from Trace/Table/Coverage continues where the
    // reader was instead of jumping to the top of the document.
    //
    // The document is server-rendered, but positions shift while the hidden
    // panel is laid out again, so the scroll is repeated on the next frames.
    function revealSelectedInDocument() {
      var id = state.selected;
      if (!id) return;
      var target = document.getElementById(id);
      if (!target) return;

      var html = document.documentElement;
      var smooth = html.style.scrollBehavior;
      html.style.scrollBehavior = 'auto';
      var offset = stickyChromeHeight() + 8;

      // Scrolling to a position derived from getBoundingClientRect would
      // compound earlier scrolls; use the absolute document offset instead.
      function jump() {
        var top = target.getBoundingClientRect().top + window.scrollY - offset;
        window.scrollTo(0, Math.max(0, top));
      }
      function highlight() {
        document.querySelectorAll('.toc__link.is-active').forEach(function (a) { a.classList.remove('is-active'); });
        var link = document.querySelector('.toc__link[href="#' + id + '"]');
        if (link) link.classList.add('is-active');
      }

      scrollSpySuspended = true;
      jump();
      highlight();
      setTimeout(function () {
        jump();
        highlight();
        html.style.scrollBehavior = smooth;
        // Let the observer settle on the final position before it resumes,
        // and re-assert the selection highlight afterwards.
        setTimeout(function () {
          scrollSpySuspended = false;
          highlight();
        }, 60);
      }, 0);
    }

    // ── navigation helpers ──
    // jumpTo reveals a requirement without discarding the reviewer's context.
    // keepFilters keeps search and attribute filters; the target may be hidden
    // by them, in which case the filters are cleared as a last resort.
    function jumpTo(id, keepFilters) {
      state.view = 'document';
      state.selected = id;
      if (!keepFilters) {
        state.query = '';
        state.filters = {};
        state.explicitFilters = true;
        if (searchEl) searchEl.value = '';
      }
      computeVisible();
      if (keepFilters && !visibleSet.has(id)) {
        state.query = '';
        state.filters = {};
        state.explicitFilters = true;
        if (searchEl) searchEl.value = '';
      }
      apply();
      setView('document');
      var target = document.getElementById(id);
      if (target) {
        window.scrollTo(0, Math.max(0, target.getBoundingClientRect().top + window.scrollY - stickyChromeHeight() - 8));
        target.classList.add('is-flash');
        setTimeout(function () { target.classList.remove('is-flash'); }, 1400);
        try { history.replaceState(null, '', location.pathname + location.search + '#' + id); } catch (err) { /* file:// */ }
      }
    }

    // ── overlays ──
    var overlayOpen = null;

    function openOverlay(name) {
      var overlay = overlays[name];
      if (!overlay) return;
      lastFocus = document.activeElement;
      overlayOpen = name;
      overlay.hidden = false;
      document.body.classList.add('overlay-open');
      if (name === 'palette') {
        paletteIndex = 0;
        if (paletteInput) { paletteInput.value = ''; }
        renderPalette('');
        if (paletteInput) paletteInput.focus();
      } else {
        // The panel itself is the first tab stop when it has no controls.
        var focusable = overlay.querySelector(FOCUSABLE) || overlay.querySelector('[tabindex="-1"]');
        if (focusable) focusable.focus();
      }
    }

    function closeOverlays() {
      Object.keys(overlays).forEach(function (k) {
        if (overlays[k]) overlays[k].hidden = true;
      });
      overlayOpen = null;
      document.body.classList.remove('overlay-open');
      if (lastFocus && lastFocus.focus) lastFocus.focus();
    }

    var commands = [
      { key: 'view:document', glyph: '▤', label: 'View: Document', hint: 'reading view', run: function () { setView('document'); } },
      { key: 'view:table', glyph: '▦', label: 'View: Table', hint: 'dense sortable grid', run: function () { setView('table'); } },
      { key: 'view:trace', glyph: '⌘', label: 'View: Trace', hint: 'trace explorer', run: function () { setView('trace'); } },
      { key: 'view:coverage', glyph: '◔', label: 'View: Coverage', hint: 'coverage and verification', run: function () { setView('coverage'); } },
      { key: 'toggle:density', glyph: '⇕', label: 'Toggle density', hint: 'comfortable / compact', run: function () { toggleDensity(); } },
      { key: 'filters:clear', glyph: '×', label: 'Clear search and filters', hint: 'reset the current view', run: function () {
        state.filters = {}; state.query = ''; state.explicitFilters = true; if (searchEl) searchEl.value = ''; apply();
      } },
      { key: 'help:shortcuts', glyph: '?', label: 'Keyboard shortcuts', hint: 'show all shortcuts', run: function () { openOverlay('shortcuts'); } }
    ];

    // paletteItems returns the results for a query. Requirements are matched
    // on ID and title, ranked by match quality so an exact ID prefix like
    // "SYS-FMT-" lists the SYS-FMT requirements first.
    function paletteItems(q) {
      var out = [];
      commands.forEach(function (c) {
        if (!q || c.label.toLowerCase().indexOf(q) !== -1) out.push(c);
      });
      if (!q) return out.slice(0, 40);

      // Normalized query: punctuation and case are ignored so "sysfm1" finds
      // SYS-FMT-001. Fuzzy matches are only considered from 2 characters on,
      // otherwise a single letter would match almost every requirement.
      var needle = normalize(q);
      var fuzzy = needle.length >= 2;
      var reqHits = [];
      for (var i = 0; i < reqNodes.length; i++) {
        var node = reqNodes[i];
        var score = matchScore(node, q, needle, fuzzy);
        if (score < 0) continue;
        reqHits.push({ node: node, score: score, index: i });
      }
      reqHits.sort(function (a, b) {
        if (a.score !== b.score) return a.score - b.score;
        return a.index - b.index;
      });
      var limit = fuzzy ? 30 : 40;
      reqHits.slice(0, limit).forEach(function (hit) {
        var node = hit.node;
        out.push({ key: 'req:' + node.id, glyph: '§', label: node.id, hint: node.title || '', reqId: node.id });
      });
      return out.slice(0, 40);
    }

    // normalize lowercases and drops everything that is not a letter or a
    // digit, so "sys_fmt-1", "SYS.FMT.1" and "sysfm1" compare equal.
    function normalize(s) {
      return (s || '').toLowerCase().replace(/[^a-z0-9]/g, '');
    }

    // matchScore ranks a requirement against the palette query; -1 means "no
    // match". Lower is better:
    //   0  ID equals the query
    //   1  query is an ID prefix / starts an ID segment ("sysfmt1", "fmt")
    //   2  ID contains the query
    //   3  title starts with the query
    //   4  title contains the query
    //   5+ fuzzy: the query's characters appear in order
    //
    // A query that spells an ID segment prefix ("graph" → SW-GRA-001) is
    // treated like an ID hit rather than a title hit; otherwise a title that
    // merely starts with the word would outrank it.
    function matchScore(node, q, needle, fuzzy) {
      var id = node.id.toLowerCase();
      var title = (node.title || '').toLowerCase();
      var dense = normalize(q);
      if (id === q) return 0;
      if (dense.length >= 2 && denseStartsSegment(node.id, dense)) return 1;
      if (id.indexOf(q) === 0) return 1;
      if (idSegmentStart(id, q)) return 1;
      if (id.indexOf(q) !== -1) return 2;
      if (title.indexOf(q) === 0) return 3;
      if (title.indexOf(q) !== -1) return 4;
      if (!fuzzy) return -1;

      var idScore = fuzzyScore(node.id, needle);
      var titleScore = fuzzyScore(node.title || '', needle);
      if (idScore < 0 && titleScore < 0) return -1;
      var best = idScore;
      // An ID match always beats a title match of the same quality.
      if (titleScore >= 0 && (best < 0 || titleScore + 8 < best)) best = titleScore + 8;
      return 5 + best;
    }

    // fuzzyScore scores how well needle matches hay. Two readings are tried
    // and the better one wins:
    //   literal — the needle appears contiguously (cheap, prefers the start of
    //             a segment: "graph" → "…/graph");
    //   dense   — the needle matches the separator-free haystack, so "sysfm1"
    //             finds "SYS-FMT-001" (leading characters must be a prefix of
    //             a segment);
    //   sparse  — the needle's characters appear in order anywhere.
    function fuzzyScore(hay, needle) {
      if (!needle) return 0;
      var lower = hay.toLowerCase();
      var best = literalScore(lower, needle);
      var dense = denseScore(normalize(lower), needle);
      if (dense >= 0 && (best < 0 || dense < best)) best = dense;
      var sparse = sparseScore(lower, needle);
      if (sparse >= 0 && (best < 0 || sparse < best)) best = sparse;
      return best;
    }

    // literalScore rewards a contiguous match and matches starting a segment.
    function literalScore(lower, needle) {
      var at = lower.indexOf(needle);
      if (at < 0) return -1;
      var boundary = at === 0 || !/[a-z0-9]/.test(lower[at - 1]);
      return at * 0.2 + (boundary ? 0 : 4);
    }

    // denseScore matches against the separator-free haystack. The first two
    // needle characters must occur contiguously (normally as a segment
    // prefix, e.g. "fm" in "…fmt001"), which keeps unrelated hits out; the
    // rest may only be skimmed over a short distance.
    function denseScore(dense, needle) {
      var head = needle.slice(0, 2);
      var at = dense.lastIndexOf(head);
      if (at < 0) return -1;
      var pos = at + head.length;
      var skim = 0;
      for (var i = head.length; i < needle.length; i++) {
        var next = dense.indexOf(needle[i], pos);
        if (next < 0) return -1;
        skim += next - pos;
        pos = next + 1;
      }
      return at * 0.2 + skim * 0.15;
    }

    // sparseScore is a plain subsequence walk; consecutive characters and
    // segment starts are cheap, scattered characters are expensive.
    function sparseScore(hay, needle) {
      var hi = 0;
      var score = 0;
      var prev = -2;
      for (var i = 0; i < needle.length; i++) {
        var found = hay.indexOf(needle[i], hi);
        if (found < 0) return -1;
        var gap = found - (prev + 1);
        if (gap === 0) score -= 2;
        else if (gap <= 2) score -= 1;
        else score += gap;
        if (found === 0 || !/[a-z0-9]/.test(hay[found - 1])) score -= 1.5;
        prev = found;
        hi = found + 1;
      }
      return score + 6;
    }

    // denseStartsSegment reports whether the separator-free query starts one
    // of the ID's segments after the document prefix: "swgra" and "gra" hit
    // SW-GRA-001, "graph" does not.
    function denseStartsSegment(id, denseQuery) {
      var segments = id.toLowerCase().split(/[^a-z0-9]+/).filter(Boolean);
      for (var i = 1; i < segments.length; i++) {
        if (segments[i].indexOf(denseQuery) === 0) return true;
      }
      return false;
    }

    // idSegmentStart reports whether q starts one of the ID's segments,
    // separators included ("fmt" in "sys-fmt-001").
    function idSegmentStart(id, q) {
      var at = id.indexOf(q);
      while (at !== -1) {
        if (at > 0 && !/[a-z0-9]/.test(id[at - 1])) return true;
        at = id.indexOf(q, at + 1);
      }
      return false;
    }

    function paletteQuery() {
      return (paletteInput ? paletteInput.value : '').toLowerCase().trim();
    }

    // paletteActive returns the index of the row marked active, which is the
    // single source of truth for "what will Enter run".
    function paletteActive() {
      if (!paletteList) return -1;
      var active = paletteList.querySelector('.palette__item.is-active');
      if (!active) return -1;
      return Number(active.getAttribute('data-index'));
    }

    function setPaletteActive(i) {
      if (!paletteList) return;
      var buttons = paletteList.querySelectorAll('.palette__item');
      if (!buttons.length) return;
      i = Math.max(0, Math.min(i, buttons.length - 1));
      Array.prototype.forEach.call(buttons, function (b, n) {
        var on = n === i;
        b.classList.toggle('is-active', on);
        var li = b.parentElement;
        if (li) li.setAttribute('aria-selected', on ? 'true' : 'false');
      });
      paletteIndex = i;
      var active = buttons[i];
      if (active && active.scrollIntoView) active.scrollIntoView({ block: 'nearest' });
    }

    function renderPalette(q) {
      if (!paletteList) return;
      clear(paletteList);
      var items = paletteItems(q);
      var active = paletteIndex;
      if (active >= items.length) active = 0;
      items.forEach(function (item, i) {
        var button = el('button', {
          type: 'button',
          className: 'palette__item' + (i === active ? ' is-active' : ''),
          'data-index': i,
          'data-glyph': item.glyph,
          'data-label': item.label,
          'data-hint': item.hint || ''
        }, [
          el('span', { className: 'palette__glyph', 'aria-hidden': 'true', text: item.glyph }),
          el('span', { className: 'palette__text' }, [
            el('span', { className: 'palette__label', text: item.label }),
            el('span', { className: 'palette__hint', text: item.hint || '' })
          ])
        ]);
        button.addEventListener('click', function () { runPaletteIndex(Number(button.getAttribute('data-index'))); });
        button.addEventListener('mousemove', function () {
          if (paletteActive() !== i) setPaletteActive(i);
        });
        paletteList.appendChild(el('li', { role: 'option', 'aria-selected': i === active ? 'true' : 'false' }, [button]));
      });
      paletteIndex = active;
      if (paletteEmpty) paletteEmpty.hidden = items.length > 0;
      var activeButton = paletteList.querySelector('.palette__item.is-active');
      if (activeButton && activeButton.scrollIntoView) activeButton.scrollIntoView({ block: 'nearest' });
    }

    // runPaletteIndex runs the row at index i. The row carries everything it
    // needs, so a stale list cannot run the wrong item.
    function runPaletteIndex(i) {
      var buttons = paletteList ? paletteList.querySelectorAll('.palette__item') : [];
      var button = buttons[i];
      if (!button) return;
      var reqId = null;
      var label = button.getAttribute('data-label');
      var command = commands.filter(function (c) { return c.label === label; })[0];
      if (button.getAttribute('data-glyph') === '§') reqId = label;
      closeOverlays();
      if (reqId) {
        jumpTo(reqId);
      } else if (command && typeof command.run === 'function') {
        command.run();
      }
    }

    function movePalette(delta) {
      var buttons = paletteList ? paletteList.querySelectorAll('.palette__item') : [];
      if (!buttons.length) return;
      var current = paletteActive();
      if (current < 0) current = 0;
      setPaletteActive((current + delta + buttons.length) % buttons.length);
    }

    function toggleDensity() {
      state.density = state.density === 'compact' ? 'comfortable' : 'compact';
      document.body.classList.toggle('density-compact', state.density === 'compact');
      writeState();
    }

    // ── wiring ──
    function onKey(e) {
      var t = e.target;
      var typing = t && (t.tagName === 'INPUT' || t.tagName === 'TEXTAREA' || t.tagName === 'SELECT' || t.isContentEditable);

      if (overlayOpen) {
        if (e.key === 'Escape') { closeOverlays(); return; }
        if (overlayOpen === 'palette') {
          if (e.key === 'ArrowDown') { e.preventDefault(); movePalette(1); return; }
          if (e.key === 'ArrowUp') { e.preventDefault(); movePalette(-1); return; }
          if (e.key === 'Enter') { e.preventDefault(); runPaletteIndex(paletteActive()); return; }
          trapFocus(overlays.palette.querySelector('.overlay__panel'), e);
          return;
        }
        trapFocus(overlays.shortcuts.querySelector('.overlay__panel'), e);
        return;
      }

      if ((e.metaKey || e.ctrlKey) && !e.altKey && e.key.toLowerCase() === 'k') {
        e.preventDefault();
        openOverlay('palette');
        return;
      }
      // Ignore modified keys for single-letter shortcuts.
      if (e.metaKey || e.ctrlKey || e.altKey) return;
      if (typing) {
        if (e.key === 'Escape' && t.blur) t.blur();
        return;
      }

      if (e.key === 'Escape') {
        if (document.body.classList.contains('toc-open')) {
          document.body.classList.remove('toc-open');
        }
        return;
      }
      if (e.key === '/') {
        e.preventDefault();
        if (searchEl) searchEl.focus();
        return;
      }
      if (e.key === '?') { e.preventDefault(); openOverlay('shortcuts'); return; }
      if (e.key === 't') { toggleThemeButton(); return; }
      if (e.key === 'f') { toggleDensity(); return; }
      var now = Date.now();
      if (pendingG && now - pendingG < 900) {
        pendingG = 0;
        var map = { d: 'document', t: 'table', r: 'trace', c: 'coverage' };
        if (map[e.key]) { e.preventDefault(); setView(map[e.key]); }
        return;
      }
      if (e.key === 'g') pendingG = now;
    }
    var pendingG = 0;

    function toggleThemeButton() {
      var b = document.querySelector('[data-action="theme"]');
      if (b) b.click();
    }

    document.addEventListener('keydown', onKey);

    Array.prototype.forEach.call(document.querySelectorAll('[data-action]'), function (btn) {
      var action = btn.getAttribute('data-action');
      btn.addEventListener('click', function () {
        switch (action) {
          case 'palette': openOverlay('palette'); break;
          case 'density': toggleDensity(); break;
          case 'toggle-toc': {
            var open = document.body.classList.toggle('toc-open');
            btn.setAttribute('aria-expanded', open ? 'true' : 'false');
            break;
          }
          case 'close-overlays': closeOverlays(); break;
        }
      });
    });

    // Requirement references rendered inside an analytical view open the
    // Document view, where the anchor actually lives.
    document.addEventListener('click', function (e) {
      var hit = e.target.closest ? e.target.closest('[data-open-req]') : null;
      if (!hit) return;
      var id = hit.getAttribute('data-open-req');
      if (!id || !byId[id]) return;
      e.preventDefault();
      jumpTo(id, true);
    });
    document.addEventListener('keydown', function (e) {
      if (e.key !== 'Enter' && e.key !== ' ') return;
      var hit = e.target.closest ? e.target.closest('[data-open-req]') : null;
      if (!hit || hit.tagName === 'A') return;
      var id = hit.getAttribute('data-open-req');
      if (!id || !byId[id]) return;
      e.preventDefault();
      jumpTo(id, true);
    });

    Array.prototype.forEach.call(document.querySelectorAll('.viewswitch__btn'), function (btn) {
      btn.addEventListener('click', function () { setView(btn.getAttribute('data-view')); });
    });

    if (searchEl) {
      var debounce;
      searchEl.addEventListener('input', function () {
        clearTimeout(debounce);
        debounce = setTimeout(function () {
          state.query = searchEl.value;
          apply();
        }, 120);
      });
    }

    Array.prototype.forEach.call(document.querySelectorAll('[data-filter-attr]'), function (input) {
      input.addEventListener('change', function () {
        toggleFilter(input.getAttribute('data-filter-attr'), input.value, input.checked);
      });
    });

    if (paletteInput) {
      paletteInput.addEventListener('input', function () {
        paletteIndex = 0;
        renderPalette(paletteQuery());
      });
    }

    // ── init ──
    restoreState();
    document.body.classList.toggle('density-compact', state.density === 'compact');
    setView(state.view);

    // A #REQ-ID deep link selects Document mode and reveals the target.
    if (location.hash.length > 1) {
      var hashId = decodeURIComponent(location.hash.slice(1));
      if (byId[hashId] || document.getElementById(hashId)) {
        jumpTo(hashId);
      }
    }

    setupOutlineClicks();
    setupScrollSpy();
    apply();

    // Expose a tiny surface for tests and same-page integrations.
    window.reqmd = {
      setView: setView,
      setFilter: setState,
      setSearch: function (q) {
        state.query = q;
        if (searchEl) searchEl.value = q;
        apply();
      },
      jumpTo: jumpTo,
      state: state,
      visibleIDs: function () { return Array.from(visibleSet); }
    };
  }

  // ── project page ──
  function initProjectPage() {
    var reqs = readJSON('project-reqs') || [];
    var input = document.querySelector('[data-project-search]');
    var results = document.querySelector('[data-project-results]');
    if (!input || !results) return;

    function render(q) {
      clear(results);
      q = (q || '').trim().toLowerCase();
      if (!q) return;
      var matches = [];
      for (var i = 0; i < reqs.length; i++) {
        var r = reqs[i];
        var hay = (r.id + ' ' + (r.title || '') + ' ' + (r.doc || '')).toLowerCase();
        if (hay.indexOf(q) === -1) continue;
        matches.push(r);
      }
      if (!matches.length) {
        results.appendChild(el('li', { className: 'muted', text: 'No matching requirements.' }));
        return;
      }
      var summary = matches.length > 50
        ? 'Showing the first 50 of ' + matches.length + ' matches.'
        : matches.length + (matches.length === 1 ? ' match.' : ' matches.');
      results.appendChild(el('li', { className: 'results__count', text: summary }));
      matches.slice(0, 50).forEach(function (r) {
        results.appendChild(el('li', {}, [
          el('a', { href: r.href }, [el('code', { text: r.id }), el('span', { className: 'queue__req-title', text: r.title || '' })]),
          el('span', { className: 'queue__note', text: r.doc })
        ]));
      });
    }

    input.addEventListener('keydown', function (e) {
      if (e.key !== 'ArrowDown') return;
      var first = results.querySelector('a');
      if (!first) return;
      e.preventDefault();
      first.focus();
    });
    results.addEventListener('keydown', function (e) {
      var items = Array.prototype.slice.call(results.querySelectorAll('a'));
      var i = items.indexOf(document.activeElement);
      if (i === -1) return;
      if (e.key === 'ArrowDown' && items[i + 1]) { e.preventDefault(); items[i + 1].focus(); }
      if (e.key === 'ArrowUp') { e.preventDefault(); (items[i - 1] || input).focus(); }
    });

    var debounce;
    input.addEventListener('input', function () {
      clearTimeout(debounce);
      debounce = setTimeout(function () { render(input.value); }, 120);
    });

    // The catalog table is the primary way into a document: make the whole
    // row clickable, not just its title link.
    var catalog = document.getElementById('catalog');
    if (catalog) {
      catalog.addEventListener('click', function (e) {
        if (e.target.closest('a')) return;
        var row = e.target.closest('tr[data-href]');
        if (row) location.href = row.getAttribute('data-href');
      });
    }
  }

  if (typeof module !== 'undefined' && module.exports) {
    module.exports = {
      decodeFilters: decodeFilters,
      encodeFilters: encodeFilters,
      encodeUrlState: encodeUrlState,
      decodeUrlState: decodeUrlState
    };
  }

  if (typeof document !== 'undefined') {
    document.addEventListener('DOMContentLoaded', function () {
      setupTheme();
      setupModHints();
      if (document.body.getAttribute('data-page') === 'project') initProjectPage();
      else initDocumentPage();
    });
  }
})();
