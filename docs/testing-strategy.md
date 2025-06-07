# Testing Strategy for dis.quest

## Overview

We use a pragmatic testing approach focused on integration tests for the API → Database layer, using SQLite for fast, reliable testing without external dependencies.

## Testing Architecture

### **Standard Go Tests** ✅ (Chosen Approach)
- **Native Go tooling**: `go test` with `httptest` package
- **Fast execution**: In-memory SQLite for instant startup
- **Easy debugging**: Standard Go debugging works seamlessly
- **No external dependencies**: Perfect for CI/CD

### **SQLite for Testing** ✅ (Chosen Database)
- **Lightning fast**: In-memory database (`:memory:`)
- **SQL compatibility**: Same SQLC queries work identically
- **No containers**: No Docker or external setup needed
- **CI friendly**: Zero configuration required

## Test Structure

### Integration Tests
Test the full API → Middleware → Database flow:

```go
func TestTopicsAPI_CreateTopic_Integration(t *testing.T) {
    // 1. Create test database (in-memory SQLite)
    dbService := testutil.TestDatabase(t)
    
    // 2. Set up HTTP server with real routes
    mux := http.NewServeMux()
    RegisterRoutes(mux, "/", cfg, dbService)
    
    // 3. Make HTTP request
    req := httptest.NewRequest("POST", "/api/topics", body)
    w := httptest.NewRecorder()
    mux.ServeHTTP(w, req)
    
    // 4. Assert response
    assert.Equal(t, http.StatusCreated, w.Code)
    
    // 5. Verify database state
    topics := getTopicsFromDB(dbService)
    assert.Len(t, topics, 1)
}
```

### Test Utilities

**Test Database Setup** (`internal/testutil/`):
- `TestDatabase(t)` - Creates in-memory SQLite with schema
- `CreateTestSchema(db)` - Sets up tables and indexes
- Automatic cleanup with `t.Cleanup()`

**Helper Functions**:
- `createTestTopics(t, dbService, count)` - Bulk test data
- `CreateTestUser(t, dbService)` - User setup (when auth is complete)

## Running Tests

```bash
# Run all integration tests
go test ./server/app -v

# Run specific test
go test ./server/app -v -run TestTopicsAPI_CreateTopic

# Run with coverage
go test ./server/app -cover

# Run tests in parallel
go test ./server/app -parallel 4
```

## Test Categories

### 1. **API Integration Tests** (`*_integration_test.go`)
- **HTTP API → Database** end-to-end
- **Real middleware chain** (auth, validation, etc.)
- **Database state verification**
- **Error handling scenarios**

### 2. **Unit Tests** (`*_test.go`)
- **Individual functions** and methods
- **Validation logic** (input validation)
- **Utility functions** (JWT parsing, etc.)
- **Mock dependencies** where needed

### 3. **Database Tests** (`internal/db/*_test.go`)
- **SQLC query testing**
- **Transaction behavior**
- **Driver compatibility** (SQLite vs PostgreSQL)

## Example Test Scenarios

### API Tests
```go
// Test cases for topic creation
tests := []struct {
    name           string
    requestBody    map[string]interface{}
    expectedStatus int
    expectError    bool
}{
    {
        name: "Valid topic creation",
        requestBody: map[string]interface{}{
            "subject":         "Test Topic",
            "initial_message": "Test message",
            "category":        "testing",
        },
        expectedStatus: http.StatusCreated,
        expectError:    false,
    },
    {
        name: "Invalid - empty subject",
        requestBody: map[string]interface{}{
            "subject": "",
            "initial_message": "Test message",
        },
        expectedStatus: http.StatusBadRequest,
        expectError:    true,
    },
}
```

### Database State Verification
```go
// After API call, verify data persisted correctly
topics, err := dbService.Queries().ListTopics(ctx, db.ListTopicsParams{
    Limit: 10, Offset: 0,
})
require.NoError(t, err)
assert.Len(t, topics, 1)
assert.Equal(t, "Test Topic", topics[0].Subject)
```

## Benefits of This Approach

### ✅ **Speed**
- **In-memory SQLite**: ~1ms test execution
- **No containers**: Instant startup
- **Parallel execution**: Tests don't interfere

### ✅ **Reliability**
- **Deterministic**: Same SQL queries as production
- **Isolated**: Each test gets fresh database
- **No flaky network**: Everything in-process

### ✅ **Developer Experience**
- **Standard Go tooling**: Works with any IDE
- **Easy debugging**: Breakpoints work normally
- **No setup**: `go test` just works

### ✅ **CI/CD Friendly**
- **Zero dependencies**: No Docker/PostgreSQL needed
- **Fast feedback**: Tests complete in seconds
- **No environmental issues**: Consistent across machines

## When to Use Different Approaches

### **Integration Tests**: Use for...
- API endpoint testing
- Full request/response cycles
- Database persistence verification
- Authentication/authorization flows
- Error handling scenarios

### **Unit Tests**: Use for...
- Business logic functions
- Input validation
- Utility functions
- Complex algorithms

### **Manual Testing**: Use for...
- User experience validation
- Browser compatibility
- Visual regression testing
- Performance under load

## Future Considerations

### **PostgreSQL Testing** (Optional)
If you need PostgreSQL-specific testing:
```go
// Only when testing PostgreSQL-specific features
func TestPostgreSQLSpecificFeature(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping PostgreSQL test in short mode")
    }
    // Use testcontainers-go here
}
```

### **End-to-End Testing** (Future)
For full user journey testing:
- Browser automation with Playwright/Selenium
- Run against staging environment
- Include frontend JavaScript interactions

## Key Testing Principles

1. **Fast Feedback**: Tests should run in seconds, not minutes
2. **Isolated**: Each test starts with clean state
3. **Deterministic**: Same input always produces same output
4. **Readable**: Test intent should be clear from code
5. **Maintainable**: Easy to update when requirements change

This approach gives us confidence in our API behavior while maintaining fast development cycles and reliable CI/CD pipelines.