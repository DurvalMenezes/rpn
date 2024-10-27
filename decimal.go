// This file is part of rpn, a simple and useful CLI RPN calculator.
// For further information, check https://github.com/marcopaganini/rpn
//
// (C) Sep/2024 by Marco Paganini <paganini AT paganini DOT net>
package main

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/ericlagergren/decimal"
)

// This matches the precision in decimal.Context128
const precision = 34

// big returns a new *decimal.Big
func big() *decimal.Big {
	return decimal.WithPrecision(precision)
}

// bigUint returns a new *decimal.Big from an uint64.
func bigUint(n uint64) *decimal.Big {
	return decimal.WithPrecision(precision).SetUint64(n)
}

// bigFloat returns a new *decimal.Big from a string
// Using a float64 here will introduce rounding errors.
func bigFloat(s string) *decimal.Big {
	r, _ := decimal.WithPrecision(precision).SetString(s)
	return r
}

// commafWithDigits comes straight from humanize, but modified to call
// strconv.Formatfloat with 0 as the precision. This will print the entire
// number, and not truncate to the exact precision.
func commafWithDigits(v *decimal.Big, decimals int) string {
	buf := &bytes.Buffer{}
	if v.Sign() < 0 {
		buf.Write([]byte{'-'})
		// Make v positive
		v.SetSignbit(false)
	}

	comma := []byte{','}

	parts := strings.Split(fmt.Sprintf("%f", v), ".")

	pos := 0
	if len(parts[0])%3 != 0 {
		pos += len(parts[0]) % 3
		buf.WriteString(parts[0][:pos])
		buf.Write(comma)
	}
	for ; pos < len(parts[0]); pos += 3 {
		buf.WriteString(parts[0][pos : pos+3])
		buf.Write(comma)
	}
	buf.Truncate(buf.Len() - 1)

	if len(parts) > 1 {
		buf.Write([]byte{'.'})
		buf.WriteString(parts[1])
	}
	return stripTrailingDigits(buf.String(), decimals)
}

func stripTrailingDigits(s string, digits int) string {
	if i := strings.Index(s, "."); i >= 0 {
		if digits <= 0 {
			return s[:i]
		}
		i++
		if i+digits >= len(s) {
			return s
		}
		return s[:i+digits]
	}
	return s
}

// formatNumber formats the number using base. For bases different than 10,
// non-integer floating numbers are truncated.
func formatNumber(ctx decimal.Context, n *decimal.Big, base int) string {
	// Indicate possible truncation
	suffix := ""
	// clean = double as ascii, without non-significant decimal zeroes.
	clean := fmt.Sprintf("%f", n)

	buf := &bytes.Buffer{}
	var n64 uint64

	if base != 10 {
		// For negative numbers, prefix them with a minus sign and
		// force them to be positive.
		if n.Signbit() {
			buf.Write([]byte{'-'})
			n.SetSignbit(false)
		}
		// Truncate floating point numbers to their integer representation.
		if !n.IsInt() {
			suffix = fmt.Sprintf(" (truncated from %s)", clean)
			ctx.Floor(n, n)
		}
		// Non-base 10 uses uint64s.
		var ok bool
		n64, ok = n.Uint64()
		if !ok {
			return "Invalid number: non decimal base only supports uint64 numbers."
		}
	}

	switch {
	case base == 2:
		buf.WriteString(fmt.Sprintf("0b%b%s", n64, suffix))
	case base == 8:
		buf.WriteString(fmt.Sprintf("0%o%s", n64, suffix))
	case base == 16:
		buf.WriteString(fmt.Sprintf("0x%x%s", n64, suffix))
	default:
		h := commafWithDigits(n, 16) // FIXME find out how to deal with precision properly.
		// Only print humanized format when it differs from original value.
		if h != clean {
			suffix = " (" + h + ")"
		}
		buf.WriteString(clean + suffix)
	}

	return buf.String()
}