package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseValidationTag(t *testing.T) {
	tests := []struct {
		tag  string
		want *JsonValidation
	}{
		{
			tag:  "required",
			want: &JsonValidation{Required: true},
		},
		{
			tag:  "min=1,max=10",
			want: &JsonValidation{Min: 1, Max: 10},
		},
		{
			tag:  "minlength=5,maxlength=20",
			want: &JsonValidation{MinLength: 5, MaxLength: 20},
		},
		{
			tag:  "required,min=0,max=100,minlength=2,maxlength=10",
			want: &JsonValidation{Required: true, Min: 0, Max: 100, MinLength: 2, MaxLength: 10},
		},
		{
			tag:  "min=foo,max=bar",
			want: &JsonValidation{},
		},
	}

	for _, tt := range tests {
		got := parseValidationTag(tt.tag)
		assert.Equal(t, tt.want, got, "tag: %s", tt.tag)
	}
}
