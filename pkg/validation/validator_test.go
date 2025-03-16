package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestModel for validation tests
type TestModel struct {
	ID       string
	Name     string
	Email    string
	Age      int
	Password string
}

func TestValidationError(t *testing.T) {
	// Test single validation error
	err := ValidationError{
		Field:   "name",
		Message: "Name is required",
	}

	assert.Equal(t, "name: Name is required", err.Error())

	// Test multiple validation errors
	errors := ValidationErrors{
		{Field: "name", Message: "Name is required"},
		{Field: "email", Message: "Invalid email format"},
	}

	assert.Contains(t, errors.Error(), "name: Name is required")
	assert.Contains(t, errors.Error(), "email: Invalid email format")

	// Test empty validation errors
	emptyErrors := ValidationErrors{}
	assert.Equal(t, "", emptyErrors.Error())
}

func TestRequiredRule(t *testing.T) {
	rule := RequiredRule{Message: "Field is required"}

	// Test with nil value
	err := rule.Validate(nil)
	assert.NotNil(t, err)
	assert.Equal(t, "Field is required", err.Message)

	// Test with empty string
	err = rule.Validate("")
	assert.NotNil(t, err)
	assert.Equal(t, "Field is required", err.Message)

	// Test with non-empty string
	err = rule.Validate("test")
	assert.Nil(t, err)

	// Test with empty slice
	err = rule.Validate([]string{})
	assert.NotNil(t, err)
	assert.Equal(t, "Field is required", err.Message)

	// Test with non-empty slice
	err = rule.Validate([]string{"test"})
	assert.Nil(t, err)

	// Test with nil pointer
	var ptr *string
	err = rule.Validate(ptr)
	assert.NotNil(t, err)
	assert.Equal(t, "Field is required", err.Message)

	// Test with non-nil pointer
	str := "test"
	ptr = &str
	err = rule.Validate(ptr)
	assert.Nil(t, err)
}

func TestStringLengthRule(t *testing.T) {
	// Test with min length only
	rule := StringLengthRule{
		Min:     3,
		Message: "String must be at least 3 characters",
	}

	err := rule.Validate("ab")
	assert.NotNil(t, err)
	assert.Equal(t, "String must be at least 3 characters", err.Message)

	err = rule.Validate("abc")
	assert.Nil(t, err)

	// Test with max length only
	rule = StringLengthRule{
		Max:     5,
		Message: "String must be at most 5 characters",
	}

	err = rule.Validate("abcdef")
	assert.NotNil(t, err)
	assert.Equal(t, "String must be at most 5 characters", err.Message)

	err = rule.Validate("abcde")
	assert.Nil(t, err)

	// Test with both min and max length
	rule = StringLengthRule{
		Min:     3,
		Max:     5,
		Message: "String must be between 3 and 5 characters",
	}

	err = rule.Validate("ab")
	assert.NotNil(t, err)
	assert.Equal(t, "String must be between 3 and 5 characters", err.Message)

	err = rule.Validate("abcdef")
	assert.NotNil(t, err)
	assert.Equal(t, "String must be between 3 and 5 characters", err.Message)

	err = rule.Validate("abcd")
	assert.Nil(t, err)

	// Test with nil value
	err = rule.Validate(nil)
	assert.Nil(t, err)

	// Test with non-string value
	err = rule.Validate(123)
	assert.Nil(t, err)
}

func TestPatternRule(t *testing.T) {
	// Test with email pattern
	rule := PatternRule{
		Pattern: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
		Message: "Invalid email format",
	}

	err := rule.Validate("invalid-email")
	assert.NotNil(t, err)
	assert.Equal(t, "Invalid email format", err.Message)

	err = rule.Validate("test@example.com")
	assert.Nil(t, err)

	// Test with nil value
	err = rule.Validate(nil)
	assert.Nil(t, err)

	// Test with non-string value
	err = rule.Validate(123)
	assert.Nil(t, err)
}

func TestStructValidator(t *testing.T) {
	// Create a validator
	validator := NewStructValidator()

	// Add rules
	validator.Required("Name", "Name is required")
	validator.StringLength("Name", 3, 50, "Name must be between 3 and 50 characters")
	validator.Pattern("Email", `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, "Invalid email format")

	// Test with valid model
	model := TestModel{
		ID:    "1",
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}

	errors := validator.Validate(model)
	assert.Empty(t, errors)

	// Test with invalid model
	model = TestModel{
		ID:    "1",
		Name:  "Jo",
		Email: "invalid-email",
		Age:   30,
	}

	errors = validator.Validate(model)
	assert.Len(t, errors, 2)

	// Check error messages
	nameError := findErrorByField(errors, "Name")
	assert.NotNil(t, nameError)
	assert.Equal(t, "Name must be between 3 and 50 characters", nameError.Message)

	emailError := findErrorByField(errors, "Email")
	assert.NotNil(t, emailError)
	assert.Equal(t, "Invalid email format", emailError.Message)

	// Test with nil model
	errors = validator.Validate(nil)
	assert.Empty(t, errors)

	// Test with non-struct model
	errors = validator.Validate("not a struct")
	assert.Empty(t, errors)
}

// Helper function to find a validation error by field
func findErrorByField(errors ValidationErrors, field string) *ValidationError {
	for _, err := range errors {
		if err.Field == field {
			return &err
		}
	}
	return nil
}
