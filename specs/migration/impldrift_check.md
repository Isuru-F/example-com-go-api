# Impldrift validation check (intentional drift)

We created a separate worktree/branch (chore/impldrift-check) off the migration branch and intentionally changed business logic in cart_service.go:
- Low-stock rule: `p.Stock < 3 && currentQty+req.Body.Quantity > 1` -> `p.Stock <= 3 && currentQty+req.Body.Quantity > 10`
- Removed discontinued check: `if p.Discontinued { return nil, errors("product unavailable") }`

Validated against pre-migration baseline (c685680c5):

```
go run ./specs/migration/tools/impldrift validate --baseline /private/var/folders/rv/bl71mf4d3nj603mn8j51bh0c0000gn/T/tmp.61hHbNU9G8-baseline.json --dir /private/var/folders/rv/bl71mf4d3nj603mn8j51bh0c0000gn/T/tmp.c3vWuX1u7V
```

Validator output:
```
VALIDATION FAILURES:
-  rule drift (s *CartService).AddToCart[2]: p.Stock < 3 -> p.Stock <= 3
-  rule drift (s *CartService).AddToCart[3]: currentQty+req.Quantity > 1 -> currentQty+req.Body.Quantity > 10
exit status 1
```

Observations:
- Numeric threshold/operator drift was detected (<= vs <; 10 vs 1) in AddToCart.
- Removal of the boolean `p.Discontinued` check is not flagged because impldrift currently filters to numeric relational comparisons. This is a known limitation we can address if required.


## Deterministic per-method verification (migration vs baseline)

| Method | Signature/Scope | Rule count (baseline -> current) |
|---|---|---|
| . | OK (validate) | 0 -> 0 |


## Per-method rule comparison (keys union)

| Method | Rule count (baseline -> current) |
|---|---|
| (s *CartService).AddToCart | 7 -> 7 |
| (s *CartService).GetCart | 0 -> 0 |
| (s *CartService).RemoveFromCart | 0 -> 0 |
| (s *OrderService).PlaceOrder | 9 -> 9 |
| (s *ProductService).CreateProduct | 0 -> 0 |
| (s *ProductService).DeleteProduct | 0 -> 0 |
| (s *ProductService).GetProduct | 0 -> 0 |
| (s *ProductService).ListProducts | 0 -> 0 |
| (s *ProductService).UpdateProduct | 0 -> 0 |
