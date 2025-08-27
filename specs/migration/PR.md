# Migrate services to endpoint-style generics; add tree-sitter validation (impldrift); update tests/handlers

## Signatures before/after (core services)

ProductService
- Before: `func (s *ProductService) CreateProduct(ctx context.Context, req *dto.CreateProductRequest) (*dto.Product, error)`
- After:  `func (s *ProductService) CreateProduct(ctx context.Context, req *endpoint.HTTPRequest[*dto.CreateProductRequest]) (*endpoint.HTTPResponse[*dto.Product], error)`

CartService
- Before: `func (s *CartService) AddToCart(ctx context.Context, req *dto.AddToCartRequest) (*dto.Cart, error)`
- After:  `func (s *CartService) AddToCart(ctx context.Context, req *endpoint.HTTPRequest[*dto.AddToCartRequest]) (*endpoint.HTTPResponse[*dto.Cart], error)`

OrderService
- Before: `func (s *OrderService) PlaceOrder(ctx context.Context, req *dto.PlaceOrderRequest) (*dto.Order, error)`
- After:  `func (s *OrderService) PlaceOrder(ctx context.Context, req *endpoint.HTTPRequest[*dto.PlaceOrderRequest]) (*endpoint.HTTPResponse[*dto.Order], error)`

## 1) What changed and why (high level)
- Updated all service method signatures to use endpoint-style generics (request/response), and minimally adjusted handlers and tests to call the new API.
- Added a tree-sitter based CLI (impldrift) to validate signatures, detect business-rule drift, and check API surface scope.
- Centralized migration specs and tools under specs/migration/.

Rationale: bring services in line with Go 1.24 era endpoint-style generic signatures while preserving business logic and ensuring verifiable, automated validation.

## 2) Project goal (re-stated)
Migrate service method signatures from legacy `*dto.Request`/`*dto.Response` to:
- `func (s *X) M(ctx context.Context, req *endpoint.HTTPRequest[*dto.Req]) (*endpoint.HTTPResponse[*dto.Resp], error)`

Scope-limited to signature change and necessary handler/test adaptations, with no business-logic change.

## 3) Why I’m confident this migration is successful (evidence)

### Build output
```
ecom-book-store-sample-api/internal/endpoint

```

### Test output
```
?   	ecom-book-store-sample-api/cmd	[no test files]
?   	ecom-book-store-sample-api/internal/dto	[no test files]
?   	ecom-book-store-sample-api/internal/endpoint	[no test files]
ok  	ecom-book-store-sample-api/internal/handlers	(cached)
?   	ecom-book-store-sample-api/internal/models	[no test files]
ok  	ecom-book-store-sample-api/internal/services	(cached)
?   	ecom-book-store-sample-api/internal/storage	[no test files]
?   	ecom-book-store-sample-api/specs/migration/tools/impldrift	[no test files]
```

### Tree-sitter validation (pre-migration baseline vs current)
Command: go run ./specs/migration/tools/impldrift validate --baseline <pre-migration-baseline.json> --dir .

Result:
```
validation OK
```

Baseline exports (from pre-migration c685680c5):
```json
[
  "CartService.AddToCart",
  "CartService.GetCart",
  "CartService.RemoveFromCart",
  "OrderService.PlaceOrder",
  "ProductService.CreateProduct",
  "ProductService.DeleteProduct",
  "ProductService.GetProduct",
  "ProductService.ListProducts",
  "ProductService.UpdateProduct"
]
```

Current exports (after migration):
```json
[
  "CartService.AddToCart",
  "CartService.GetCart",
  "CartService.RemoveFromCart",
  "OrderService.PlaceOrder",
  "ProductService.CreateProduct",
  "ProductService.DeleteProduct",
  "ProductService.GetProduct",
  "ProductService.ListProducts",
  "ProductService.UpdateProduct"
]
```

Baseline rule counts (by method):
```json
[
  {"method":"(s *CartService).AddToCart","count":7},
  {"method":"(s *CartService).GetCart","count":0},
  {"method":"(s *CartService).RemoveFromCart","count":0},
  {"method":"(s *OrderService).PlaceOrder","count":9},
  {"method":"(s *ProductService).CreateProduct","count":0},
  {"method":"(s *ProductService).DeleteProduct","count":0},
  {"method":"(s *ProductService).GetProduct","count":0},
  {"method":"(s *ProductService).ListProducts","count":0},
  {"method":"(s *ProductService).UpdateProduct","count":0}
]
```

Current rule counts (by method):
```json
[
  {"method":"(s *CartService).AddToCart","count":7},
  {"method":"(s *CartService).GetCart","count":0},
  {"method":"(s *CartService).RemoveFromCart","count":0},
  {"method":"(s *OrderService).PlaceOrder","count":9},
  {"method":"(s *ProductService).CreateProduct","count":0},
  {"method":"(s *ProductService).DeleteProduct","count":0},
  {"method":"(s *ProductService).GetProduct","count":0},
  {"method":"(s *ProductService).ListProducts","count":0},
  {"method":"(s *ProductService).UpdateProduct","count":0}
]
```

### How the business logic was checked (process)
1. Extracted a baseline from the pre-migration commit (c685680c5) using the impldrift `extract` command. This produced a machine-readable AST snapshot (methods, exported APIs, and business-rule comparisons filtered to numeric-threshold comparisons; constants resolved from internal/services/rules.go).
2. Ran `validate` against the current workspace using that baseline. The validator:
   - Ensures each service method uses the new signature: second param is `endpoint.HTTPRequest[...]`, returns `*endpoint.HTTPResponse[...]` plus `error`.
   - Compares exported method sets (scope check) — no new exports added.
   - Compares business-rule guard expressions (operator tokens and right-hand thresholds after constant/literal normalization, and normalizes request paths like `.Body.` to avoid false drift).

Raw outputs above (exports and rule counts per method) are included for independent review.

### How the CLI tool was created and what it does
- Location: specs/migration/tools/impldrift (Go + tree-sitter-go).
- Commands:
  - `extract --dir <src> --out <file>`: builds AST snapshot with methods, exported names, resolved constants, and filtered comparison expressions.
  - `validate --baseline <file> --dir <src>`: loads baseline, re-extracts current, checks signature shape, exported-surface drift, and rule drift.
- Confidence: The tool operates directly on AST using the official tree-sitter Go grammar and resolves constants from rules.go; it filters for relational guards and numeric thresholds (constants or literals), normalizes request path changes introduced by the migration (`.Body.`), and reports exact drifts with file/line anchoring if found.

## Directory changes
- Tools: specs/migration/tools/
  - impldrift (source) and impldrift-bin (local binary copy)
  - ast-baseline.json (baseline)
- Services/handlers/tests updated to compile under new signatures.

## How to reproduce locally
- Build: `go build ./...`
- Test: `go test ./...`
- Validate (using saved baseline):
  - `go run ./specs/migration/tools/impldrift validate --baseline specs/migration/tools/ast-baseline.json --dir .`
- Validate against pre-migration (on demand):
  - `go run ./specs/migration/tools/impldrift extract --dir $(mktemp -d)/<checkout@c685680c5> --out /tmp/pre.json`
  - `go run ./specs/migration/tools/impldrift validate --baseline /tmp/pre.json --dir .`


## Impldrift sanity check
- See [specs/migration/impldrift_check.md](specs/migration/impldrift_check.md) for an intentional drift test and the validator output.


## Tools used to verify (no assumptions)
- Build & Test: `go build ./...`, `go test ./...` (outputs included above).
- AST validation: `impldrift` (tree-sitter-go) — baseline extracted from c685680c5, and validated current and intentionally modified worktree.
- Raw data included: exported method sets and per-method rule counts (pretty-printed), and the full validator output from the drift test.


- Per-method AST comparison appended to [impldrift_check.md](specs/migration/impldrift_check.md) (deterministic, method-by-method).


**Direct report link:** https://github.com/Isuru-F/example-com-go-api/blob/chore/migrate-endpoint-generics-with-impldrift/specs/migration/impldrift_check.md
