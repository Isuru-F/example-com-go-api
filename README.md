# ecom-book-store-sample-api

Simple Go (1.24) ecommerce sample API for a book store. Demonstrates pre-context migration method signatures in services (no `context.Context` parameters). Uses Gin and in-memory storage.

## Run

- Build: `make build`
- Run: `make run` (serves on `:8080`)
- Test: `make test`
- Quick API smoke test: `make test-api`

## API

Base: `/api/v1`

Products:
- GET `/products` — list
- POST `/products` — create
- GET `/products/:id` — get
- PUT `/products/:id` — update
- DELETE `/products/:id` — delete

Product payload supports optional flags:
- `discontinued` (bool) — unavailable for adding to cart
- `isSpecial` (bool) — must be ordered alone with quantity 1

Cart:
- POST `/cart/user/:id/items` — add item `{ "productId": 1, "quantity": 2 }`
- DELETE `/cart/user/:id/items` — remove item `{ "productId": 1 }`

Orders:
- POST `/orders/user/:id` — place order from the user's cart

## Business rules

Enforced in services (400/409 errors via handlers):

Cart
- Max distinct items per cart: 3
- Max quantity per line item: 5
- Max total items (sum of quantities): 10
- Cart total risk cap: ≤ 5000 (on add and checkout)
- Do not exceed available stock on add
- Low-stock rule: if product stock < 3, limit to quantity 1 per cart
- Discontinued products cannot be added

Order
- Minimum order amount: ≥ 5.00
- High-value review: if total > 3000, order status = `PENDING_REVIEW`
- Duplicate checkout guard: reject orders placed within 5s of previous order for the same user
- Price drift protection: if product price changed since item was added to cart, reject and ask to refresh cart
- Special items must be purchased alone with quantity 1
- Daily user spend cap: sum of today’s orders per user must not exceed 10000

Product
- Title required (≤ 200 chars), author required, description ≤ 2000 chars
- Price in [0.01, 10000]
- Stock in [0, 10000]
- Prevent deleting a product that exists in any user cart

## Rate limits (demo-only, in-memory, per-process)
- Cart add/remove: ≤ 10 ops per user per minute → 429 Too Many Requests
- Product create/update: ≤ 5 ops per minute (global) → 429 Too Many Requests

## Example

```bash
curl -s http://localhost:8080/api/v1/products | jq . | head

curl -s -X POST http://localhost:8080/api/v1/products \
 -H 'Content-Type: application/json' \
 -d '{"title":"New Book","author":"Anon","description":"Desc","price":24.99,"stock":10}'

curl -s -X POST http://localhost:8080/api/v1/cart/user/1/items \
 -H 'Content-Type: application/json' -d '{"productId":1,"quantity":2}'

curl -s -X POST http://localhost:8080/api/v1/orders/user/1 | jq .
```

## Notes

- Storage and service method signatures intentionally omit `context.Context` to model a codebase before framework migration.
- Data is in-memory with auto-increment IDs; restarting resets state.
- Rate limiting and business rules are in-memory and for demo only; disable or replace in production.

