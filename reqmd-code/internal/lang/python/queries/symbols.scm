;; Tree-sitter query for top-level Python symbols.
;;
;; Captures:
;;   @func.decl / @func.name     - top-level functions (def and async def)
;;   @method.decl / @method.name - methods (function definitions inside class bodies)
;;   @type.decl / @type.name     - classes
;;   @const.decl / @const.name   - module-level UPPER_CASE identifier assignments
;;
;; The same `function_definition` node type is used for both plain
;; `def` and `async def` in tree-sitter-python; the Go-side dispatch
;; does not differentiate between them.
;;
;; Decorated definitions (`@decorator\ndef foo(): ...`) are unwrapped
;; by the field-anchored pattern: we capture the inner
;; function_definition / class_definition, not the decorated_definition
;; wrapper. This avoids double-counting when the same node is also
;; matched by the undecorated pattern.
;;
;; Method vs. top-level function is distinguished by parent anchoring:
;; the @method.* captures are only produced when the function_definition
;; sits inside a class body (block of a class_definition). The
;; Go-side Parse() then sees a match with @method.decl and treats it
;; as a method; matches with @func.decl (and no @method.decl) are
;; top-level functions.

;; Top-level function definitions: covers both `def` and `async def`.
;; Matches every function_definition in the source; the Go-side parent
;; walk disambiguates "inside a class body" (method) from "inside a
;; module" (function). The decorated_definition pattern below emits a
;; second @func.decl for the same inner node, but Go-side dedup on
;; node identity collapses them.
(function_definition
  name: (identifier) @func.name) @func.decl

;; Methods: function definitions nested inside a class body.
;; Anchoring on the class body via field-anchored patterns keeps the
;; @method.decl capture pointing at the function_definition (not the
;; wrapping class_definition), so the plugin can read the method's
;; own line span and parameter list.
(class_definition
  body: (block
    (function_definition
      name: (identifier) @method.name) @method.decl))

;; Class definitions.
(class_definition
  name: (identifier) @type.name) @type.decl

;; Decorated class definitions: unwrap the decorator wrapper so we
;; don't double-count. The `@type.decl` capture is the inner
;; class_definition, not the decorated_definition node.
(decorated_definition
  (class_definition
    name: (identifier) @type.name) @type.decl)

;; Module-level constant assignments: `FOO = 1` etc. The `assignment`
;; node is wrapped in an `expression_statement`; the LHS is anchored
;; to `identifier` (a subtype of `pattern`) so tuple/list destructuring
;; and attribute targets are excluded. The Go-side parent walk drops
;; matches whose parent chain ends at a class body (those are
;; class-level attributes, not module-level constants).
(expression_statement
  (assignment
    left: (identifier) @const.name)) @const.decl
