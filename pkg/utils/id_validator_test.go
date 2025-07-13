package utils

import (
	"reflect"
	"testing"
)

func TestValidateIDFormat(t *testing.T) {
	tests := []struct {
		name    string
		id      interface{}
		wantErr bool
		errCode string
	}{
		// Valid cases
		{"valid string", "abc123", false, ""},
		{"valid int", 123, false, ""},
		{"valid int64", int64(456), false, ""},
		{"valid uint", uint(789), false, ""},
		{"valid float", 123.0, false, ""},
		{"valid string with hyphens", "abc-123", false, ""},
		{"valid string with underscores", "abc_123", false, ""},
		{"valid string with dots", "abc.123", false, ""},

		// Invalid cases
		{"nil ID", nil, true, ErrCodeEmptyID},
		{"empty string", "", true, ErrCodeEmptyID},
		{"whitespace string", "   ", true, ErrCodeEmptyID},
		{"zero int", 0, true, ErrCodeEmptyID},
		{"zero uint", uint(0), true, ErrCodeEmptyID},
		{"zero float", 0.0, true, ErrCodeEmptyID},
		{"negative int", -1, true, ErrCodeInvalidIDValue},
		{"negative float", -1.5, true, ErrCodeInvalidIDValue},
		{"invalid string chars", "abc@123", true, ErrCodeInvalidIDValue},
		{"invalid string chars 2", "abc#123", true, ErrCodeInvalidIDValue},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIDFormat(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateIDFormat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if code := GetIDValidationErrorCode(err); code != tt.errCode {
					t.Errorf("ValidateIDFormat() error code = %v, want %v", code, tt.errCode)
				}
			}
		})
	}
}

func TestValidateIDSlice(t *testing.T) {
	tests := []struct {
		name    string
		ids     []interface{}
		wantErr bool
		errCode string
	}{
		// Valid cases
		{"valid slice", []interface{}{"abc", "def", "ghi"}, false, ""},
		{"valid numeric slice", []interface{}{1, 2, 3}, false, ""},
		{"valid mixed slice", []interface{}{"abc", 123, "def"}, false, ""},

		// Invalid cases
		{"empty slice", []interface{}{}, true, ErrCodeEmptyID},
		{"nil slice", nil, true, ErrCodeEmptyID},
		{"contains empty string", []interface{}{"abc", "", "def"}, true, ErrCodeEmptyID},
		{"contains zero", []interface{}{1, 0, 3}, true, ErrCodeEmptyID},
		{"contains negative", []interface{}{1, -1, 3}, true, ErrCodeInvalidIDValue},
		{"duplicate IDs", []interface{}{"abc", "abc", "def"}, true, ErrCodeDuplicateID},
		{"duplicate numeric IDs", []interface{}{1, 1, 3}, true, ErrCodeDuplicateID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIDSlice(tt.ids)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateIDSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if code := GetIDValidationErrorCode(err); code != tt.errCode {
					t.Errorf("ValidateIDSlice() error code = %v, want %v", code, tt.errCode)
				}
			}
		})
	}
}

func TestParseAndValidateIDs(t *testing.T) {
	tests := []struct {
		name    string
		idData  interface{}
		want    []interface{}
		wantErr bool
		errCode string
	}{
		// Valid cases
		{
			name:    "interface slice",
			idData:  []interface{}{"abc", "def"},
			want:    []interface{}{"abc", "def"},
			wantErr: false,
		},
		{
			name:    "string slice",
			idData:  []string{"abc", "def"},
			want:    []interface{}{"abc", "def"},
			wantErr: false,
		},
		{
			name:    "int slice",
			idData:  []int{1, 2, 3},
			want:    []interface{}{1, 2, 3},
			wantErr: false,
		},
		{
			name:    "single string",
			idData:  "abc",
			want:    []interface{}{"abc"},
			wantErr: false,
		},
		{
			name:    "single int",
			idData:  123,
			want:    []interface{}{123},
			wantErr: false,
		},

		// Invalid cases
		{
			name:    "nil data",
			idData:  nil,
			wantErr: true,
			errCode: ErrCodeEmptyID,
		},
		{
			name:    "empty slice",
			idData:  []interface{}{},
			wantErr: true,
			errCode: ErrCodeEmptyID,
		},
		{
			name:    "contains invalid ID",
			idData:  []interface{}{"abc", "", "def"},
			wantErr: true,
			errCode: ErrCodeEmptyID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAndValidateIDs(tt.idData)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAndValidateIDs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if code := GetIDValidationErrorCode(err); code != tt.errCode {
					t.Errorf("ParseAndValidateIDs() error code = %v, want %v", code, tt.errCode)
				}
			}
			if !tt.wantErr {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ParseAndValidateIDs() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestConvertIDToString(t *testing.T) {
	tests := []struct {
		name    string
		id      interface{}
		want    string
		wantErr bool
	}{
		{"string ID", "abc123", "abc123", false},
		{"int ID", 123, "123", false},
		{"int64 ID", int64(456), "456", false},
		{"uint ID", uint(789), "789", false},
		{"float ID", 123.0, "123", false},
		{"float with decimal", 123.5, "123.5", false},
		{"zero int", 0, "", true},
		{"empty string", "", "", true},
		{"nil", nil, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertIDToString(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertIDToString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ConvertIDToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertStringToID(t *testing.T) {
	tests := []struct {
		name       string
		idStr      string
		targetType reflect.Type
		want       interface{}
		wantErr    bool
	}{
		{"string to string", "abc123", reflect.TypeOf(""), "abc123", false},
		{"string to int", "123", reflect.TypeOf(0), 123, false},
		{"string to int64", "456", reflect.TypeOf(int64(0)), int64(456), false},
		{"string to uint", "789", reflect.TypeOf(uint(0)), uint(789), false},
		{"string to float64", "123.5", reflect.TypeOf(0.0), 123.5, false},
		{"empty string", "", reflect.TypeOf(""), nil, true},
		{"invalid int", "abc", reflect.TypeOf(0), nil, true},
		{"invalid float", "abc", reflect.TypeOf(0.0), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertStringToID(tt.idStr, tt.targetType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertStringToID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertStringToID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsIDValidationError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"IDValidationError", IDValidationError{Code: "TEST", Message: "test"}, true},
		{"regular error", &testError{}, false},
		{"nil error", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsIDValidationError(tt.err); got != tt.want {
				t.Errorf("IsIDValidationError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetIDValidationErrorCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"IDValidationError", IDValidationError{Code: "TEST", Message: "test"}, "TEST"},
		{"regular error", &testError{}, ""},
		{"nil error", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetIDValidationErrorCode(tt.err); got != tt.want {
				t.Errorf("GetIDValidationErrorCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper type for testing
type testError struct{}

func (e *testError) Error() string {
	return "test error"
}