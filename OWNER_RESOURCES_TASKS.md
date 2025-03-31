# Owner Resources Implementation Tasks

This document outlines the tasks needed to implement owner-based resources in the refine-gin framework. These tasks should be performed carefully to maintain compatibility with existing code.

## 1. Middleware Implementation

- [x] Create `OwnerContextKey` constant for storing owner ID in context
- [x] Implement `ExtractOwnerIDFunc` type for owner ID extraction functions
- [x] Create `OwnerContext` middleware to extract and store owner IDs
- [x] Implement extraction strategies:
  - [x] `ExtractOwnerIDFromJWT` - Extract from JWT claims
  - [x] `ExtractOwnerIDFromHeader` - Extract from HTTP headers
  - [x] `ExtractOwnerIDFromQuery` - Extract from query parameters
  - [x] `ExtractOwnerIDFromCookie` - Extract from cookies
  - [x] `CombineExtractors` - Chain multiple extractors

## 2. Resource Definition

- [x] Define `OwnerResource` interface extending `Resource`
- [x] Create `OwnerConfig` struct for ownership configuration
- [x] Implement `DefaultOwnerResource` wrapper for existing resources
- [x] Add `NewOwnerResource` factory function

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

- [x] Create `OwnerHandlerOptions` struct
- [x] Implement `RegisterOwnerResource` function
- [x] Create handler generators for owner resources:
  - [x] `GenerateOwnerListHandler`
  - [x] `GenerateOwnerCountHandler`
  - [x] `GenerateOwnerCreateHandler`
  - [x] `GenerateOwnerGetHandler`
  - [x] `GenerateOwnerUpdateHandler`
  - [x] `GenerateOwnerDeleteHandler`
  - [x] `GenerateOwnerCreateManyHandler`
  - [x] `GenerateOwnerUpdateManyHandler`
  - [x] `GenerateOwnerDeleteManyHandler`

## 5. Tests

- [x] Write middleware tests
- [x] Write owner resource tests
- [x] Write owner repository tests
- [x] Write handler tests
- [x] Create example app

## Implementation Notes

1. Always check for compatibility with existing code
2. Don't create duplicate definitions
3. Follow existing naming conventions
4. Add proper documentation
5. Make sure all tests pass 