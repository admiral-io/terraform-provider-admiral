package provider

import (
	"fmt"
	"strings"
)

// The server's filter DSL reads single-quoted strings as:
// "'" ( "\'" / !"'" . )* "'". The only escape is \' for a literal quote; a
// backslash before anything else is itself literal. So every user-supplied
// value must have its quotes escaped or it can break out of the literal, and
// a value ending in a backslash cannot be expressed at all, because its
// closing "\'" would read as an escaped quote. Mirrors admiral-cli/internal/filter.

// filterQuote renders v as a single-quoted DSL string literal.
func filterQuote(v string) (string, error) {
	if strings.HasSuffix(v, `\`) {
		return "", fmt.Errorf("value %q cannot be used in a filter: it ends with a backslash", v)
	}
	return "'" + strings.ReplaceAll(v, "'", `\'`) + "'", nil
}

// filterEq builds a single equality predicate, field['<field>'] = '<value>'.
func filterEq(field, value string) (string, error) {
	f, err := filterQuote(field)
	if err != nil {
		return "", err
	}
	v, err := filterQuote(value)
	if err != nil {
		return "", err
	}
	return "field[" + f + "] = " + v, nil
}
