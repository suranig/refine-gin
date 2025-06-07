package utils

import "testing"

func TestSplitTagParts(t *testing.T) {
	tests := []struct {
		tag      string
		expected []string
	}{
		{"", nil},
		{"name", []string{"name"}},
		{"name,omitempty", []string{"name", "omitempty"}},
		{" name , omitempty ", []string{"name", "omitempty"}},
	}

	for _, tt := range tests {
		got := SplitTagParts(tt.tag)
		if len(got) != len(tt.expected) {
			t.Fatalf("SplitTagParts(%q) length=%d want %d", tt.tag, len(got), len(tt.expected))
		}
		for i, v := range got {
			if v != tt.expected[i] {
				t.Fatalf("SplitTagParts(%q)[%d]=%q want %q", tt.tag, i, v, tt.expected[i])
			}
		}
	}
}

func TestPluralize(t *testing.T) {
	if Pluralize("box") != "boxes" {
		t.Fatalf("unexpected plural for box")
	}
	if Pluralize("person") != "people" {
		t.Fatalf("unexpected plural for person")
	}
}
