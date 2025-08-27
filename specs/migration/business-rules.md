# Business Rule Extraction (Baseline)

Source files:
- [product_service.go](file:///Users/isurufonseka/grab/example-ecom-go-api/internal/services/product_service.go)
- [cart_service.go](file:///Users/isurufonseka/grab/example-ecom-go-api/internal/services/cart_service.go)
- [order_service.go](file:///Users/isurufonseka/grab/example-ecom-go-api/internal/services/order_service.go)
- [rules.go](file:///Users/isurufonseka/grab/example-ecom-go-api/internal/services/rules.go)

| Service | Method | Source | Business logic? | Rules to preserve |
|---|---|---|---|---|
| Product | ListProducts | [product_service.go#L18-L21](file:///Users/isurufonseka/grab/example-ecom-go-api/internal/services/product_service.go#L18-L21) | No | — |
| Product | CreateProduct | [product_service.go#L34-L41](file:///Users/isurufonseka/grab/example-ecom-go-api/internal/services/product_service.go#L34-L41) | Yes | validateProductInput applies: title trimmed non-empty ≤200; author trimmed non-empty; description ≤2000; price 0.01–10000; stock 0–10000 ([product_service.go#L23-L31](file:///Users/isurufonseka/grab/example-ecom-go-api/internal/services/product_service.go#L23-L31)) |
| Product | GetProduct | [product_service.go#L43-L46](file:///Users/isurufonseka/grab/example-ecom-go-api/internal/services/product_service.go#L43-L46) | No | — |
| Product | UpdateProduct | [product_service.go#L48-L55](file:///Users/isurufonseka/grab/example-ecom-go-api/internal/services/product_service.go#L48-L55) | Yes | Same as CreateProduct validateProductInput |
| Product | DeleteProduct | [product_service.go#L57-L63](file:///Users/isurufonseka/grab/example-ecom-go-api/internal/services/product_service.go#L57-L63) | Yes | Block delete if product appears in any cart |
| Cart | AddToCart | [cart_service.go#L15-L46](file:///Users/isurufonseka/grab/example-ecom-go-api/internal/services/cart_service.go#L15-L46) | Yes | - Product not discontinued. - Max distinct items ≤3. - Per-line max ≤5. - If stock<3 then per-user max 1. - Requested qty ≤ stock. - Max total items in cart ≤10. - Cart risk total after add ≤5000. Thresholds in [rules.go#L3-L12](file:///Users/isurufonseka/grab/example-ecom-go-api/internal/services/rules.go#L3-L12). |
| Cart | RemoveFromCart | [cart_service.go#L49-L52](file:///Users/isurufonseka/grab/example-ecom-go-api/internal/services/cart_service.go#L49-L52) | No | — |
| Cart | GetCart | [cart_service.go#L54-L57](file:///Users/isurufonseka/grab/example-ecom-go-api/internal/services/cart_service.go#L54-L57) | No | — |
| Order | PlaceOrder | [order_service.go#L16-L66](file:///Users/isurufonseka/grab/example-ecom-go-api/internal/services/order_service.go#L16-L66) | Yes | - Duplicate order window 5s. - Cart not empty. - Special items: not mixed; qty must be 1. - Price drift: unit price must equal current. - Stock availability. - Total ≥5. - Daily user spend cap ≤10000. - If total>3000, status PENDING_REVIEW. Thresholds in [rules.go#L3-L12](file:///Users/isurufonseka/grab/example-ecom-go-api/internal/services/rules.go#L3-L12). |

Notes
- This table is the human-readable baseline. Phase 0 will also generate a machine-readable AST and rules JSON used by validation.
