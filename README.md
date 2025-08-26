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

Cart:
- POST `/cart/user/:id/items` — add item `{ "productId": 1, "quantity": 2 }`
- DELETE `/cart/user/:id/items` — remove item `{ "productId": 1 }`

Orders:
- POST `/orders/user/:id` — place order from the user's cart

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
