# Implementation Drift Validator (impldrift)

Goal
- Ensure service method signatures use endpoint-style generics.
- Detect behavior drift in business rules (operators/thresholds).
- Enforce migration scope (no new exported service APIs).

Inputs
- Baseline: specs/migration/tools/ast-baseline.json (Phase 0 output)
- Current tree: `internal/services/**/*.go`, [rules.go](file:///Users/isurufonseka/grab/example-ecom-go-api/internal/services/rules.go)

Checks
1) Signature shape
- For each method in baseline under `internal/services`:
  - Must still exist with same receiver type and name.
  - Params: exactly 2, the second is `endpoint.HTTPRequest[*<Dto>]` (or `compat.HTTPRequest[*<Dto>]` if shim retained).
  - Returns: first `*endpoint.HTTPResponse[*<Resp>]` (or compat equivalent), second `error`.

2) Business rule drift
- Traverse method bodies for those marked “business logic” in [business-rules.md](./business-rules.md):
  - Extract comparisons and logical guards (`<, <=, >, >=, ==, !=`, bool ops).
  - Resolve identifiers against constants from [rules.go](file:///Users/isurufonseka/grab/example-ecom-go-api/internal/services/rules.go).
  - Compare operator tokens and resolved numeric values to baseline JSON. If a threshold or operator differs (e.g., `> 100` → `< 98`), flag BLOCKER.

3) Scope check
- List new exported methods in `internal/services`. If not present in baseline, fail.

Output
- JSON report at `specs/migration/tools/report.json` with sections: signatures, rules, scope.
- Exit codes: 0 ok; 1 violations; 2 internal error.

CLI
- `go run ./specs/migration/tools/impldrift extract --out specs/migration/tools/ast-baseline.json`
- `go run ./specs/migration/tools/impldrift validate --baseline specs/migration/tools/ast-baseline.json`

Implementation notes
- Use tree-sitter-go for parsing; walk AST to collect function decls, parameters, return types, and binary expressions.
- Include a resolver to fold const identifiers from rules.go into literal values for comparison.
- Keep reporting deterministic and file/line anchored.
