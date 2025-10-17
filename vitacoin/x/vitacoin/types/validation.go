package types

import (
	"fmt"
)

// Note: DefaultParams, Validate, and String methods for Params are defined in params.go

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:       DefaultParams(),
		MerchantList: []Merchant{},
		PaymentList:  []Payment{},
		VaultList:    []Vault{},
		PoolList:     []RewardPool{},
	}
}

// Validate performs basic genesis state validation
func (gs GenesisState) Validate() error {
	// Validate params
	if err := gs.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	
	// Validate merchants
	merchantAddresses := make(map[string]bool)
	for i, merchant := range gs.MerchantList {
		if err := merchant.Validate(); err != nil {
			return fmt.Errorf("invalid merchant at index %d: %w", i, err)
		}
		
		// Check for duplicate merchant addresses
		if merchantAddresses[merchant.Address] {
			return fmt.Errorf("duplicate merchant address: %s", merchant.Address)
		}
		merchantAddresses[merchant.Address] = true
	}
	
	// Validate payments
	paymentIDs := make(map[string]bool)
	for i, payment := range gs.PaymentList {
		if err := payment.Validate(); err != nil {
			return fmt.Errorf("invalid payment at index %d: %w", i, err)
		}
		
		// Check for duplicate payment IDs
		if paymentIDs[payment.Id] {
			return fmt.Errorf("duplicate payment ID: %s", payment.Id)
		}
		paymentIDs[payment.Id] = true
	}
	
	// Validate vaults
	vaultIDs := make(map[string]bool)
	for i, vault := range gs.VaultList {
		if err := vault.Validate(); err != nil {
			return fmt.Errorf("invalid vault at index %d: %w", i, err)
		}
		
		// Check for duplicate vault IDs
		if vaultIDs[vault.Id] {
			return fmt.Errorf("duplicate vault ID: %s", vault.Id)
		}
		vaultIDs[vault.Id] = true
	}
	
	// Validate reward pools
	poolIDs := make(map[string]bool)
	for i, pool := range gs.PoolList {
		if err := pool.Validate(); err != nil {
			return fmt.Errorf("invalid reward pool at index %d: %w", i, err)
		}
		
		// Check for duplicate pool IDs
		if poolIDs[pool.Id] {
			return fmt.Errorf("duplicate pool ID: %s", pool.Id)
		}
		poolIDs[pool.Id] = true
	}
	
	return nil
}
