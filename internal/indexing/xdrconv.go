package indexing

import (
	"fmt"
	"math/big"

	"github.com/stellar/go-stellar-sdk/xdr"

	"github.com/StellarAsset-Lab/nexus-app/internal/model"
)

// scValToAddress decodes an Address-typed ScVal to its strkey string, using
// the SDK's own XDR/strkey conversion — never a hand-rolled parser.
func scValToAddress(v xdr.ScVal) (string, error) {
	addr, ok := v.GetAddress()
	if !ok {
		return "", fmt.Errorf("expected an Address ScVal, got %s", v.Type)
	}
	s, err := addr.String()
	if err != nil {
		return "", fmt.Errorf("encode address: %w", err)
	}
	return s, nil
}

func scValToU32(v xdr.ScVal) (uint32, error) {
	u, ok := v.GetU32()
	if !ok {
		return 0, fmt.Errorf("expected a U32 ScVal, got %s", v.Type)
	}
	return uint32(u), nil
}

func scValToU64(v xdr.ScVal) (uint64, error) {
	u, ok := v.GetU64()
	if !ok {
		return 0, fmt.Errorf("expected a U64 ScVal, got %s", v.Type)
	}
	return uint64(u), nil
}

func scValToBool(v xdr.ScVal) (bool, error) {
	b, ok := v.GetB()
	if !ok {
		return false, fmt.Errorf("expected a Bool ScVal, got %s", v.Type)
	}
	return b, nil
}

// scValToAmount decodes an i128-typed ScVal into an exact model.Amount,
// never through a float64.
func scValToAmount(v xdr.ScVal) (model.Amount, error) {
	parts, ok := v.GetI128()
	if !ok {
		return model.Amount{}, fmt.Errorf("expected an I128 ScVal, got %s", v.Type)
	}

	hi := big.NewInt(int64(parts.Hi))
	hi.Lsh(hi, 64)

	lo := new(big.Int).SetUint64(uint64(parts.Lo))

	result := new(big.Int).Add(hi, lo)
	return model.Amount{Int: *result}, nil
}

func scValToSymbol(v xdr.ScVal) (string, error) {
	sym, ok := v.GetSym()
	if !ok {
		return "", fmt.Errorf("expected a Symbol ScVal, got %s", v.Type)
	}
	return string(sym), nil
}
