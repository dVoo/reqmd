// ── Alpine component ──
// Owns: theme, mobile-sidebar open state, search term, active filter
// chips, the parsed requirement index, and per-card visibility logic.
function reqmdApp() {
  return {
    theme: (typeof localStorage !== 'undefined' && localStorage.getItem('theme')) || 'light',
    tocOpen: false,
    searchTerm: '',
    totalCount: 0,
    visibleCount: 0,
    activeFiltersByAttr: {},
    filterOptions: {},
    reqIndex: [],
    reqMap: new Map(),

    initApp() {
      // Apply the stored theme to <html> before paint to avoid a flash.
      this.theme = (typeof localStorage !== 'undefined' && localStorage.getItem('theme')) || 'light';
      if (this.theme === 'dark') {
        document.documentElement.classList.add('dark');
      } else {
        document.documentElement.classList.remove('dark');
      }

      // Parse the requirement index (used by isCardVisible and the count).
      var indexEl = document.getElementById('req-index');
      if (indexEl) {
        try { this.reqIndex = JSON.parse(indexEl.textContent); } catch (e) { this.reqIndex = []; }
      }
      var self = this;
      this.reqMap = new Map(this.reqIndex.map(function (r) { return [r.id, r]; }));
      this.totalCount = this.reqIndex.length;
      this.visibleCount = this.totalCount;

      // Parse the filter options (built in Go from the same data the
      // old vanilla-JS code computed; Alpine renders the <select>s from
      // this declaratively).
      var filterEl = document.getElementById('req-filter-index');
      if (filterEl) {
        try { this.filterOptions = JSON.parse(filterEl.textContent); } catch (e) { this.filterOptions = {}; }
      }

      // F1: pre-check the default status filter (declared on the toolbar
      // as data-default-status. An empty value skips this entirely
      // (used for ignore-status documents where the lifecycle is
      // bypassed).
      var toolbar = document.querySelector('.search-toolbar');
      var defaultStatus = toolbar ? toolbar.getAttribute('data-default-status') : '';
      if (defaultStatus && this.filterOptions && this.filterOptions.status && this.filterOptions.status.indexOf(defaultStatus) !== -1) {
        this.activeFiltersByAttr.status = [defaultStatus];
      }

      // Re-derive visibleCount whenever any input changes. Alpine's
      // $watch is the canonical way to subscribe to reactive state.
      this.$watch('searchTerm', function () { self.recomputeCounts(); });
      this.$watch('activeFiltersByAttr', function () { self.recomputeCounts(); }, { deep: true });

      // Initial count.
      this.recomputeCounts();
    },

    toggleTheme() {
      this.theme = this.theme === 'dark' ? 'light' : 'dark';
      if (this.theme === 'dark') {
        document.documentElement.classList.add('dark');
      } else {
        document.documentElement.classList.remove('dark');
      }
      try { localStorage.setItem('theme', this.theme); } catch (e) { /* private mode */ }
    },

    toggleFilter(attr, value) {
      if (!this.activeFiltersByAttr[attr]) {
        this.activeFiltersByAttr[attr] = [];
      }
      var idx = this.activeFiltersByAttr[attr].indexOf(value);
      if (idx === -1) {
        this.activeFiltersByAttr[attr].push(value);
      } else {
        this.activeFiltersByAttr[attr].splice(idx, 1);
        if (this.activeFiltersByAttr[attr].length === 0) {
          delete this.activeFiltersByAttr[attr];
        }
      }
      // Force reactivity: reassign the outer object
      this.activeFiltersByAttr = Object.assign({}, this.activeFiltersByAttr);
    },

    removeFilter(attr, value) {
      var arr = this.activeFiltersByAttr[attr];
      if (!arr) return;
      this.activeFiltersByAttr[attr] = arr.filter(function (v) { return v !== value; });
      if (this.activeFiltersByAttr[attr].length === 0) {
        delete this.activeFiltersByAttr[attr];
      }
      // Trigger reactivity by reassigning
      this.activeFiltersByAttr = Object.assign({}, this.activeFiltersByAttr);
    },

    clearAll() {
      this.activeFiltersByAttr = {};
      this.searchTerm = '';
    },

    prettyName(attr) {
      if (!attr) return '';
      return attr.charAt(0).toUpperCase() + attr.slice(1).replace(/[-_]/g, ' ');
    },

    // Per-card visibility. Used by x-show on every requirement article.
    // Returns true when the card matches the current search term AND
    // every active filter (AND semantics across filters).
    isCardVisible(cardId) {
      var entry = this.reqMap.get(cardId);
      if (!entry) return true;

      var match = true;

      // Text search across id, body, rationale, and all scalar attribute values.
      if (this.searchTerm && this.searchTerm.trim()) {
        var term = this.searchTerm.toLowerCase().trim();
        var attrTexts = [];
        if (entry.attrs) {
          Object.keys(entry.attrs).forEach(function (k) {
            var v = entry.attrs[k];
            if (Array.isArray(v)) {
              v.forEach(function (item) {
                if (item !== null && typeof item !== 'object') attrTexts.push(String(item));
              });
            } else if (v !== null && typeof v !== 'object') {
              attrTexts.push(String(v));
            }
          });
        }
        var searchText = (entry.id + ' ' + (entry.body || '') + ' ' + (entry.rationale || '') + ' ' + attrTexts.join(' ')).toLowerCase();
        if (searchText.indexOf(term) === -1) match = false;
      }

      // Attribute filters: OR within each attr, AND across attrs.
      if (match && Object.keys(this.activeFiltersByAttr).length > 0 && entry.attrs) {
        var attrNames = Object.keys(this.activeFiltersByAttr);
        for (var a = 0; a < attrNames.length; a++) {
          var attr = attrNames[a];
          var selectedValues = this.activeFiltersByAttr[attr];
          if (!selectedValues || selectedValues.length === 0) continue;
          var attrVal = entry.attrs[attr];
          var hasMatch = false;
          if (attrVal !== null && attrVal !== undefined) {
            if (Array.isArray(attrVal)) {
              for (var si = 0; si < selectedValues.length; si++) {
                if (attrVal.some(function (v) { return String(v) === selectedValues[si]; })) {
                  hasMatch = true;
                  break;
                }
              }
            } else {
              var strVal = String(attrVal);
              for (var si = 0; si < selectedValues.length; si++) {
                if (strVal === selectedValues[si]) { hasMatch = true; break; }
              }
            }
          }
          if (!hasMatch) { match = false; break; }
        }
      }

      return match;
    },

    // Recompute visibleCount by walking the index. Triggered by $watch on
    // searchTerm and activeFilters. x-text binds visibleCount + '/' +
    // totalCount into the search-count span.
    recomputeCounts() {
      var n = 0;
      for (var i = 0; i < this.reqIndex.length; i++) {
        if (this.isCardVisible(this.reqIndex[i].id)) n++;
      }
      this.visibleCount = n;
    }
  };
}

// ── Vanilla JS: TOC tree + scroll-spy ──
// Runs on DOMContentLoaded. By this point Alpine's <script defer> has
// parsed but the Alpine component may not have booted yet — the TOC
// tree is built from #req-index independently of Alpine.
document.addEventListener('DOMContentLoaded', function () {
  // ── Build Tree TOC Sidebar ──
  var tocList = document.getElementById('req-toc-list');
  var tocCount = document.getElementById('req-toc-count');
  var indexEl = document.getElementById('req-index');

  if (tocList && indexEl) {
    var reqs;
    try { reqs = JSON.parse(indexEl.textContent); } catch (e) { reqs = []; }

    if (reqs && reqs.length) {
      var topLevel = [];
      var childMap = {};
      var entryMap = {};

      reqs.forEach(function (r) {
        entryMap[r.id] = r;
        if (r.parentId) {
          if (!childMap[r.parentId]) childMap[r.parentId] = [];
          childMap[r.parentId].push(r);
        } else {
          topLevel.push(r);
        }
      });

      var childCount = 0;
      for (var pid in childMap) childCount += childMap[pid].length;

      if (tocCount) {
        tocCount.textContent = reqs.length + ' requirements (' + topLevel.length + ' parents \u00B7 ' + childCount + ' children)';
      }

      function createTreeItem(r, isChild) {
        var wrapper = document.createElement('div');
        wrapper.className = isChild ? 'req-toc-item req-toc-item--child' : 'req-toc-item req-toc-item--parent';
        wrapper.setAttribute('data-id', r.id);

        var hasKids = r.children && r.children.length > 0;

        if (hasKids) {
          var toggle = document.createElement('button');
          toggle.className = 'req-toc-toggle';
          toggle.textContent = '\u25B6';
          toggle.setAttribute('aria-label', 'Expand ' + r.id);
          toggle.setAttribute('aria-expanded', 'false');
          wrapper.appendChild(toggle);
        }

        var mainDiv = document.createElement('div');
        mainDiv.className = 'req-toc-item-main';

        var idSpan = document.createElement('span');
        idSpan.className = 'req-toc-item-id';
        idSpan.innerHTML = '<code>' + r.id + '</code>';
        mainDiv.appendChild(idSpan);

        if (r.title) {
          var titleSpan = document.createElement('span');
          titleSpan.className = 'req-toc-item-title';
          titleSpan.textContent = r.title;
          mainDiv.appendChild(titleSpan);
        }

        var bodySpan = document.createElement('span');
        bodySpan.className = 'req-toc-item-body';
        bodySpan.textContent = r.body || '(no description)';
        mainDiv.appendChild(bodySpan);

        wrapper.appendChild(mainDiv);

        if (hasKids) {
          var badge = document.createElement('span');
          badge.className = 'req-toc-child-count';
          badge.textContent = r.children.length;
          wrapper.appendChild(badge);
        }

        wrapper.addEventListener('click', function (e) {
          if (e.target.classList.contains('req-toc-toggle')) return;
          var target = document.getElementById(r.id);
          if (target) {
            target.scrollIntoView({ behavior: 'smooth', block: 'start' });
            target.style.transition = 'border-color 0.3s ease';
            target.style.borderColor = 'var(--accent)';
            setTimeout(function () { target.style.borderColor = ''; }, 2000);
          }
        });

        if (hasKids) {
          var childrenWrap = document.createElement('div');
          childrenWrap.className = 'req-toc-children';
          wrapper.appendChild(childrenWrap);

          var kids = r.children || [];
          kids.forEach(function (kidId) {
            var kidEntry = entryMap[kidId];
            if (kidEntry) {
              childrenWrap.appendChild(createTreeItem(kidEntry, true));
            }
          });

          toggle.addEventListener('click', function (e) {
            e.stopPropagation();
            var isExpanded = childrenWrap.classList.toggle('req-toc-children--expanded');
            toggle.classList.toggle('req-toc-toggle--expanded', isExpanded);
            toggle.setAttribute('aria-expanded', isExpanded ? 'true' : 'false');
          });
        }

        return wrapper;
      }

      topLevel.forEach(function (r) {
        tocList.appendChild(createTreeItem(r, false));
      });
    }
  }

  // ── Scroll-spy: highlight active TOC item ──
  var tocItems = document.querySelectorAll('.req-toc-item');
  var cards = document.querySelectorAll('.req-card, .req-card-child');
  if (tocItems.length && cards.length) {
    var observer = new IntersectionObserver(function (entries) {
      entries.forEach(function (entry) {
        if (entry.isIntersecting) {
          tocItems.forEach(function (item) {
            item.classList.remove('req-toc-item--active');
            item.classList.remove('req-toc-item--active-parent');
          });
          var activeItem = document.querySelector('.req-toc-item[data-id="' + entry.target.id + '"]');
          if (activeItem) {
            activeItem.classList.add('req-toc-item--active');
            var parent = activeItem.parentElement;
            while (parent) {
              if (parent.classList.contains('req-toc-item--parent')) {
                parent.classList.add('req-toc-item--active-parent');
                break;
              }
              parent = parent.parentElement;
            }
          }
        }
      });
    }, { rootMargin: '-80px 0px -60% 0px' });
    cards.forEach(function (c) { observer.observe(c); });
  }

  // ── No additional vanilla JS needed for the mobile sidebar ──
  // Alpine owns the open/close state via @click on the toggle button,
  // @click.away on the sidebar, and @keydown.escape.window on <body>.
});
