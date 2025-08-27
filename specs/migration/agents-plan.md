# Subagent Orchestration

- ParserAgent (tree-sitter-go)
  - Tasks: AST inventory, method signatures, constant extraction, rules JSON.
  - Outputs: tools/baseline/ast-baseline.json; enriches business-rules.md.

- MigrationAgent
  - Tasks: change service method signatures to endpoint generics; replace `req.*` with `req.Body.*`; wrap responses.
  - Scope: [product_service.go](file:///Users/isurufonseka/grab/example-ecom-go-api/internal/services/product_service.go), [cart_service.go](file:///Users/isurufonseka/grab/example-ecom-go-api/internal/services/cart_service.go), [order_service.go](file:///Users/isurufonseka/grab/example-ecom-go-api/internal/services/order_service.go)

- TestAgent
  - Tasks: adapt unit tests and handlers to use `WrapReq/UnwrapResp`; keep assertions unchanged.
  - Scope: [product_service_test.go](file:///Users/isurufonseka/grab/example-ecom-go-api/internal/services/product_service_test.go), [cart_service_test.go](file:///Users/isurufonseka/grab/example-ecom-go-api/internal/services/cart_service_test.go), [order_service_test.go](file:///Users/isurufonseka/grab/example-ecom-go-api/internal/services/order_service_test.go), handler tests in [handlers_test.go](file:///Users/isurufonseka/grab/example-ecom-go-api/internal/handlers/handlers_test.go)

- ValidationAgent
  - Tasks: build cmd/impldrift; signature, rule, scope checks; CI wiring.

- ScopeAgent
  - Tasks: compare exported service APIs vs baseline; ensure only necessary files changed.
