;; Tree-sitter query for top-level Rust symbols.
;;
;; Captures:
;;   @func.name / @func.decl      - free `fn` items at module level
;;   @method.name / @method.decl  - `fn` items inside an `impl` block and
;;                                  trait signatures/defaults inside a
;;                                  `trait`; receiver text is fetched in Go
;;                                  via the parent chain walk
;;   @type.name / @type.decl      - struct / enum / trait / union items and
;;                                  `type Foo = ...;` aliases
;;   @const.name / @const.decl    - `const` items
;;   @var.name / @var.decl        - `static` items
;;
;; A method's `function_item` matches BOTH the bare `function_item`
;; pattern (@func.decl) and the impl-anchored pattern (@method.decl).
;; The Go-side Parse() runs a two-pass dedup on node identity so methods
;; win over free functions (mirroring the Python plugin).

;; Free function items.
(function_item name: (identifier) @func.name) @func.decl

;; Methods: function items inside an impl block.
(impl_item
  body: (declaration_list
    (function_item
      name: (identifier) @method.name) @method.decl))

;; Trait methods: signatures (no body) and default implementations
;; (with a body) inside a trait's declaration list.
(trait_item
  body: (declaration_list
    (function_signature_item
      name: (identifier) @method.name) @method.decl))
(trait_item
  body: (declaration_list
    (function_item
      name: (identifier) @method.name) @method.decl))

;; Type items.
(struct_item name: (type_identifier) @type.name) @type.decl
(enum_item name: (type_identifier) @type.name) @type.decl
(trait_item name: (type_identifier) @type.name) @type.decl
(union_item name: (type_identifier) @type.name) @type.decl
(type_item name: (type_identifier) @type.name) @type.decl

;; Constants and statics.
(const_item name: (identifier) @const.name) @const.decl
(static_item name: (identifier) @var.name) @var.decl
