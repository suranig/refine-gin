# Owner Resources Implementation Tasks

This document outlines the tasks needed to implement owner-based resources in the refine-gin framework. These tasks should be performed carefully to maintain compatibility with existing code.

## 1. Middleware Implementation

- [] Create `OwnerContextKey` constant for storing owner ID in context
- [] Implement `ExtractOwnerIDFunc` type for owner ID extraction functions
- [] Create `OwnerContext` middleware to extract and store owner IDs
- [] Implement extraction strategies:
  - [] `ExtractOwnerIDFromJWT` - Extract from JWT claims
  - [] `ExtractOwnerIDFromHeader` - Extract from HTTP headers
  - [] `ExtractOwnerIDFromQuery` - Extract from query parameters
  - [] `ExtractOwnerIDFromCookie` - Extract from cookies
  - [] `CombineExtractors` - Chain multiple extractors

## 2. Resource Definition

- [ ] Define `OwnerResource` interface extending `Resource`
- [ ] Create `OwnerConfig` struct for ownership configuration
- [ ] Implement `DefaultOwnerResource` wrapper for existing resources
- [ ] Add `NewOwnerResource` factory function

## 3. Repository Implementation

- [ ] Define error constants (`ErrOwnerMismatch`, `ErrOwnerIDNotFound`)
- [ ] Create `OwnerGenericRepository` extending standard repository
- [ ] Implement `NewOwnerRepository` factory function
- [ ] Add private helper methods:
  - [ ] `extractOwnerID` - Get owner ID from context
  - [ ] `applyOwnerFilter` - Add owner filter to queries
  - [ ] `verifyOwnership` - Check if user owns a record
  - [ ] `setOwnership` - Set owner field on new records
- [ ] Override CRUD operations with owner checks:
  - [ ] `List` - Filter by owner
  - [ ] `Count` - Filter by owner
  - [ ] `Get` - Verify ownership
  - [ ] `Create` - Set ownership
  - [ ] `Update` - Verify ownership
  - [ ] `Delete` - Verify ownership
  - [ ] `CreateMany` - Set ownership for all
  - [ ] `UpdateMany` - Verify ownership for all
  - [ ] `DeleteMany` - Verify ownership for all

## 4. Handler Implementation

- [ ] Create `OwnerHandlerOptions` struct
- [ ] Implement `RegisterOwnerResource` function
- [ ] Create handler generators for owner resources:
  - [ ] `GenerateOwnerListHandler`
  - [ ] `GenerateOwnerCountHandler`
  - [ ] `GenerateOwnerCreateHandler`
  - [ ] `GenerateOwnerGetHandler`
  - [ ] `GenerateOwnerUpdateHandler`
  - [ ] `GenerateOwnerDeleteHandler`
  - [ ] `GenerateOwnerCreateManyHandler`
  - [ ] `GenerateOwnerUpdateManyHandler`
  - [ ] `GenerateOwnerDeleteManyHandler`

## 5. Tests

- [ ] Write middleware tests
- [ ] Write owner resource tests
- [ ] Write owner repository tests
- [ ] Write handler tests
- [ ] Create example app

## Implementation Notes

1. Always check for compatibility with existing code
2. Don't create duplicate definitions
3. Follow existing naming conventions
4. Add proper documentation
5. Make sure all tests pass 