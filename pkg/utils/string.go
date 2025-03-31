package utils

import (
	"strings"

	"github.com/jinzhu/inflection"
)

// SplitTagParts splits a struct tag value into its parts.
// For example, "json:\"name,omitempty\"" would return []string{"name", "omitempty"}
func SplitTagParts(tag string) []string {
	if tag == "" {
		return nil
	}

	if strings.Contains(tag, ",") {
		parts := strings.Split(tag, ",")
		for i, part := range parts {
			parts[i] = strings.TrimSpace(part)
		}
		return parts
	}

	return []string{tag}
}

// Pluralize returns the plural form of a word using the inflection library
func Pluralize(word string) string {
	return inflection.Plural(word)
}
