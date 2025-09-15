# Business Rules Documentation

**Generated from baseline extraction on:** September 15, 2025  
**Version:** v0-pre-sig  
**Source:** Extracted from service layer using bizlint-go tool

## Overview

This document captures the critical business rules across all services to ensure migration integrity. Any changes to these rules during signature migration will be flagged as drift.

## Product Service Rules

### Product Validation Rules
- **Price Range:** $0.01 - $10,000
- **Stock Range:** 0 - 10,000 units
- **Title Length:** 1 - 200 characters (required)
- **Author:** Required field
- **Description Length:** Max 2,000 characters

### Business Logic
- **Delete Protection:** Products cannot be deleted if present in any cart
- **Validation Methods:** CreateProduct and UpdateProduct both use `validateProductInput()`

### Critical Rules (High Risk Score: 10)
```go
// Price validation
if price < 0.01 || price > 10000 {
    return errors.New("price out of bounds")
}

// Stock validation  
if stock < 0 || stock > 10000 {
    return errors.New("invalid stock")
}

// Cart dependency check
if s.store.IsProductInAnyCart(req.ID) {
    return errors.New("product is present in carts")
}
```

## Cart Service Rules

### Cart Limits (Constants from rules.go)
- **Max Distinct Items:** 3 items
- **Max Quantity Per Item:** 5 units
- **Max Total Items:** 10 units total
- **Risk Limit Total:** $5,000
- **Low Stock Threshold:** 3 units (special handling)

### AddToCart Business Logic (Complexity: 29, Risk: 10)
1. **Product Availability:** Reject discontinued products
2. **Distinct Item Limit:** Max 3 different products per cart
3. **Quantity Limits:** Max 5 units per line item
4. **Stock Validation:** Ensure sufficient inventory
5. **Low Stock Rules:** Items with <3 stock limited to 1 per order
6. **Total Item Limit:** Max 10 total units across all items
7. **Price Consistency:** Handle price changes and zero prices
8. **Risk Management:** Cart total cannot exceed $5,000
9. **Inventory Reservation:** Real-time stock checks

### Critical Validation Rules
```go
// Low stock item restriction
if p.Stock < 3 && currentQty+req.Quantity > 1 {
    return nil, errors.New("low-stock item limited to 1 per order")
}

// Quantity per line item
if currentQty+req.Quantity > MaxQuantityPerLineItem {
    return nil, errors.New("quantity exceeds per-item limit")
}

// Total cart risk limit
if total > CartRiskLimitTotal {
    return nil, errors.New("cart total exceeds limit")
}
```

## Order Service Rules

### Order Constraints (Constants from rules.go)
- **Minimum Order:** $5.00
- **High Value Review:** $3,000+ orders flagged
- **Daily Spend Cap:** $10,000 per user
- **Duplicate Window:** 5 seconds between identical orders

### PlaceOrder Business Logic (Complexity: 35, Risk: 10)
1. **Duplicate Prevention:** Block identical orders within 5-second window
2. **Cart Validation:** Non-empty cart required
3. **Special Item Rules:** Special items must be purchased alone, quantity = 1
4. **Price Verification:** Detect price changes since cart creation
5. **Stock Validation:** Final inventory check before order creation
6. **Minimum Order:** Must meet $5.00 threshold
7. **Daily Limits:** User cannot exceed $10,000/day spending
8. **High Value Review:** Orders >$3,000 flagged for review
9. **Inventory Reservation:** Stock decremented atomically
10. **Cart Clearing:** User cart cleared on successful order

### Critical Business Rules
```go
// Duplicate order prevention
if time.Since(last.CreatedAt) <= time.Duration(DuplicateOrderWindowSec)*time.Second {
    return nil, errors.New("duplicate order detected")
}

// Special item restrictions
if p.IsSpecial && it.Quantity != 1 {
    return nil, errors.New("special items must have quantity 1")
}

// Daily spending cap
if todayTotal+total > DailyUserSpendCap {
    return nil, errors.New("daily spend limit reached")
}

// Minimum order amount
if total < MinOrderAmount {
    return nil, errors.New("order total below minimum")
}
```

## Migration Validation Strategy

### Baseline Comparison
The extracted baseline contains:
- **90 total business logic patterns** across all services
- **9 service methods** with their exact signatures
- **Critical rule thresholds and constants**
- **Complexity and risk scores** for drift detection

### Validation During Migration
1. **bizlint-go ci** will compare post-migration rules against this baseline
2. **Any rule changes** will be flagged as drift and block migration
3. **Signature changes only** are permitted - business logic must remain identical
4. **New exported methods** require explicit approval in scope-allowlist.txt

### Zero Drift Requirement
The migration **MUST NOT** change any business rule, validation threshold, or error condition. Only method signatures may change from:
- `func (s *Service) Method(ctx context.Context, req *dto.Request) (*dto.Response, error)`

To:
- `func (s *Service) Method(ctx context.Context, req *endpoint.HTTPRequest[*dto.Request]) (*endpoint.HTTPResponse[*dto.Response], error)`

Business logic within method bodies must remain byte-for-byte identical.
