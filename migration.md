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

### Phase 0: Tool Validation (Pre-migration Testing)
**Goals:**
- Validate existing tools work correctly with our codebase
- Test symbol extraction and rule drift detection
- Establish confidence in validation toolchain

**Tasks:**
- [ ] Test `parser-go` tool extracts expected symbols from service files
  - Run on `internal/services/product_service.go`
  - Verify it identifies all methods: `ListProducts`, `CreateProduct`, `GetProduct`, `UpdateProduct`, `DeleteProduct`
  - Check method signatures are correctly parsed
  - Validate receiver types are properly identified
- [ ] Test `bizlint-go` tool detects rule drift correctly
  - Create test service method with known business rules
  - Intentionally modify a rule (e.g., change `>100` to `<100`)
  - Verify `bizlint-go diff` detects the change and fails appropriately
  - Test scope validation catches new exported methods
- [ ] Validate tool integration with our codebase structure
  - Ensure tools work with our Go module setup
  - Test on all three service files (product, cart, order)
  - Confirm JSON output format matches expected schema
- [ ] Document tool usage patterns for our specific codebase
  - Create example commands and expected outputs
  - Document any configuration needed for our module structure

### Phase 1: Baseline Establishment
**Branch:** `v0-pre-sig` (tagged)
**Goals:**
- Freeze current codebase state
- Extract and catalog all business logic rules
- Establish validation baseline using validated tools

**Tasks:**
- [ ] Tag current state as `v0-pre-sig`
- [ ] Run baseline extraction using `bizlint-go extract --pkg=internal/services`
- [ ] Create `logic_maps/` directory and commit `baseline.json`
- [ ] Generate `docs/business-rules.md` from extracted rules
- [ ] Verify all tests pass: `make test && go vet ./...`
- [ ] Validate baseline can detect future changes by testing with a temporary modification

### Phase 2: Introduce Generic Transport Types
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

### Phase 3: Handler Migration
**Goals:**
- Update handlers to use new endpoint signatures
- Maintain backward compatibility via build tags

**Tasks:**
- [ ] Update handlers under `phase3` build tag to use new signatures
- [ ] Keep legacy handler code behind `!phase3` build tag  
- [ ] Test both paths work correctly
- [ ] Verify: `go build -tags phase3 && go test -tags phase3 ./...`

### Phase 4: Service-by-Service Migration
**Goals:**
- Complete signature migration for each service
- Remove shim wrappers
- Update all tests

**Tasks:**
- [ ] **Phase 4a:** ProductService
  - Delete legacy method signatures
  - Rename `...EP` methods to final names
  - Update unit tests
  - Run `bizlint-go ci` to validate no business logic drift
- [ ] **Phase 4b:** CartService  
  - Same process as ProductService
  - Run `bizlint-go ci` to validate no business logic drift
- [ ] **Phase 4c:** OrderService
  - Same process as ProductService
  - Run `bizlint-go ci` to validate no business logic drift
- [ ] Verify all tests pass after each service migration

### Phase 5: Cleanup
**Goals:**
- Remove all legacy code and shims
- Clean up dependencies

**Tasks:**
- [ ] Remove shim methods and core implementations
- [ ] Clean up unused DTO references
- [ ] Remove build tags
- [ ] Run `go mod tidy`
- [ ] Verify final state: `make deps && make build && make test && go vet ./...`

### Phase 6: Post-Migration Validation
**Goals:**
- Validate business logic integrity
- Tag final migration state

**Tasks:**
- [ ] Run final business logic validation: `bizlint-go ci --baseline=logic_maps/baseline.json --pkg=internal/services`
- [ ] Compare against baseline rules to ensure zero drift
- [ ] Verify no new methods added (scope validation)
- [ ] Generate final migration report
- [ ] Tag final state as `v1-sig-generic`

## Validation Strategy

### Business Logic Drift Detection
**Tool:** `bizlint-go` CLI (external tool)

**Capabilities:**
- Extract current business rules from AST using tree-sitter
- Compare against baseline to detect drift
- Flag critical changes like `>100` becoming `<100` or `<98`
- Validate method scope (no new exports without approval)
- Integration with CI/CD pipelines

**Usage:**
```bash
# Extract current rules
bizlint-go extract --pkg=internal/services --output=current-rules.json

# Compare against baseline  
bizlint-go diff --baseline=logic_maps/baseline.json --current=current-rules.json

# Full CI validation
bizlint-go ci --baseline=logic_maps/baseline.json --pkg=internal/services

# Scope validation
bizlint-go scope --pkg=internal/services --allowlist=migrations/scope-allowlist.txt
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
├── phase-00-tool-validation/
│   ├── README.md
│   ├── tracker.yaml
│   └── test-results.md
├── phase-01-baseline/
│   ├── README.md
│   ├── tracker.yaml
│   └── baseline-extraction.md
├── phase-02-add-endpoint/
│   ├── README.md
│   ├── tracker.yaml
│   └── patch.diff
├── phase-03-handlers/
├── phase-04a-product-service/
├── phase-04b-cart-service/
├── phase-04c-order-service/
├── phase-05-cleanup/
├── phase-06-validation/
└── scope-allowlist.txt

logic_maps/
└── baseline.json

docs/
└── business-rules.md

tests/
└── tool-validation/
    ├── sample-service.go
    └── expected-rules.json
```

## Success Criteria
- [ ] Tool validation passes (Phase 0)
- [ ] All phases complete with passing tests
- [ ] Zero business logic drift detected by `bizlint-go`
- [ ] No new functionality added beyond signature changes
- [ ] Full CI/CD pipeline remains green
- [ ] Documentation updated and accurate
- [ ] Symbol extraction validated with `parser-go`

## Rollback Strategy
Each phase is tagged for easy rollback:
- Phase issues: `git checkout <previous-phase-tag>`
- Complete rollback: `git checkout v0-pre-sig`
- Tool validation issues: `git checkout migration-sept-15`

## Risk Mitigation
- Phase 0 validates tools work with our codebase before migration
- Automated business logic validation prevents drift using `bizlint-go`
- Phased approach allows incremental validation
- Build tags maintain compatibility during transition
- Comprehensive test coverage at each phase
- Clear rollback points via git tags
- External tools (`parser-go`, `bizlint-go`) provide independent validation
