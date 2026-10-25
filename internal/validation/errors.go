// Package validation describes invalid input without depending on HTTP or a
// particular feature. Field names match the feature's input contract.
package validation

type Fields map[string][]string

type Error struct{ Fields Fields }

func (e *Error) Error() string { return "validation failed" }
