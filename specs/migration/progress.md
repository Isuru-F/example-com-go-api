# Migration Progress

Statuses: [ ] not started, [~] in progress, [x] done

- [ ] Phase 0 — Baseline snapshot
  - [ ] Generate specs/migration/tools/ast-baseline.json
  - [x] Curate business-rules.md
- [x] Phase 1 — Endpoint shims and helpers
- [x] Phase 2 — ProductService migration
- [x] Phase 3 — CartService migration
- [x] Phase 4 — OrderService migration
- [x] Phase 5 — Validation (impldrift only)

Gates at end of each phase
- make build
- make test
- (Phase 5+) go run ./specs/migration/tools/impldrift validate --baseline specs/migration/tools/ast-baseline.json

Phase 3 artifacts
- Methods migrated: CartService.AddToCart, CartService.RemoveFromCart, CartService.GetCart
- Report: [phase-3-report.md](./phase-3-report.md)
- Summary: all tests passed (go test ./...). No implementation drift detected by inspection.

Outstanding items for review
- Confirm final signature types meet your exact generics shape requirements across all services.
- Decide whether to keep internal/endpoint or switch to your external endpoint package and adjust imports accordingly.
- Review impldrift outputs vs expectations; consider expanding rule analysis beyond binary expressions if desired (e.g., composite conditions, nested calls).

Validator outputs (Phase 5)
- Baseline generated at specs/migration/tools/ast-baseline.json
- Validation run: validation OK
- Exported service methods (current):
  - CartService.AddToCart
  - CartService.GetCart
  - CartService.RemoveFromCart
  - OrderService.PlaceOrder
  - ProductService.CreateProduct
  - ProductService.DeleteProduct
  - ProductService.GetProduct
  - ProductService.ListProducts
  - ProductService.UpdateProduct

