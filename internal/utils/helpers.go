package utils

import (
	"fmt"
	"strings"
)

// GetContent extracts text content from LLM response
func GetContent(content interface{}) string {
	switch v := content.(type) {
	case string:
		return v
	case []interface{}:
		var parts []string
		for _, part := range v {
			if m, ok := part.(map[string]interface{}); ok {
				if text, ok := m["text"].(string); ok {
					parts = append(parts, text)
				}
			}
		}
		return strings.Join(parts, "\n")
	default:
		return fmt.Sprintf("%v", v)
	}
}

// FloatPtr returns a pointer to the given float64
func FloatPtr(f float64) *float64 {
	return &f
}

// IntPtr returns a pointer to the given int
func IntPtr(i int) *int {
	return &i
}

