# Phase 3 Report — CartService migration

Changed methods (new signatures):
- CartService.AddToCart(ctx context.Context, req *endpoint.HTTPRequest[*dto.AddToCartRequest]) (*endpoint.HTTPResponse[*dto.Cart], error)
- CartService.RemoveFromCart(ctx context.Context, req *endpoint.HTTPRequest[*dto.RemoveFromCartRequest]) (*endpoint.HTTPResponse[*dto.Cart], error)
- CartService.GetCart(ctx context.Context, req *endpoint.HTTPRequest[*dto.GetCartRequest]) (*endpoint.HTTPResponse[*dto.Cart], error)

Handlers updated:
- [cart_handler.go](file:///Users/isurufonseka/grab/example-ecom-go-api/internal/handlers/cart_handler.go) now wraps/unwraps endpoint requests/responses.

Tests updated:
- [cart_service_test.go](file:///Users/isurufonseka/grab/example-ecom-go-api/internal/services/cart_service_test.go)
- Cross-tests using CartService in [order_service_test.go](file:///Users/isurufonseka/grab/example-ecom-go-api/internal/services/order_service_test.go) and [product_service_test.go](file:///Users/isurufonseka/grab/example-ecom-go-api/internal/services/product_service_test.go)

Test results (phase end):
- go test ./... — PASS

Validation summary for Phase 3:
- Business rules retained (by inspection):
  - MaxDistinctCartItems (3), MaxQuantityPerLineItem (5), MaxTotalItemsInCart (10), CartRiskLimitTotal (5000.0), low-stock rule (<3 => max 1).
  - No operator or threshold changes detected in diff.
- Scope: only services/handlers/tests updated to adjust signatures. No new exported methods added.

Items for your review:
- Verify that the above three signatures match the target pattern for endpoint HTTP generics.
- Spot-check rule guards in [cart_service.go](file:///Users/isurufonseka/grab/example-ecom-go-api/internal/services/cart_service.go) around lines 15–46 for unchanged semantics.
