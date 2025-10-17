package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/vitacoin/vitacoin/vitacoin/vitacoin/x/vitacoin/types"
)

// RegisterInvariants registers all vitacoin module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "total-supply", TotalSupplyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "payment-consistency", PaymentConsistencyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "vault-consistency", VaultConsistencyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "reward-pool-consistency", RewardPoolConsistencyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "merchant-stake-consistency", MerchantStakeConsistencyInvariant(k))
}

// TotalSupplyInvariant checks that the total supply is consistent
func TotalSupplyInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// This invariant will check total supply consistency
		// Implementation will be enhanced when token economics are implemented
		return "", false
	}
}

// PaymentConsistencyInvariant checks payment state consistency
func PaymentConsistencyInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		payments, err := k.GetAllPayments(ctx)
		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName, "payment-consistency",
				fmt.Sprintf("failed to get payments: %v", err),
			), true
		}

		var msg string
		broken := false

		for _, payment := range payments {
			// Check payment status consistency
			if payment.Status == types.PaymentStatusCompleted && payment.CompletionHeight == 0 {
				broken = true
				msg += fmt.Sprintf("Payment %s is completed but has no completion height\n", payment.Id)
			}

			// Check that completion height is not before creation height
			if payment.CompletionHeight > 0 && payment.CompletionHeight < payment.CreationHeight {
				broken = true
				msg += fmt.Sprintf("Payment %s completion height (%d) is before creation height (%d)\n",
					payment.Id, payment.CompletionHeight, payment.CreationHeight)
			}

			// Check amount consistency
			if payment.Amount.IsNil() || payment.Amount.IsNegative() {
				broken = true
				msg += fmt.Sprintf("Payment %s has invalid amount: %s\n", payment.Id, payment.Amount)
			}

			// TODO: Check fee consistency when Fee field is added to Payment proto
			// if payment.Fee.IsNil() || payment.Fee.IsNegative() {
			//	broken = true
			//	msg += fmt.Sprintf("Payment %s has invalid fee: %s\n", payment.Id, payment.Fee)
			// }

			// Check address validity
			if _, err := sdk.AccAddressFromBech32(payment.FromAddress); err != nil {
				broken = true
				msg += fmt.Sprintf("Payment %s has invalid from address: %s\n", payment.Id, payment.FromAddress)
			}

			if _, err := sdk.AccAddressFromBech32(payment.ToAddress); err != nil {
				broken = true
				msg += fmt.Sprintf("Payment %s has invalid to address: %s\n", payment.Id, payment.ToAddress)
			}
		}

		return sdk.FormatInvariant(
			types.ModuleName, "payment-consistency",
			fmt.Sprintf("Payment consistency check:\n%s", msg),
		), broken
	}
}

// VaultConsistencyInvariant checks vault state consistency
func VaultConsistencyInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		vaults, err := k.GetAllVaults(ctx)
		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName, "vault-consistency",
				fmt.Sprintf("failed to get vaults: %v", err),
			), true
		}

		var msg string
		broken := false

		for _, vault := range vaults {
			// Check amount consistency
			if vault.Amount.IsNil() || vault.Amount.IsNegative() || vault.Amount.IsZero() {
				broken = true
				msg += fmt.Sprintf("Vault %s has invalid amount: %s\n", vault.Id, vault.Amount)
			}

			// Check unlock height consistency
			if vault.UnlockHeight <= vault.CreationHeight {
				broken = true
				msg += fmt.Sprintf("Vault %s unlock height (%d) is not after creation height (%d)\n",
					vault.Id, vault.UnlockHeight, vault.CreationHeight)
			}

			// Check owner address validity
			if _, err := sdk.AccAddressFromBech32(vault.Owner); err != nil {
				broken = true
				msg += fmt.Sprintf("Vault %s has invalid owner address: %s\n", vault.Id, vault.Owner)
			}

			// Check reward multiplier consistency
			if vault.RewardMultiplier.IsNil() || vault.RewardMultiplier.IsNegative() {
				broken = true
				msg += fmt.Sprintf("Vault %s has invalid reward multiplier: %s\n", vault.Id, vault.RewardMultiplier)
			}

			// TODO: Check withdrawal consistency when IsActive field is properly implemented
			// currentHeight := ctx.BlockHeight()
			// if !vault.IsActive && currentHeight < vault.UnlockHeight {
			//	broken = true
			//	msg += fmt.Sprintf("Vault %s is withdrawn before unlock height (%d < %d)\n",
			//		vault.Id, currentHeight, vault.UnlockHeight)
			// }
		}

		return sdk.FormatInvariant(
			types.ModuleName, "vault-consistency",
			fmt.Sprintf("Vault consistency check:\n%s", msg),
		), broken
	}
}

// RewardPoolConsistencyInvariant checks reward pool state consistency
func RewardPoolConsistencyInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		pools, err := k.GetAllRewardPools(ctx)
		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName, "reward-pool-consistency",
				fmt.Sprintf("failed to get reward pools: %v", err),
			), true
		}

		var msg string
		broken := false

		for _, pool := range pools {
			// Check amounts consistency
			if pool.TotalRewards.IsNil() || pool.TotalRewards.IsNegative() {
				broken = true
				msg += fmt.Sprintf("Pool %s has invalid total rewards: %s\n", pool.Id, pool.TotalRewards)
			}

			if pool.DistributedRewards.IsNil() || pool.DistributedRewards.IsNegative() {
				broken = true
				msg += fmt.Sprintf("Pool %s has invalid distributed rewards: %s\n", pool.Id, pool.DistributedRewards)
			}

			// Check that distributed rewards don't exceed total rewards
			if pool.DistributedRewards.GT(pool.TotalRewards) {
				broken = true
				msg += fmt.Sprintf("Pool %s distributed rewards (%s) exceed total rewards (%s)\n",
					pool.Id, pool.DistributedRewards, pool.TotalRewards)
			}

			// Check height consistency
			if pool.EndHeight != 0 && pool.EndHeight <= pool.StartHeight {
				broken = true
				msg += fmt.Sprintf("Pool %s end height (%d) is not after start height (%d)\n",
					pool.Id, pool.EndHeight, pool.StartHeight)
			}

			// Check merchant address validity
			if _, err := sdk.AccAddressFromBech32(pool.MerchantAddress); err != nil {
				broken = true
				msg += fmt.Sprintf("Pool %s has invalid merchant address: %s\n", pool.Id, pool.MerchantAddress)
			}

			// Check active status consistency
			currentHeight := ctx.BlockHeight()
			shouldBeActive := currentHeight >= pool.StartHeight && 
				(pool.EndHeight == 0 || currentHeight < pool.EndHeight)
			
			if pool.IsActive != shouldBeActive {
				broken = true
				msg += fmt.Sprintf("Pool %s active status (%t) inconsistent with height range (%d-%d, current: %d)\n",
					pool.Id, pool.IsActive, pool.StartHeight, pool.EndHeight, currentHeight)
			}
		}

		return sdk.FormatInvariant(
			types.ModuleName, "reward-pool-consistency",
			fmt.Sprintf("Reward pool consistency check:\n%s", msg),
		), broken
	}
}

// MerchantStakeConsistencyInvariant checks merchant stake consistency
func MerchantStakeConsistencyInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		merchants, err := k.GetAllMerchants(ctx)
		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName, "merchant-stake-consistency",
				fmt.Sprintf("failed to get merchants: %v", err),
			), true
		}

		var msg string
		broken := false

		for _, merchant := range merchants {
			// Check stake amount consistency
			if merchant.StakeAmount.IsNil() || merchant.StakeAmount.IsNegative() {
				broken = true
				msg += fmt.Sprintf("Merchant %s has invalid stake amount: %s\n", merchant.Address, merchant.StakeAmount)
			}

			// Check total volume consistency
			if merchant.TotalVolume.IsNil() || merchant.TotalVolume.IsNegative() {
				broken = true
				msg += fmt.Sprintf("Merchant %s has invalid total volume: %s\n", merchant.Address, merchant.TotalVolume)
			}

			// Check address validity
			if _, err := sdk.AccAddressFromBech32(merchant.Address); err != nil {
				broken = true
				msg += fmt.Sprintf("Merchant has invalid address: %s\n", merchant.Address)
			}

			// Check tier consistency with volume
			expectedTier := k.calculateMerchantTier(merchant.TotalVolume)
			if merchant.Tier != expectedTier {
				broken = true
				msg += fmt.Sprintf("Merchant %s tier (%s) inconsistent with volume (%s), expected (%s)\n",
					merchant.Address, merchant.Tier, merchant.TotalVolume, expectedTier)
			}

			// Check business name
			if merchant.BusinessName == "" {
				broken = true
				msg += fmt.Sprintf("Merchant %s has empty business name\n", merchant.Address)
			}

			// Check registration height
			if merchant.RegistrationHeight < 0 {
				broken = true
				msg += fmt.Sprintf("Merchant %s has invalid registration height: %d\n", 
					merchant.Address, merchant.RegistrationHeight)
			}
		}

		return sdk.FormatInvariant(
			types.ModuleName, "merchant-stake-consistency",
			fmt.Sprintf("Merchant stake consistency check:\n%s", msg),
		), broken
	}
}

// AllInvariants runs all registered invariants for the vitacoin module
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var res string
		var broken bool

		// Run total supply invariant
		if msg, broke := TotalSupplyInvariant(k)(ctx); broke {
			broken = true
			res += msg
		}

		// Run payment consistency invariant
		if msg, broke := PaymentConsistencyInvariant(k)(ctx); broke {
			broken = true
			res += msg
		}

		// Run vault consistency invariant
		if msg, broke := VaultConsistencyInvariant(k)(ctx); broke {
			broken = true
			res += msg
		}

		// Run reward pool consistency invariant
		if msg, broke := RewardPoolConsistencyInvariant(k)(ctx); broke {
			broken = true
			res += msg
		}

		// Run merchant stake consistency invariant
		if msg, broke := MerchantStakeConsistencyInvariant(k)(ctx); broke {
			broken = true
			res += msg
		}

		return res, broken
	}
}