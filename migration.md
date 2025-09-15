# Service Method Signature Migration Plan

## Overview
Migrate service method signatures from Go 1.10 style to modern generic endpoint pattern while maintaining identical business logic.

**Signature Change:**
- **BEFORE:** `func (s *SampleService) Hello(ctx context.Context, req *dto.HelloRequest) (*dto.HelloResponse, error)`
- **AFTER:** `func (s *SampleService) Hello(ctx context.Context, req *endpoint.HTTPRequest[*sampleService.HelloRequest]) (*endpoint.HTTPResponse[*grabservicepb.HelloResponse], error)`

## Affected Services
- **ProductService:** ListProducts, CreateProduct, GetProduct, UpdateProduct, DeleteProduct
- **CartService:** AddToCart, RemoveFromCart, GetCart  
- **OrderService:** PlaceOrder

## Phase Structure

### Phase 0: Baseline Establishment (Pre-migration)
**Branch:** `v0-pre-sig` (tagged)
**Goals:**
- Freeze current codebase state
- Extract and catalog all business logic rules
- Establish validation baseline

**Tasks:**
- [ ] Tag current state as `v0-pre-sig`
- [ ] Create `tools/logicmap` CLI tool for business rule extraction
- [ ] Run baseline extraction: `go run ./tools/logicmap --mode=extract --pkg=internal/services`
- [ ] Commit `logic_maps/baseline.json`
- [ ] Generate `docs/business-rules.md` from extracted rules
- [ ] Verify all tests pass: `make test && go vet ./...`

### Phase 1: Introduce Generic Transport Types
**Goals:**
- Add new endpoint types without breaking existing code
- Create shim wrappers for dual compatibility
- Ensure build continues to pass

**Tasks:**
- [ ] Create `internal/endpoint` package with generic types:
  - `HTTPRequest[T any]` struct
  - `HTTPResponse[T any]` struct
- [ ] Generate shim wrappers for all service methods:
  - Keep legacy methods (delegate to core implementation)
  - Add new `...EP` methods with new signatures
  - Extract shared logic to `...Core` methods
- [ ] Add build tag support for gradual migration
- [ ] Verify: `make deps && make build && make test && go vet ./...`

### Phase 2: Handler Migration
**Goals:**
- Update handlers to use new endpoint signatures
- Maintain backward compatibility via build tags

**Tasks:**
- [ ] Update handlers under `phase2` build tag to use new signatures
- [ ] Keep legacy handler code behind `!phase2` build tag  
- [ ] Test both paths work correctly
- [ ] Verify: `go build -tags phase2 && go test -tags phase2 ./...`

### Phase 3: Service-by-Service Migration
**Goals:**
- Complete signature migration for each service
- Remove shim wrappers
- Update all tests

**Tasks:**
- [ ] **Phase 3a:** ProductService
  - Delete legacy method signatures
  - Rename `...EP` methods to final names
  - Update unit tests
- [ ] **Phase 3b:** CartService  
  - Same process as ProductService
- [ ] **Phase 3c:** OrderService
  - Same process as ProductService
- [ ] Verify all tests pass after each service migration

### Phase 4: Cleanup
**Goals:**
- Remove all legacy code and shims
- Clean up dependencies

**Tasks:**
- [ ] Remove shim methods and core implementations
- [ ] Clean up unused DTO references
- [ ] Remove build tags
- [ ] Run `go mod tidy`
- [ ] Verify final state: `make deps && make build && make test && go vet ./...`

### Phase 5: Post-Migration Validation
**Goals:**
- Validate business logic integrity
- Tag final migration state

**Tasks:**
- [ ] Run business logic validation: `go run ./tools/bizlint ci`
- [ ] Compare against baseline rules
- [ ] Verify no new methods added (scope validation)
- [ ] Tag final state as `v1-sig-generic`

## Validation Strategy

### Business Logic Drift Detection
**Tool:** `tools/bizlint` CLI

**Capabilities:**
- Extract current business rules from AST
- Compare against baseline to detect drift
- Flag critical changes like `>100` becoming `<100` or `<98`
- Validate method scope (no new exports without approval)

**Usage:**
```bash
# Extract current rules
go run ./tools/bizlint extract

# Compare against baseline  
go run ./tools/bizlint diff --baseline=logic_maps/baseline.json

# Full CI validation
go run ./tools/bizlint ci --baseline=logic_maps/baseline.json --pkg=internal/services
```

### Key Business Rules to Preserve

| Service | Method | Critical Rules |
|---------|--------|----------------|
| Product | CreateProduct/UpdateProduct | title/author length bounds, price range (0.01-10000), stock limits |
| Product | DeleteProduct | Reject if product in any cart |
| Cart | AddToCart | 9 validation rules including stock checks, quantity limits, cart totals |
| Order | PlaceOrder | Duplicate order prevention, stock validation, spending limits |

### Scope Validation
- No new exported service methods unless whitelisted
- Business logic must remain functionally identical
- Only signature-related changes permitted

## Directory Structure
```
migrations/
├── phase-01-add-endpoint/
│   ├── README.md
│   ├── tracker.yaml
│   └── patch.diff
├── phase-02-handlers/
├── phase-03a-product-service/
├── phase-03b-cart-service/
├── phase-03c-order-service/
├── phase-04-cleanup/
├── phase-05-validation/
└── scope-allowlist.txt

tools/
├── logicmap/
│   └── main.go
└── bizlint/
    └── main.go

logic_maps/
└── baseline.json

docs/
└── business-rules.md
```

## Success Criteria
- [ ] All phases complete with passing tests
- [ ] Zero business logic drift detected
- [ ] No new functionality added beyond signature changes
- [ ] Full CI/CD pipeline remains green
- [ ] Documentation updated and accurate

## Rollback Strategy
Each phase is tagged for easy rollback:
- Phase issues: `git checkout <previous-phase-tag>`
- Complete rollback: `git checkout v0-pre-sig`

## Risk Mitigation
- Automated business logic validation prevents drift
- Phased approach allows incremental validation
- Build tags maintain compatibility during transition
- Comprehensive test coverage at each phase
- Clear rollback points via git tags
