# Impldrift validation check (intentional drift)

We created a separate worktree/branch (chore/impldrift-check) off the migration branch and intentionally changed business logic in cart_service.go:
- Low-stock rule: `p.Stock < 3 && currentQty+req.Body.Quantity > 1` -> `p.Stock <= 3 && currentQty+req.Body.Quantity > 10`
- Removed discontinued check: `if p.Discontinued { return nil, errors("product unavailable") }`

Validated against pre-migration baseline (c685680c5):

```
go run ./specs/migration/tools/impldrift validate --baseline None --dir None
```

Validator output:
```
None
```

Observations:
- Numeric threshold/operator drift was detected as expected.
- Removal of the boolean guard on `p.Discontinued` is not flagged because impldrift focuses on numeric relational expressions (by design, to reduce false positives). This can be extended in a future iteration.
