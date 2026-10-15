;; Tree-sitter query for top-level C++ symbols.
;;
;; Captures:
;;   @func.name / @func.decl        - free function definitions
;;   @method.name / @method.decl    - class/struct member functions (inline
;;                                    definitions and in-class declarations);
;;                                    receiver text is fetched in Go via the
;;                                    parent chain walk because the method
;;                                    node does not carry its owner
;;   @type.name / @type.decl        - class / struct / enum / union
;;                                    specifiers, `using` aliases, typedefs
;;   @var.name / @var.decl          - top-level variable declarations
;;   @const.name / @const.decl      - object-like #define macros
;;
;; A member function's name is a `field_identifier`; a free function's
;; name is an `identifier`. The two function patterns are therefore
;; mutually exclusive on the name node type, so inline methods are never
;; double-counted as free functions. Out-of-line member definitions
;; (`int Calculator::add(...) { ... }` at namespace scope) use a
;; qualified declarator and are not captured by either pattern.

;; Free function definitions.
(function_definition
  declarator: (function_declarator
    declarator: (identifier) @func.name)) @func.decl

;; Inline member function definitions: function_definition inside a class
;; body, whose declarator name is a field_identifier.
(function_definition
  declarator: (function_declarator
    declarator: (field_identifier) @method.name)) @method.decl

;; In-class member function declarations (`int sub(int a, int b);`).
(field_declaration
  declarator: (function_declarator
    declarator: (field_identifier) @method.name)) @method.decl

;; Types: class / struct / enum / union specifiers. The Go-side parent
;; walk drops specifiers nested inside a type_definition so a
;; `typedef struct Foo { ... } Foo;` emits a single symbol.
(class_specifier name: (type_identifier) @type.name) @type.decl
(struct_specifier name: (type_identifier) @type.name) @type.decl
(enum_specifier name: (type_identifier) @type.name) @type.decl
(union_specifier name: (type_identifier) @type.name) @type.decl

;; `using Foo = ...;` aliases.
(alias_declaration name: (type_identifier) @type.name) @type.decl

;; typedef aliases: the declarator is the alias being introduced.
(type_definition declarator: (type_identifier) @type.name) @type.decl

;; Top-level variables: `int x;` and `int x = 5;` (the latter wraps the
;; identifier in an init_declarator). Multiple declarators in one
;; declaration (`int a, b;`) produce one match per identifier.
(declaration declarator: (identifier) @var.name) @var.decl
(declaration
  declarator: (init_declarator
    declarator: (identifier) @var.name)) @var.decl

;; Object-like macros: `#define FOO 100`. Function-like macros
;; (preproc_function_def) are skipped.
(preproc_def name: (identifier) @const.name) @const.decl
