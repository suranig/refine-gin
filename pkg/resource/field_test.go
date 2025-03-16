package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringValidatorValidate(t *testing.T) {
	// Test with MinLength
	validator := StringValidator{MinLength: 5}
	err := validator.Validate("test")
	assert.Error(t, err, "Should return error for string shorter than minimum length")

	err = validator.Validate("testtest")
	assert.NoError(t, err, "Should not return error for string with sufficient length")

	// Test with MaxLength
	validator = StringValidator{MaxLength: 5}
	err = validator.Validate("testtest")
	assert.Error(t, err, "Should return error for string longer than maximum length")

	err = validator.Validate("test")
	assert.NoError(t, err, "Should not return error for string with acceptable length")

	// Test with both MinLength and MaxLength
	validator = StringValidator{MinLength: 3, MaxLength: 5}
	err = validator.Validate("te")
	assert.Error(t, err, "Should return error for string shorter than minimum length")

	err = validator.Validate("testtest")
	assert.Error(t, err, "Should return error for string longer than maximum length")

	err = validator.Validate("test")
	assert.NoError(t, err, "Should not return error for string with length within range")

	// Test with Pattern
	validator = StringValidator{Pattern: "^[a-z]+$"}
	err = validator.Validate("test")
	assert.NoError(t, err, "Should not return error for string matching pattern")

	err = validator.Validate("Test123")
	assert.Error(t, err, "Should return error for string not matching pattern")

	// Test with invalid pattern
	validator = StringValidator{Pattern: "["}
	err = validator.Validate("test")
	assert.Error(t, err, "Should return error for invalid pattern")
	assert.Contains(t, err.Error(), "invalid pattern")

	// Test with invalid type
	err = validator.Validate(123)
	assert.Error(t, err, "Should return error for non-string value")
	assert.Contains(t, err.Error(), "value must be a string")

	// Test with no validation rules
	validator = StringValidator{}
	err = validator.Validate("test")
	assert.NoError(t, err, "Should not return error when no validation rules are set")
}

func TestNumberValidatorValidate(t *testing.T) {
	// Test with Min
	validator := NumberValidator{Min: 5}
	err := validator.Validate(4)
	assert.Error(t, err, "Should return error for number less than minimum")

	err = validator.Validate(6)
	assert.NoError(t, err, "Should not return error for number greater than minimum")

	// Test with Max
	validator = NumberValidator{Max: 5}
	err = validator.Validate(6)
	assert.Error(t, err, "Should return error for number greater than maximum")

	err = validator.Validate(4)
	assert.NoError(t, err, "Should not return error for number less than maximum")

	// Test with both Min and Max
	validator = NumberValidator{Min: 3, Max: 5}
	err = validator.Validate(2)
	assert.Error(t, err, "Should return error for number less than minimum")

	err = validator.Validate(6)
	assert.Error(t, err, "Should return error for number greater than maximum")

	err = validator.Validate(4)
	assert.NoError(t, err, "Should not return error for number within range")

	// Test with float
	validator = NumberValidator{Min: 3.5, Max: 5.5}
	err = validator.Validate(3.0)
	assert.Error(t, err, "Should return error for number less than minimum")

	err = validator.Validate(6.0)
	assert.Error(t, err, "Should return error for number greater than maximum")

	err = validator.Validate(4.5)
	assert.NoError(t, err, "Should not return error for number within range")

	// Test with various numeric types
	validator = NumberValidator{Min: 3, Max: 10}

	// Test with float32
	err = validator.Validate(float32(4.5))
	assert.NoError(t, err, "Should not return error for float32 within range")

	// Test with int8
	err = validator.Validate(int8(5))
	assert.NoError(t, err, "Should not return error for int8 within range")

	// Test with int16
	err = validator.Validate(int16(6))
	assert.NoError(t, err, "Should not return error for int16 within range")

	// Test with int32
	err = validator.Validate(int32(7))
	assert.NoError(t, err, "Should not return error for int32 within range")

	// Test with int64
	err = validator.Validate(int64(8))
	assert.NoError(t, err, "Should not return error for int64 within range")

	// Test with uint
	err = validator.Validate(uint(9))
	assert.NoError(t, err, "Should not return error for uint within range")

	// Test with uint8
	err = validator.Validate(uint8(4))
	assert.NoError(t, err, "Should not return error for uint8 within range")

	// Test with uint16
	err = validator.Validate(uint16(5))
	assert.NoError(t, err, "Should not return error for uint16 within range")

	// Test with uint32
	err = validator.Validate(uint32(6))
	assert.NoError(t, err, "Should not return error for uint32 within range")

	// Test with uint64
	err = validator.Validate(uint64(7))
	assert.NoError(t, err, "Should not return error for uint64 within range")

	// Test with string that can be converted to number
	err = validator.Validate("4.5")
	assert.NoError(t, err, "Should not return error for string that can be converted to number")

	// Test with invalid string
	err = validator.Validate("abc")
	assert.Error(t, err, "Should return error for value that cannot be converted to number")

	// Test with invalid type
	err = validator.Validate(struct{}{})
	assert.Error(t, err, "Should return error for value that cannot be converted to number")
}
