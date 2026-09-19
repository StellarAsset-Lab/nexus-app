// Package network provides network configuration and Stellar RPC health
// checks shared by the Go services.
package network

import (
	"fmt"
	"os"
)

type Name string

const (
	Testnet Name = "testnet"
	Mainnet Name = "mainnet"
)

const (
	testnetPassphrase = "Test SDF Network ; September 2015"
	mainnetPassphrase = "Public Global Stellar Network ; September 2015"
)

type Config struct {
	Name       Name
	Passphrase string
	RPCURL     string
	HorizonURL string
}

// LoadConfig reads network configuration from the environment. This
// repository keeps a single .env shared by the Next.js app and the Go
// services, so the Go side reads the same NEXT_PUBLIC_* variables the
// frontend uses — the prefix only controls Next.js browser-bundle inlining,
// it does not make the variable Next-exclusive.
//
// As in the TypeScript SDK, the configured passphrase must match the known
// passphrase for the configured network name; a mismatch fails loudly rather
// than silently letting the service run against an inconsistent config.
func LoadConfig() (Config, error) {
	name := Name(os.Getenv("NEXT_PUBLIC_NETWORK"))
	passphrase := os.Getenv("NEXT_PUBLIC_NETWORK_PASSPHRASE")
	rpcURL := os.Getenv("NEXT_PUBLIC_SOROBAN_RPC_URL")
	horizonURL := os.Getenv("NEXT_PUBLIC_HORIZON_URL")

	var expectedPassphrase string
	switch name {
	case Testnet:
		expectedPassphrase = testnetPassphrase
	case Mainnet:
		expectedPassphrase = mainnetPassphrase
	default:
		return Config{}, fmt.Errorf(`NEXT_PUBLIC_NETWORK must be "testnet" or "mainnet", got %q`, name)
	}

	if passphrase != expectedPassphrase {
		return Config{}, fmt.Errorf(
			"NEXT_PUBLIC_NETWORK_PASSPHRASE does not match the passphrase for NEXT_PUBLIC_NETWORK=%q: expected %q, got %q",
			name, expectedPassphrase, passphrase,
		)
	}

	if rpcURL == "" {
		return Config{}, fmt.Errorf("NEXT_PUBLIC_SOROBAN_RPC_URL is not set")
	}
	if horizonURL == "" {
		return Config{}, fmt.Errorf("NEXT_PUBLIC_HORIZON_URL is not set")
	}

	return Config{Name: name, Passphrase: passphrase, RPCURL: rpcURL, HorizonURL: horizonURL}, nil
}
