// Package model contains transport-independent records and utility helpers.
package model

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"regexp"
	"strings"

	"ptmuda.tech/backend/internal/apperror"
)

type Record map[string]any

func Text(r Record, k string) string {
	s, _ := r[k].(string)
	return s
}

func Number(r Record, k string) int {
	switch v := r[k].(type) {
	case json.Number:
		n, _ := v.Int64()
		return int(n)
	case int64:
		return int(v)
	case int:
		return v
	case float64:
		return int(v)
	default:
		return 0
	}
}

func Float(r Record, k string) float64 {
	switch v := r[k].(type) {
	case json.Number:
		f, _ := v.Float64()
		return f
	case float64:
		return v
	case string:
		var f float64
		fmt.Sscanf(v, "%f", &f)
		return f
	default:
		return 0.0
	}
}

func Flag(r Record, k string) bool {
	v, _ := r[k].(bool)
	return v
}

func Cents(v any) (int64, error) {
	r, ok := new(big.Rat).SetString(fmt.Sprint(v))
	if !ok {
		return 0, apperror.ErrInvalid
	}
	r.Mul(r, big.NewRat(100, 1))
	if !r.IsInt() || !r.Num().IsInt64() {
		return 0, apperror.ErrInvalid
	}
	return r.Num().Int64(), nil
}

func Amount(c int64) json.Number {
	return json.Number(fmt.Sprintf("%d.%02d", c/100, c%100))
}

var uuidPattern = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)

func UUID(s string) bool {
	return uuidPattern.MatchString(s)
}

func NewID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 15) | 64
	b[8] = (b[8] & 63) | 128
	s := hex.EncodeToString(b[:])
	return s[:8] + "-" + s[8:12] + "-" + s[12:16] + "-" + s[16:20] + "-" + s[20:], nil
}

func ValidateFields(r Record, allowed string) error {
	fields := strings.Fields(allowed)
	for k := range r {
		ok := false
		for _, f := range fields {
			if k == f {
				ok = true
			}
		}
		if !ok {
			return apperror.Invalid("Unknown or protected field: " + k)
		}
	}
	return nil
}

func TextNumber(r Record, k string) string {
	if n, ok := r[k].(json.Number); ok {
		return n.String()
	}
	return fmt.Sprint(r[k])
}
