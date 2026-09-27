// Package query provides descriptor-owned query expressions and a parameterized
// PostgreSQL compiler. It intentionally accepts only validated identifiers and
// typed values; request code cannot supply SQL fragments.
package query

import (
	"fmt"
	"strings"
	"unicode"
)

// Table identifies a descriptor-owned database table.
type Table struct{ identifier }

// Column identifies a descriptor-owned database column.
type Column struct{ identifier }

// SortKey identifies a configured sortable field.
type SortKey struct{ value string }

// SearchKey identifies a configured searchable field.
type SearchKey struct{ value string }

// ProjectionKey identifies a configured projection.
type ProjectionKey struct{ value string }

type identifier struct{ parts []string }

// NewTable validates a static table identifier.
func NewTable(value string) (Table, error) {
	identifier, err := newIdentifier(value)
	if err != nil {
		return Table{}, err
	}
	return Table{identifier: identifier}, nil
}

// MustTable creates a static table identifier and panics for invalid metadata.
func MustTable(value string) Table {
	table, err := NewTable(value)
	if err != nil {
		panic(err)
	}
	return table
}

// NewColumn validates a static column identifier.
func NewColumn(value string) (Column, error) {
	identifier, err := newIdentifier(value)
	if err != nil {
		return Column{}, err
	}
	return Column{identifier: identifier}, nil
}

// MustColumn creates a static column identifier and panics for invalid metadata.
func MustColumn(value string) Column {
	column, err := NewColumn(value)
	if err != nil {
		panic(err)
	}
	return column
}

// NewSortKey validates a descriptor-owned sort key.
func NewSortKey(value string) (SortKey, error) {
	if !validKey(value) {
		return SortKey{}, newError(InvalidSortKey, "invalid sort key")
	}
	return SortKey{value: value}, nil
}

// MustSortKey creates a static sort key and panics for invalid metadata.
func MustSortKey(value string) SortKey {
	key, err := NewSortKey(value)
	if err != nil {
		panic(err)
	}
	return key
}

// NewSearchKey validates a descriptor-owned search key.
func NewSearchKey(value string) (SearchKey, error) {
	if !validKey(value) {
		return SearchKey{}, newError(InvalidSearchKey, "invalid search key")
	}
	return SearchKey{value: value}, nil
}

// MustSearchKey creates a static search key and panics for invalid metadata.
func MustSearchKey(value string) SearchKey {
	key, err := NewSearchKey(value)
	if err != nil {
		panic(err)
	}
	return key
}

// NewProjectionKey validates a descriptor-owned projection key.
func NewProjectionKey(value string) (ProjectionKey, error) {
	if !validKey(value) {
		return ProjectionKey{}, newError(InvalidProjectionKey, "invalid projection key")
	}
	return ProjectionKey{value: value}, nil
}

// MustProjectionKey creates a static projection key and panics for invalid metadata.
func MustProjectionKey(value string) ProjectionKey {
	key, err := NewProjectionKey(value)
	if err != nil {
		panic(err)
	}
	return key
}

// SQL returns the safely quoted static table identifier.
func (value Table) SQL() string { return value.identifier.sql() }

// SQL returns the safely quoted static column identifier.
func (value Column) SQL() string { return value.identifier.sql() }

func (value Table) sql() string  { return value.SQL() }
func (value Column) sql() string { return value.SQL() }

func newIdentifier(value string) (identifier, error) {
	parts := strings.Split(value, ".")
	if len(parts) == 0 {
		return identifier{}, newError(InvalidIdentifier, "identifier is empty")
	}
	for _, part := range parts {
		if !validIdentifierPart(part) {
			return identifier{}, newError(InvalidIdentifier, fmt.Sprintf("invalid identifier %q", value))
		}
	}
	return identifier{parts: parts}, nil
}

func (value identifier) sql() string {
	parts := make([]string, len(value.parts))
	for index, part := range value.parts {
		parts[index] = `"` + part + `"`
	}
	return strings.Join(parts, ".")
}

func validIdentifierPart(value string) bool {
	if value == "" {
		return false
	}
	for index, character := range value {
		if character == '_' || unicode.IsLetter(character) || (index > 0 && unicode.IsDigit(character)) {
			continue
		}
		return false
	}
	return true
}

func validKey(value string) bool {
	return validIdentifierPart(value)
}
