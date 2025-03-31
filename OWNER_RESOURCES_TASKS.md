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

- [x] Define error constants (`ErrOwnerMismatch`, `ErrOwnerIDNotFound`)
- [x] Create `OwnerGenericRepository` extending standard repository
- [x] Implement `NewOwnerRepository` factory function
- [x] Add private helper methods:
  - [x] `extractOwnerID` - Get owner ID from context
  - [x] `applyOwnerFilter` - Add owner filter to queries
  - [x] `verifyOwnership` - Check if user owns a record
  - [x] `setOwnership` - Set owner field on new records
- [x] Override CRUD operations with owner checks:
  - [x] `List` - Filter by owner
  - [x] `Count` - Filter by owner
  - [x] `Get` - Verify ownership
  - [x] `Create` - Set ownership
  - [x] `Update` - Verify ownership
  - [x] `Delete` - Verify ownership
  - [x] `CreateMany` - Set ownership for all
  - [x] `UpdateMany` - Verify ownership for all
  - [x] `DeleteMany` - Verify ownership for all

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
- [x] Write owner repository tests
- [ ] Write handler tests
- [ ] Create example app

## Implementation Notes

1. Always check for compatibility with existing code
2. Don't create duplicate definitions
3. Follow existing naming conventions
4. Add proper documentation
5. Make sure all tests pass 