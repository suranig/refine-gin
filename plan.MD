# Test Coverage Improvement Plan

## Goal
Increase the total test coverage from the current 87.0% to at least 94%.

## Current Coverage Status
- Total: 87.6% ⬆️ (was 87.0%)
- auth: 89.4%
- dto: 98.7% ✅
- handler: 81.8% ⚠️
- middleware: 89.9%
- naming: 97.1% ✅
- query: 94.7% ✅
- repository: 82.1% ⚠️
- resource: 90.4% ✅
- swagger: 97.0% ✅
- utils: 88.7% ⚠️

## Focus Areas (Prioritized)
We need to focus on packages with the lowest coverage first:

### 1. Handler Package (81.8%)
- [ ] Improve coverage for custom.go (especially AttachAction at 52.0%)
- [ ] Improve coverage for owner_handlers.go (especially GenerateOwnerCreateHandler at 64.7%)
- [ ] Improve coverage for update.go (especially GenerateCustomUpdateHandler at 66.7%)
- [ ] Improve coverage for form_metadata.go (RegisterResourceFormEndpoints at 54.1%)

### 2. Repository Package (82.1%)
- [ ] Improve coverage for owner_repository.go (especially Update at 63.8% and Create at 66.7%)
- [ ] Improve coverage for generic_repository.go (especially Update at 74.2%)

### 3. Utils Package (88.7%)
- [ ] Improve coverage for reflection.go (especially SetID at 65.4%)
- [ ] Improve coverage for type_mapping.go (GetTypeMapping at 76.9%)

### 4. Auth Package (89.4%)
- [ ] Improve coverage for middleware.go (CanAccessRecord at 76.9%)
- [ ] Improve coverage for jwt.go (JWTMiddleware at 80.8%)

### 5. Middleware Package (89.9%)
- [ ] Improve coverage for NamingConventionMiddleware (75.9%)
- [ ] Improve coverage for owner.go (GetOwnerID at 76.2%)

### 6. Resource Package (90.4%) ✅
- [x] Add tests for utils.go functions with 0% coverage (IsSlice, SetFieldValue)
- [x] Improve coverage for GetFieldValue (72.7% -> 100%)
- [x] Improve coverage for convertToNumber (80.0% -> 100%)
- [x] Add tests for relation_validator.go (0% -> 76.6%)

## Progress Tracking

### Resource Package Improvements
- [x] Identified functions with high cyclomatic complexity
- [x] Added tests for helper functions in utils.go
- [x] Fixed test failures in processArrayAccess
- [x] Added test for validateProperty
- [x] Added test for convertToNumber
- [x] Added test for detectTextComponent
- [x] Added tests for IsSlice (0% -> 100%)
- [x] Added tests for SetFieldValue (0% -> 89.5%)
- [x] Improved tests for GetFieldValue (72.7% -> 100%)
- [x] Improved tests for convertToNumber (80.0% -> 100%)
- [x] Added basic tests for relation_validator.go functions
- [x] Improved overall resource package coverage from 88.4% to 90.4%
- [x] Improved total project coverage from 87.0% to 87.6%

### Next Steps
- [x] Complete Resource Package improvements
- [ ] Move to Handler Package improvements
- [ ] Move to Repository Package improvements
- [ ] Move to Utils Package improvements
- [ ] Move to Auth Package improvements
- [ ] Move to Middleware Package improvements
- [ ] Final verification of total coverage

## Commit Checklist
- [x] Initial project analysis and test coverage assessment
- [x] Added tests for resource/utils.go helper functions
- [x] Added tests for resource/utils.go previously untested functions
- [x] Improved tests for convertToNumber function to 100% coverage
- [x] Added basic tests for relation_validator.go
- [x] Completed resource package coverage improvements (90.4%)
- [ ] Handler package coverage improvements
- [ ] Repository package coverage improvements
- [ ] Utils package coverage improvements
- [ ] Auth package coverage improvements
- [ ] Middleware package coverage improvements
- [ ] Final coverage verification and cleanup 