package main

import (
	"crypto/sha256"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func main() {
	// Configure bech32 prefix BEFORE generating addresses
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("vita", "vitapub")
	config.SetBech32PrefixForValidator("vitavaloper", "vitavaloperpub")
	config.SetBech32PrefixForConsensusNode("vitavalcons", "vitavalconspub")
	config.Seal()
	
	// Generate addresses with proper checksums
	testSeeds := []string{
		"testmerchant1",
		"testmerchant2",
		"testpayer1",
		"testpayer2",
		"testowner1",
		"testowner2",
		"testsender1",
		"testsender2",
		"testunauthorized",
		"testrecipient",
	}
	
	fmt.Println("// Valid vita bech32 addresses for testing:")
	for i, seed := range testSeeds {
		// Use hash to create deterministic 20-byte address
		hash := sha256.Sum256([]byte(seed))
		addr := sdk.AccAddress(hash[:20])
		fmt.Printf("const validAddress%d = \"%s\" // %s\n", i+1, addr.String(), seed)
	}
	
	fmt.Println("\n// Additional addresses for tier tests:")
	for i := 1; i <= 5; i++ {
		seed := fmt.Sprintf("testtier%d", i)
		hash := sha256.Sum256([]byte(seed))
		addr := sdk.AccAddress(hash[:20])
		fmt.Printf("const tierTestAddr%d = \"%s\"\n", i, addr.String())
	}
	
	fmt.Println("\n// Additional addresses for vault tests:")
	for i := 1; i <= 3; i++ {
		seed := fmt.Sprintf("testvault%d", i)
		hash := sha256.Sum256([]byte(seed))
		addr := sdk.AccAddress(hash[:20])
		fmt.Printf("const vaultTestAddr%d = \"%s\"\n", i, addr.String())
	}
}
