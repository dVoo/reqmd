;; Tree-sitter query for top-level Go symbols.
;;
;; Captures:
;;   @func.name / @func.decl        - top-level functions
;;   @method.name / @method.decl    - methods on a receiver (receiver text
;;                                    is fetched in Go via ChildByFieldName
;;                                    because field-anchored patterns don't
;;                                    compile under all binding versions)
;;   @type.name / @type.decl /
;;     @type.body                    - type declarations
;;   @const.name / @const.decl      - const declarations
;;   @var.name / @var.decl          - var declarations

;; Top-level function declarations
(function_declaration
  name: (identifier) @func.name) @func.decl

;; Method declarations: functions with a receiver.
(method_declaration
  name: (field_identifier) @method.name) @method.decl

;; Type declarations: structs, interfaces, aliases, defined types.
;; Captures every type_spec inside the block so a single
;; `type ( A int; B string )` group yields two @type.decl matches.
(type_declaration
  (type_spec
    name: (type_identifier) @type.name)) @type.decl

;; Const declarations: `const X = 1` and grouped `const ( ... )`.
(const_declaration
  (const_spec
    name: (identifier) @const.name)) @const.decl

;; Var declarations: same shape as const.
(var_declaration
  (var_spec
    name: (identifier) @var.name)) @var.decl

;; Package clause: the first non-comment child of the source file.
;; Captures the package name; used to populate Symbol.Package.
(source_file
  (package_clause
    (package_identifier) @pkg.name)) @pkg.decl
