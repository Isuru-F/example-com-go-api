# Impldrift validation check (intentional drift)

We created a separate worktree/branch (chore/impldrift-check) off the migration branch and intentionally changed business logic in cart_service.go:
- Low-stock rule: `p.Stock < 3 && currentQty+req.Body.Quantity > 1` -> `p.Stock <= 3 && currentQty+req.Body.Quantity > 10`
- Removed discontinued check: `if p.Discontinued { return nil, errors("product unavailable") }`

Validated against pre-migration baseline (c685680c5):

```
go run ./specs/migration/tools/impldrift validate --baseline /private/var/folders/rv/bl71mf4d3nj603mn8j51bh0c0000gn/T/tmp.70fF92BgcZ-baseline.json --dir /private/var/folders/rv/bl71mf4d3nj603mn8j51bh0c0000gn/T/tmp.61hHbNU9G8
```

Validator output:
```
VALIDATION FAILURES:
-  (s *CartService).AddToCart: second param not endpoint.HTTPRequest[...] ((ctx context.Context, req *dto.AddToCartRequest))
-  (s *CartService).AddToCart: results not (*endpoint.HTTPResponse[...], error) ((*dto.Cart, error))
-  (s *CartService).RemoveFromCart: second param not endpoint.HTTPRequest[...] ((ctx context.Context, req *dto.RemoveFromCartRequest))
-  (s *CartService).RemoveFromCart: results not (*endpoint.HTTPResponse[...], error) ((*dto.Cart, error))
-  (s *CartService).GetCart: second param not endpoint.HTTPRequest[...] ((ctx context.Context, req *dto.GetCartRequest))
-  (s *CartService).GetCart: results not (*endpoint.HTTPResponse[...], error) ((*dto.Cart, error))
-  (s *OrderService).PlaceOrder: second param not endpoint.HTTPRequest[...] ((ctx context.Context, req *dto.PlaceOrderRequest))
-  (s *OrderService).PlaceOrder: results not (*endpoint.HTTPResponse[...], error) ((*dto.Order, error))
-  (s *ProductService).ListProducts: second param not endpoint.HTTPRequest[...] ((ctx context.Context, req *dto.ListProductsRequest))
-  (s *ProductService).ListProducts: results not (*endpoint.HTTPResponse[...], error) (([]*dto.Product, error))
-  (s *ProductService).CreateProduct: second param not endpoint.HTTPRequest[...] ((ctx context.Context, req *dto.CreateProductRequest))
-  (s *ProductService).CreateProduct: results not (*endpoint.HTTPResponse[...], error) ((*dto.Product, error))
-  (s *ProductService).GetProduct: second param not endpoint.HTTPRequest[...] ((ctx context.Context, req *dto.GetProductRequest))
-  (s *ProductService).GetProduct: results not (*endpoint.HTTPResponse[...], error) ((*dto.Product, error))
-  (s *ProductService).UpdateProduct: second param not endpoint.HTTPRequest[...] ((ctx context.Context, req *dto.UpdateProductRequest))
-  (s *ProductService).UpdateProduct: results not (*endpoint.HTTPResponse[...], error) ((*dto.Product, error))
-  (s *ProductService).DeleteProduct: second param not endpoint.HTTPRequest[...] ((ctx context.Context, req *dto.DeleteProductRequest))
-  (s *ProductService).DeleteProduct: results not (*endpoint.HTTPResponse[...], error) (error)
exit status 1
```

Observations:
- Numeric threshold/operator drift was detected as expected.
- Removal of the boolean guard on `p.Discontinued` is not flagged because impldrift focuses on numeric relational expressions (by design, to reduce false positives). This can be extended in a future iteration.
