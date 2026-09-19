package model

import (
	"database/sql/driver"
	"fmt"
	"math/big"
)

// Amount is an exact, arbitrary-precision token amount matching Soroban's
// i128. It round-trips through PostgreSQL NUMERIC columns and JSON API
// responses as exact decimal strings — never through float64 or a JSON
// number, neither of which can represent an i128 exactly. See spec §10/§32.
type Amount struct {
	big.Int
}

func NewAmount(i int64) Amount {
	var a Amount
	a.Int.SetInt64(i)
	return a
}

// ParseAmount parses an exact decimal integer string (no fractional part,
// no scientific notation). Returns an error rather than silently rounding
// or truncating on invalid input.
func ParseAmount(s string) (Amount, error) {
	i, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return Amount{}, fmt.Errorf("invalid exact amount %q", s)
	}
	return Amount{Int: *i}, nil
}

func (a Amount) Value() (driver.Value, error) {
	return a.Int.String(), nil
}

func (a *Amount) Scan(src any) error {
	switch v := src.(type) {
	case string:
		return a.scanString(v)
	case []byte:
		return a.scanString(string(v))
	case nil:
		return fmt.Errorf("amount: cannot scan NULL")
	default:
		return fmt.Errorf("amount: unsupported scan type %T", src)
	}
}

func (a *Amount) scanString(s string) error {
	i, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return fmt.Errorf("amount: invalid decimal string %q", s)
	}
	a.Int = *i
	return nil
}

func (a Amount) MarshalJSON() ([]byte, error) {
	return []byte(`"` + a.Int.String() + `"`), nil
}

func (a *Amount) UnmarshalJSON(data []byte) error {
	s := string(data)
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return fmt.Errorf("amount: expected a JSON string, got %s", s)
	}
	return a.scanString(s[1 : len(s)-1])
}
