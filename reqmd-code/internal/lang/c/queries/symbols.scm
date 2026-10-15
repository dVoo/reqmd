;; Tree-sitter query for top-level C symbols.
;;
;; Captures:
;;   @func.name / @func.decl      - function definitions (with bodies)
;;   @type.name / @type.decl      - struct / enum / union specifiers and
;;                                  typedef aliases
;;   @var.name / @var.decl        - top-level variable declarations
;;   @const.name / @const.decl    - object-like #define macros
;;
;; Function prototypes (`int foo(void);` without a body) are not
;; captured: their node kind is `declaration`, and the @var.* patterns
;; below anchor on an `identifier` declarator, which a prototype's
;; `function_declarator` does not satisfy. A prototype plus its
;; definition in the same file would otherwise produce duplicate
;; requirement IDs (the ID hashes over package + name only).

;; Function definitions.
(function_definition
  declarator: (function_declarator
    declarator: (identifier) @func.name)) @func.decl

;; Types: struct, enum, and union specifiers. The body: anchor rejects
;; forward declarations (`struct Foo;`), which carry no definition. The
;; Go-side parent walk drops specifiers nested inside a type_definition
;; so a `typedef struct Foo { ... } Foo;` emits a single symbol.
(struct_specifier name: (type_identifier) @type.name body: (_)) @type.decl
(enum_specifier name: (type_identifier) @type.name body: (_)) @type.decl
(union_specifier name: (type_identifier) @type.name body: (_)) @type.decl

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
