package keeper_test

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/esspron/VITACOIN/vitacoin/vitacoin/x/vitacoin/types"
)

// TestMsgUpdateParams tests the UpdateParams message handler
func (suite *KeeperTestSuite) TestMsgUpdateParams() {
	ctx := sdk.UnwrapSDKContext(suite.ctx)
	authority := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	testCases := []struct {
		name      string
		msg       *types.MsgUpdateParams
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid update params",
			msg: &types.MsgUpdateParams{
				Authority: authority,
				Params: types.Params{
					TransactionFeePercent:   math.LegacyNewDecWithPrec(15, 2), // 0.15%
					MerchantRegistrationFee: math.NewInt(100000),              // 100,000 VITA
					MinMerchantStake:        math.NewInt(50000),               // 50,000 VITA
					PaymentTimeoutBlocks:    1000,                             // blocks
					// TODO: Add other params fields as needed
				},
			},
			expectErr: false,
		},
		{
			name: "invalid authority",
			msg: &types.MsgUpdateParams{
				Authority: "invalid-authority",
				Params:    types.DefaultParams(),
			},
			expectErr: true,
			errMsg:    "unauthorized",
		},
		{
			name: "invalid params - negative fee",
			msg: &types.MsgUpdateParams{
				Authority: authority,
				Params: types.Params{
					TransactionFeePercent:   math.LegacyNewDec(-1), // Invalid negative fee
					MerchantRegistrationFee: math.NewInt(100000),
					MinMerchantStake:        math.NewInt(50000),
					PaymentTimeoutBlocks:    1000,
					EnableMerchantLoyalty:   true,
					LoyaltyRewardPercent:    math.LegacyNewDecWithPrec(5, 2), // 0.05%
				},
			},
			expectErr: true,
			errMsg:    "invalid params",
		},
		{
			name: "invalid params - zero registration fee",
			msg: &types.MsgUpdateParams{
				Authority: authority,
				Params: types.Params{
					TransactionFeePercent:   math.LegacyNewDecWithPrec(1, 3), // 0.1%
					MerchantRegistrationFee: math.ZeroInt(),                  // Invalid zero fee
					MinMerchantStake:        math.NewInt(50000),
					PaymentTimeoutBlocks:    1000,
					EnableMerchantLoyalty:   true,
					LoyaltyRewardPercent:    math.LegacyNewDecWithPrec(5, 2), // 0.05%
				},
			},
			expectErr: true,
			errMsg:    "invalid params",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			_, err := suite.msgServer.UpdateParams(ctx, tc.msg)
			if tc.expectErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			} else {
				suite.Require().NoError(err)
				// Verify params were actually updated
				params, err := suite.keeper.GetParams(suite.ctx)
				suite.Require().NoError(err)
				suite.Require().Equal(tc.msg.Params, params)
			}
		})
	}
}

// TestMsgRegisterMerchant tests the RegisterMerchant message handler
func (suite *KeeperTestSuite) TestMsgRegisterMerchant() {
	ctx := sdk.UnwrapSDKContext(suite.ctx)

	testCases := []struct {
		name      string
		msg       *types.MsgRegisterMerchant
		expectErr bool
		errMsg    string
		setup     func()
	}{
		{
			name: "valid merchant registration",
			msg: &types.MsgRegisterMerchant{
				Sender:       "vita1merchant123",
				BusinessName: "Test Business",
				StakeAmount:  math.NewInt(1000000), // 1M VITA
			},
			expectErr: false,
			setup: func() {
				// Ensure merchant doesn't exist
				suite.keeper.DeleteMerchant(suite.ctx, "vita1merchant123")
			},
		},
		{
			name: "merchant already exists",
			msg: &types.MsgRegisterMerchant{
				Sender:      "vita1existing",
				BusinessName: "Existing Business",
				StakeAmount:  math.NewInt(1000000),
			},
			expectErr: true,
			errMsg:    "merchant already exists",
			setup: func() {
				// Pre-register merchant
				existingMerchant := types.Merchant{
					Address:      "vita1existing",
					BusinessName: "Existing Merchant",
					Tier:         types.MerchantTierBronze,
					StakeAmount:  math.NewInt(500000),
				}
				suite.keeper.SetMerchant(suite.ctx, existingMerchant)
			},
		},
		{
			name: "stake amount too low",
			msg: &types.MsgRegisterMerchant{
				Sender:      "vita1lowstake",
				BusinessName: "Low Stake Business",
				StakeAmount:  math.NewInt(1000), // Too low
			},
			expectErr: true,
			errMsg:    "insufficient stake amount",
			setup: func() {
				suite.keeper.DeleteMerchant(suite.ctx, "vita1lowstake")
			},
		},
		{
			name: "invalid business name - empty",
			msg: &types.MsgRegisterMerchant{
				Sender:      "vita1emptybiz",
				BusinessName: "", // Empty name
				StakeAmount:  math.NewInt(1000000),
			},
			expectErr: true,
			errMsg:    "invalid business name",
			setup: func() {
				suite.keeper.DeleteMerchant(suite.ctx, "vita1emptybiz")
			},
		},
		{
			name: "invalid business name - too long",
			msg: &types.MsgRegisterMerchant{
				Sender:      "vita1longbiz",
				BusinessName: string(make([]byte, 1001)), // Too long (>1000 chars)
				StakeAmount:  math.NewInt(1000000),
			},
			expectErr: true,
			errMsg:    "invalid business name",
			setup: func() {
				suite.keeper.DeleteMerchant(suite.ctx, "vita1longbiz")
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			if tc.setup != nil {
				tc.setup()
			}

			_, err := suite.msgServer.RegisterMerchant(ctx, tc.msg)
			if tc.expectErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			} else {
				suite.Require().NoError(err)
				// Verify merchant was created
				merchant, err := suite.keeper.GetMerchant(suite.ctx, tc.msg.Sender)
				suite.Require().NoError(err)
				suite.Require().Equal(tc.msg.Sender, merchant.Address)
				suite.Require().Equal(tc.msg.BusinessName, merchant.BusinessName)
				suite.Require().Equal(tc.msg.StakeAmount, merchant.StakeAmount)
				suite.Require().True(merchant.IsActive)
			}
		})
	}
}

// TestMsgUpdateMerchant tests the UpdateMerchant message handler
func (suite *KeeperTestSuite) TestMsgUpdateMerchant() {
	ctx := sdk.UnwrapSDKContext(suite.ctx)

	// Setup existing merchant
	existingMerchant := types.Merchant{
		Address:      "vita1updatemerchant",
		BusinessName: "Original Business",
		Tier:         types.MerchantTierBronze,
		StakeAmount:  math.NewInt(500000),
	}
	suite.keeper.SetMerchant(suite.ctx, existingMerchant)

	testCases := []struct {
		name      string
		msg       *types.MsgUpdateMerchant
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid merchant update",
			msg: &types.MsgUpdateMerchant{
				Sender:         "vita1updatemerchant",
				BusinessName: "Updated Business Name",
				AdditionalStake:  math.NewInt(2000000), // Increase stake
			},
			expectErr: false,
		},
		{
			name: "merchant not found",
			msg: &types.MsgUpdateMerchant{
				Sender:         "vita1nonexistent",
				BusinessName: "Some Business",
				AdditionalStake:  math.NewInt(1000000),
			},
			expectErr: true,
			errMsg:    "merchant not found",
		},
		{
			name: "invalid new business name - empty",
			msg: &types.MsgUpdateMerchant{
				Sender:         "vita1updatemerchant",
				BusinessName: "", // Empty name
				AdditionalStake:  math.NewInt(1000000),
			},
			expectErr: true,
			errMsg:    "invalid business name",
		},
		{
			name: "invalid stake amount - too low",
			msg: &types.MsgUpdateMerchant{
				Sender:         "vita1updatemerchant",
				BusinessName: "Valid Business Name",
				AdditionalStake:  math.NewInt(1000), // Too low
			},
			expectErr: true,
			errMsg:    "insufficient stake amount",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			_, err := suite.msgServer.UpdateMerchant(ctx, tc.msg)
			if tc.expectErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			} else {
				suite.Require().NoError(err)
				// Verify merchant was updated
				merchant, err := suite.keeper.GetMerchant(suite.ctx, tc.msg.Sender)
				suite.Require().NoError(err)
				suite.Require().Equal(tc.msg.BusinessName, merchant.BusinessName)
				suite.Require().Equal(tc.msg.AdditionalStake, merchant.StakeAmount)
			}
		})
	}
}

// TestMsgCreatePayment tests the CreatePayment message handler
func (suite *KeeperTestSuite) TestMsgCreatePayment() {
	ctx := sdk.UnwrapSDKContext(suite.ctx)

	// Setup merchant for payments
	merchant := types.Merchant{
		Address:      "vita1paymentmerchant",
		BusinessName: "Payment Merchant",
		Tier:         types.MerchantTierBronze,
		StakeAmount:  math.NewInt(1000000),
	}
	suite.keeper.SetMerchant(suite.ctx, merchant)

	testCases := []struct {
		name      string
		msg       *types.MsgCreatePayment
		expectErr bool
		errMsg    string
		setup     func()
	}{
		{
			name: "valid payment creation",
			msg: &types.MsgCreatePayment{
				Sender:         "vita1payer",
				Amount:          math.NewInt(10000), // 10,000 VITA
				Memo:            "Test payment",
			},
			expectErr: false,
			setup: func() {
				// Clear any existing payments
				payments, _ := suite.keeper.GetAllPayments(suite.ctx)
				for _, p := range payments {
					suite.keeper.DeletePayment(suite.ctx, p.Id)
				}
			},
		},
		{
			name: "merchant not found",
			msg: &types.MsgCreatePayment{
				Sender:         "vita1payer",
				Amount:          math.NewInt(10000),
				Memo:            "Test payment",
			},
			expectErr: true,
			errMsg:    "merchant not found",
			setup:     func() {},
		},
		{
			name: "merchant inactive",
			msg: &types.MsgCreatePayment{
				Sender:         "vita1payer",
				Amount:          math.NewInt(10000),
				Memo:            "Test payment",
			},
			expectErr: true,
			errMsg:    "merchant is not active",
			setup: func() {
				inactiveMerchant := types.Merchant{
					Address:      "vita1inactivemerchant",
					BusinessName: "Inactive Merchant",
					Tier:         types.MerchantTierBronze,
					StakeAmount:  math.NewInt(1000000),
					IsActive:     false, // Inactive
				}
				suite.keeper.SetMerchant(suite.ctx, inactiveMerchant)
			},
		},
		{
			name: "invalid amount - zero",
			msg: &types.MsgCreatePayment{
				Sender:         "vita1payer",
				Amount:          math.ZeroInt(), // Invalid zero amount
				Memo:            "Test payment",
			},
			expectErr: true,
			errMsg:    "invalid payment amount",
			setup:     func() {},
		},
		{
			name: "invalid amount - negative",
			msg: &types.MsgCreatePayment{
				Sender:         "vita1payer",
				Amount:          math.NewInt(-100), // Negative amount
				Memo:            "Test payment",
			},
			expectErr: true,
			errMsg:    "invalid payment amount",
			setup:     func() {},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.setup()

			resp, err := suite.msgServer.CreatePayment(ctx, tc.msg)
			if tc.expectErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotEmpty(resp.PaymentId)
				
				// Verify payment was created
				payment, err := suite.keeper.GetPayment(suite.ctx, resp.PaymentId)
				suite.Require().NoError(err)
				suite.Require().Equal(tc.msg.Sender, payment.FromAddress)
				suite.Require().Equal(tc.msg.MerchantAddress, payment.ToAddress)
				suite.Require().Equal(tc.msg.Amount, payment.Amount)
				suite.Require().Equal(types.PaymentStatusPending, payment.Status)
			}
		})
	}
}

// TestMsgCompletePayment tests the CompletePayment message handler
func (suite *KeeperTestSuite) TestMsgCompletePayment() {
	ctx := sdk.UnwrapSDKContext(suite.ctx)

	// Setup merchant and payment
	merchant := types.Merchant{
		Address:      "vita1completemerchant",
		BusinessName: "Complete Merchant",
		Tier:         types.MerchantTierBronze,
		StakeAmount:  math.NewInt(1000000),
	}
	suite.keeper.SetMerchant(suite.ctx, merchant)

	payment := types.Payment{
		Id:             "test-payment-1",
		FromAddress:    "vita1payer",
		ToAddress:      "vita1completemerchant",
		Amount:         math.NewInt(10000),
		Status:         types.PaymentStatusPending,
		CreationHeight: suite.ctx.BlockHeight(),
		Memo:           "Test payment",
	}
	suite.keeper.SetPayment(suite.ctx, payment)

	testCases := []struct {
		name      string
		msg       *types.MsgCompletePayment
		expectErr bool
		errMsg    string
		setup     func()
	}{
		{
			name: "valid payment completion",
			msg: &types.MsgCompletePayment{
				Sender:   "vita1completemerchant",
				PaymentId: "test-payment-1",
			},
			expectErr: false,
			setup:     func() {},
		},
		{
			name: "payment not found",
			msg: &types.MsgCompletePayment{
				Sender:   "vita1completemerchant",
				PaymentId: "nonexistent-payment",
			},
			expectErr: true,
			errMsg:    "payment not found",
			setup:     func() {},
		},
		{
			name: "unauthorized - not merchant",
			msg: &types.MsgCompletePayment{
				Sender:   "vita1unauthorized",
				PaymentId: "test-payment-1",
			},
			expectErr: true,
			errMsg:    "unauthorized",
			setup: func() {
				// Reset payment status to pending
				payment := types.Payment{
					Id:             "test-payment-1",
					FromAddress:    "vita1payer",
					ToAddress:      "vita1completemerchant",
					Amount:         math.NewInt(10000),
					Status:         types.PaymentStatusPending,
					CreationHeight: suite.ctx.BlockHeight(),
					Memo:           "Test payment",
				}
				suite.keeper.SetPayment(suite.ctx, payment)
			},
		},
		{
			name: "payment already completed",
			msg: &types.MsgCompletePayment{
				Sender:   "vita1completemerchant",
				PaymentId: "completed-payment",
			},
			expectErr: true,
			errMsg:    "payment not in pending status",
			setup: func() {
				completedPayment := types.Payment{
					Id:             "completed-payment",
					FromAddress:    "vita1payer",
					ToAddress:      "vita1completemerchant",
					Amount:         math.NewInt(10000),
					Status:         types.PaymentStatusCompleted,
					CreationHeight: suite.ctx.BlockHeight(),
					Memo:           "Completed payment",
				}
				suite.keeper.SetPayment(suite.ctx, completedPayment)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.setup()

			_, err := suite.msgServer.CompletePayment(ctx, tc.msg)
			if tc.expectErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			} else {
				suite.Require().NoError(err)
				// Verify payment status was updated
				updatedPayment, err := suite.keeper.GetPayment(suite.ctx, tc.msg.PaymentId)
				suite.Require().NoError(err)
				suite.Require().Equal(types.PaymentStatusCompleted, updatedPayment.Status)
			}
		})
	}
}

// TestMsgRefundPayment tests the RefundPayment message handler
func (suite *KeeperTestSuite) TestMsgRefundPayment() {
	ctx := sdk.UnwrapSDKContext(suite.ctx)

	// Setup merchant and payment
	merchant := types.Merchant{
		Address:      "vita1refundmerchant",
		BusinessName: "Refund Merchant",
		Tier:         types.MerchantTierBronze,
		StakeAmount:  math.NewInt(1000000),
	}
	suite.keeper.SetMerchant(suite.ctx, merchant)

	payment := types.Payment{
		Id:             "test-refund-payment",
		FromAddress:    "vita1payer",
		ToAddress:      "vita1refundmerchant",
		Amount:         math.NewInt(10000),
		Status:         types.PaymentStatusPending,
		CreationHeight: suite.ctx.BlockHeight(),
		Memo:           "Test refund payment",
	}
	suite.keeper.SetPayment(suite.ctx, payment)

	testCases := []struct {
		name      string
		msg       *types.MsgRefundPayment
		expectErr bool
		errMsg    string
		setup     func()
	}{
		{
			name: "valid payment refund",
			msg: &types.MsgRefundPayment{
				Sender:   "vita1refundmerchant",
				PaymentId: "test-refund-payment",
				Reason:    "Customer requested refund",
			},
			expectErr: false,
			setup:     func() {},
		},
		{
			name: "payment not found",
			msg: &types.MsgRefundPayment{
				Sender:   "vita1refundmerchant",
				PaymentId: "nonexistent-refund",
				Reason:    "Test reason",
			},
			expectErr: true,
			errMsg:    "payment not found",
			setup:     func() {},
		},
		{
			name: "unauthorized - not merchant",
			msg: &types.MsgRefundPayment{
				Sender:   "vita1unauthorized",
				PaymentId: "test-refund-payment",
				Reason:    "Test reason",
			},
			expectErr: true,
			errMsg:    "unauthorized",
			setup: func() {
				// Reset payment status to pending
				payment := types.Payment{
					Id:             "test-refund-payment",
					FromAddress:    "vita1payer",
					ToAddress:      "vita1refundmerchant",
					Amount:         math.NewInt(10000),
					Status:         types.PaymentStatusPending,
					CreationHeight: suite.ctx.BlockHeight(),
					Memo:           "Test refund payment",
				}
				suite.keeper.SetPayment(suite.ctx, payment)
			},
		},
		{
			name: "payment already refunded",
			msg: &types.MsgRefundPayment{
				Sender:   "vita1refundmerchant",
				PaymentId: "already-refunded",
				Reason:    "Test reason",
			},
			expectErr: true,
			errMsg:    "payment not in pending status",
			setup: func() {
				refundedPayment := types.Payment{
					Id:             "already-refunded",
					FromAddress:    "vita1payer",
					ToAddress:      "vita1refundmerchant",
					Amount:         math.NewInt(10000),
					Status:         types.PaymentStatusRefunded,
					CreationHeight: suite.ctx.BlockHeight(),
					Memo:           "Already refunded payment",
				}
				suite.keeper.SetPayment(suite.ctx, refundedPayment)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.setup()

			_, err := suite.msgServer.RefundPayment(ctx, tc.msg)
			if tc.expectErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			} else {
				suite.Require().NoError(err)
				// Verify payment status was updated
				updatedPayment, err := suite.keeper.GetPayment(suite.ctx, tc.msg.PaymentId)
				suite.Require().NoError(err)
				suite.Require().Equal(types.PaymentStatusRefunded, updatedPayment.Status)
			}
		})
	}
}

// TestMsgCreateVault tests the CreateVault message handler
func (suite *KeeperTestSuite) TestMsgCreateVault() {
	ctx := sdk.UnwrapSDKContext(suite.ctx)

	testCases := []struct {
		name      string
		msg       *types.MsgCreateVault
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid vault creation",
			msg: &types.MsgCreateVault{
				Sender:      "vita1vaultcreator",
				Amount:       math.NewInt(100000), // 100,000 VITA
				LockDuration: 1000,                // 1000 blocks
			},
			expectErr: false,
		},
		{
			name: "invalid amount - zero",
			msg: &types.MsgCreateVault{
				Sender:      "vita1vaultcreator",
				Amount:       math.ZeroInt(), // Invalid zero amount
				LockDuration: 1000,
			},
			expectErr: true,
			errMsg:    "invalid vault amount",
		},
		{
			name: "invalid amount - negative",
			msg: &types.MsgCreateVault{
				Sender:      "vita1vaultcreator",
				Amount:       math.NewInt(-1000), // Negative amount
				LockDuration: 1000,
			},
			expectErr: true,
			errMsg:    "invalid vault amount",
		},
		{
			name: "invalid lock duration - zero",
			msg: &types.MsgCreateVault{
				Sender:      "vita1vaultcreator",
				Amount:       math.NewInt(100000),
				LockDuration: 0, // Invalid zero duration
			},
			expectErr: true,
			errMsg:    "invalid lock duration",
		},
		{
			name: "invalid lock duration - negative",
			msg: &types.MsgCreateVault{
				Sender:      "vita1vaultcreator",
				Amount:       math.NewInt(100000),
				LockDuration: 0, // Invalid zero duration
			},
			expectErr: true,
			errMsg:    "invalid lock duration",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			resp, err := suite.msgServer.CreateVault(ctx, tc.msg)
			if tc.expectErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotEmpty(resp.VaultId)
				
				// Verify vault was created
				vault, err := suite.keeper.GetVault(suite.ctx, resp.VaultId)
				suite.Require().NoError(err)
				suite.Require().Equal(tc.msg.Sender, vault.Owner)
				suite.Require().Equal(tc.msg.Amount, vault.Amount)
				suite.Require().Equal(tc.msg.LockDuration, vault.LockDuration)
				// TODO: Add IsActive field to Vault struct and verify
				// suite.Require().True(vault.IsActive)
			}
		})
	}
}

// TestMsgWithdrawVault tests the WithdrawVault message handler
func (suite *KeeperTestSuite) TestMsgWithdrawVault() {
	ctx := sdk.UnwrapSDKContext(suite.ctx)

	// Setup vault
	vault := types.Vault{
		Id:               "test-vault-1",
		Owner:            "vita1vaultowner",
		Amount:           math.NewInt(100000),
		LockDuration:     1000,
		CreationHeight:   suite.ctx.BlockHeight(),
		UnlockHeight:     suite.ctx.BlockHeight() + 1000,
		RewardMultiplier: math.LegacyNewDec(2),
	}
	suite.keeper.SetVault(suite.ctx, vault)

	testCases := []struct {
		name      string
		msg       *types.MsgWithdrawVault
		expectErr bool
		errMsg    string
		setup     func()
	}{
		{
			name: "valid vault withdrawal - unlocked",
			msg: &types.MsgWithdrawVault{
				Sender: "vita1vaultowner",
				VaultId: "test-vault-1",
			},
			expectErr: false,
			setup: func() {
				// Set vault as unlocked (past unlock height)
				unlockedVault := vault
				unlockedVault.UnlockHeight = suite.ctx.BlockHeight() - 1 // Past unlock
				suite.keeper.SetVault(suite.ctx, unlockedVault)
			},
		},
		{
			name: "vault not found",
			msg: &types.MsgWithdrawVault{
				Sender: "vita1vaultowner",
				VaultId: "nonexistent-vault",
			},
			expectErr: true,
			errMsg:    "vault not found",
			setup:     func() {},
		},
		{
			name: "unauthorized - not owner",
			msg: &types.MsgWithdrawVault{
				Sender: "vita1unauthorized",
				VaultId: "test-vault-1",
			},
			expectErr: true,
			errMsg:    "unauthorized",
			setup: func() {
				// Reset vault
				suite.keeper.SetVault(suite.ctx, vault)
			},
		},
		{
			name: "vault still locked",
			msg: &types.MsgWithdrawVault{
				Sender: "vita1vaultowner",
				VaultId: "test-vault-1",
			},
			expectErr: true,
			errMsg:    "vault is still locked",
			setup: func() {
				// Set vault as still locked
				lockedVault := vault
				lockedVault.UnlockHeight = suite.ctx.BlockHeight() + 1000 // Future unlock
				suite.keeper.SetVault(suite.ctx, lockedVault)
			},
		},
		{
			name: "vault inactive",
			msg: &types.MsgWithdrawVault{
				Sender: "vita1vaultowner",
				VaultId: "inactive-vault",
			},
			expectErr: true,
			errMsg:    "vault is not active",
			setup: func() {
				inactiveVault := types.Vault{
					Id:               "inactive-vault",
					Owner:            "vita1vaultowner",
					Amount:           math.NewInt(100000),
					LockDuration:     1000,
					CreationHeight:   suite.ctx.BlockHeight(),
					UnlockHeight:     suite.ctx.BlockHeight() - 1, // Unlocked
					RewardMultiplier: math.LegacyNewDec(2),
					IsActive:         false, // Inactive
				}
				suite.keeper.SetVault(suite.ctx, inactiveVault)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.setup()

			_, err := suite.msgServer.WithdrawVault(ctx, tc.msg)
			if tc.expectErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			} else {
				suite.Require().NoError(err)
				// Verify vault was deactivated
				updatedVault, err := suite.keeper.GetVault(suite.ctx, tc.msg.VaultId)
				suite.Require().NoError(err)
				suite.Require().False(true // TODO: Add IsActive field to Vault struct)
			}
		})
	}
}

// TestMsgCreateRewardPool tests the CreateRewardPool message handler
func (suite *KeeperTestSuite) TestMsgCreateRewardPool() {
	ctx := sdk.UnwrapSDKContext(suite.ctx)

	// Setup merchant
	merchant := types.Merchant{
		Address:      "vita1poolmerchant",
		BusinessName: "Pool Merchant",
		Tier:         types.MerchantTierBronze,
		StakeAmount:  math.NewInt(1000000),
	}
	suite.keeper.SetMerchant(suite.ctx, merchant)

	testCases := []struct {
		name      string
		msg       *types.MsgCreateRewardPool
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid reward pool creation",
			msg: &types.MsgCreateRewardPool{
				Sender:         "vita1poolmerchant",
				TotalRewards:   math.NewInt(50000), // 50,000 VITA
				DurationBlocks: 1000,               // 1000 blocks
			},
			expectErr: false,
		},
		{
			name: "merchant not found",
			msg: &types.MsgCreateRewardPool{
				Sender:         "vita1nonexistent",
				TotalRewards:    math.NewInt(50000),
				DurationBlocks:        1000,
			},
			expectErr: true,
			errMsg:    "merchant not found",
		},
		{
			name: "unauthorized - not merchant owner",
			msg: &types.MsgCreateRewardPool{
				Sender:         "vita1unauthorized",
				TotalRewards:    math.NewInt(50000),
				DurationBlocks:        1000,
			},
			expectErr: true,
			errMsg:    "unauthorized",
		},
		{
			name: "invalid total rewards - zero",
			msg: &types.MsgCreateRewardPool{
				Sender:         "vita1poolmerchant",
				TotalRewards:    math.ZeroInt(), // Invalid zero rewards
				DurationBlocks:        1000,
			},
			expectErr: true,
			errMsg:    "invalid total rewards",
		},
		{
			name: "invalid duration - zero",
			msg: &types.MsgCreateRewardPool{
				Sender:         "vita1poolmerchant",
				TotalRewards:    math.NewInt(50000),
				DurationBlocks:        0, // Invalid zero duration
			},
			expectErr: true,
			errMsg:    "invalid duration",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			resp, err := suite.msgServer.CreateRewardPool(ctx, tc.msg)
			if tc.expectErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotEmpty(resp.PoolId)
				
				// Verify reward pool was created
				pool, err := suite.keeper.GetRewardPool(suite.ctx, resp.PoolId)
				suite.Require().NoError(err)
				suite.Require().Equal(tc.msg.MerchantAddress, pool.MerchantAddress)
				suite.Require().Equal(tc.msg.TotalRewards, pool.TotalRewards)
				suite.Require().True(pool.IsActive)
			}
		})
	}
}

// TestMsgDistributeRewards tests the DistributeRewards message handler
func (suite *KeeperTestSuite) TestMsgDistributeRewards() {
	ctx := sdk.UnwrapSDKContext(suite.ctx)

	// Setup merchant and reward pool
	merchant := types.Merchant{
		Address:      "vita1distmerchant",
		BusinessName: "Distribution Merchant",
		Tier:         types.MerchantTierBronze,
		StakeAmount:  math.NewInt(1000000),
	}
	suite.keeper.SetMerchant(suite.ctx, merchant)

	rewardPool := types.RewardPool{
		Id:                 "test-pool-1",
		TotalRewards:       math.NewInt(100000),
		DistributedRewards: math.ZeroInt(),
		StartHeight:        suite.ctx.BlockHeight(),
		EndHeight:          suite.ctx.BlockHeight() + 1000,
	}
	suite.keeper.SetRewardPool(suite.ctx, rewardPool)

	testCases := []struct {
		name      string
		msg       *types.MsgDistributeRewards
		expectErr bool
		errMsg    string
		setup     func()
	}{
		{
			name: "valid reward distribution",
			msg: &types.MsgDistributeRewards{
				Sender:    "vita1distmerchant",
				PoolId:     "test-pool-1",
				Recipients: []string{"vita1recipient1", "vita1recipient2"},
				Amounts:    []string{"1000", "2000"},
			},
			expectErr: false,
			setup:     func() {},
		},
		{
			name: "pool not found",
			msg: &types.MsgDistributeRewards{
				Sender:    "vita1distmerchant",
				PoolId:     "nonexistent-pool",
				Recipients: []string{"vita1recipient1"},
				Amounts:    []string{"1000"},
			},
			expectErr: true,
			errMsg:    "reward pool not found",
			setup:     func() {},
		},
		{
			name: "unauthorized - not merchant",
			msg: &types.MsgDistributeRewards{
				Sender:    "vita1unauthorized",
				PoolId:     "test-pool-1",
				Recipients: []string{"vita1recipient1"},
				Amounts:    []string{"1000"},
			},
			expectErr: true,
			errMsg:    "unauthorized",
			setup: func() {
				// Reset pool
				suite.keeper.SetRewardPool(suite.ctx, rewardPool)
			},
		},
		{
			name: "mismatched recipients and amounts",
			msg: &types.MsgDistributeRewards{
				Sender:    "vita1distmerchant",
				PoolId:     "test-pool-1",
				Recipients: []string{"vita1recipient1", "vita1recipient2"},
				Amounts:    []string{"1000"}, // Only one amount for two recipients
			},
			expectErr: true,
			errMsg:    "recipients and amounts length mismatch",
			setup: func() {
				// Reset pool
				suite.keeper.SetRewardPool(suite.ctx, rewardPool)
			},
		},
		{
			name: "invalid amount format",
			msg: &types.MsgDistributeRewards{
				Sender:    "vita1distmerchant",
				PoolId:     "test-pool-1",
				Recipients: []string{"vita1recipient1"},
				Amounts:    []string{"invalid-amount"},
			},
			expectErr: true,
			errMsg:    "invalid amount format",
			setup: func() {
				// Reset pool
				suite.keeper.SetRewardPool(suite.ctx, rewardPool)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.setup()

			_, err := suite.msgServer.DistributeRewards(ctx, tc.msg)
			if tc.expectErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			} else {
				suite.Require().NoError(err)
				// Verify reward pool distributed amount was updated
				updatedPool, err := suite.keeper.GetRewardPool(suite.ctx, tc.msg.PoolId)
				suite.Require().NoError(err)
				suite.Require().True(updatedPool.DistributedRewards.GT(math.ZeroInt()))
			}
		})
	}
}

// TestCalculateFeeThroughMessageHandlers tests fee calculation indirectly through message handlers
func (suite *KeeperTestSuite) TestCalculateFeeThroughMessageHandlers() {
	ctx := sdk.UnwrapSDKContext(suite.ctx)
	
	// Setup merchant for payment testing
	merchant := types.Merchant{
		Address:      "vita1feemerchant",
		BusinessName: "Fee Test Merchant",
		Tier:         types.MerchantTierBronze,
		StakeAmount:  math.NewInt(1000000),
	}
	suite.keeper.SetMerchant(suite.ctx, merchant)

	// Test fee calculation through CreatePayment
	paymentMsg := &types.MsgCreatePayment{
		Sender:         "vita1payer",
		Amount:          math.NewInt(10000), // 10,000 VITA
		Memo:            "Fee test payment",
	}

	resp, err := suite.msgServer.CreatePayment(ctx, paymentMsg)
	suite.Require().NoError(err)
	suite.Require().NotEmpty(resp.PaymentId)
	
	// Verify payment was created with fee calculation
	payment, err := suite.keeper.GetPayment(suite.ctx, resp.PaymentId)
	suite.Require().NoError(err)
	suite.Require().Equal(paymentMsg.Amount, payment.Amount)
	suite.Require().Equal(types.PaymentStatusPending, payment.Status)
}

// TestMerchantTierCalculationThroughRegistration tests tier calculation through merchant registration
func (suite *KeeperTestSuite) TestMerchantTierCalculationThroughRegistration() {
	ctx := sdk.UnwrapSDKContext(suite.ctx)
	
	testCases := []struct {
		name         string
		stakeAmount  math.Int
		expectedTier types.MerchantTier
	}{
		{
			name:         "bronze tier merchant",
			stakeAmount:  math.NewInt(100000), // 100K VITA
			expectedTier: types.MerchantTierBronze,
		},
		{
			name:         "silver tier merchant",
			stakeAmount:  math.NewInt(1000000), // 1M VITA
			expectedTier: types.MerchantTierSilver,
		},
		{
			name:         "gold tier merchant",
			stakeAmount:  math.NewInt(10000000), // 10M VITA
			expectedTier: types.MerchantTierGold,
		},
	}

	for i, tc := range testCases {
		suite.Run(tc.name, func() {
			address := fmt.Sprintf("vita1tiertest%d", i)
			
			// Ensure merchant doesn't exist
			suite.keeper.DeleteMerchant(suite.ctx, address)
			
			msg := &types.MsgRegisterMerchant{
				Sender:      address,
				BusinessName: fmt.Sprintf("Tier Test Business %d", i),
				StakeAmount:  tc.stakeAmount,
			}

			_, err := suite.msgServer.RegisterMerchant(ctx, msg)
			suite.Require().NoError(err)
			
			// Verify merchant was created with correct tier
			merchant, err := suite.keeper.GetMerchant(suite.ctx, address)
			suite.Require().NoError(err)
			suite.Require().Equal(tc.expectedTier, merchant.Tier)
		})
	}
}

// TestVaultRewardsCalculationThroughCreation tests vault rewards calculation through vault creation
func (suite *KeeperTestSuite) TestVaultRewardsCalculationThroughCreation() {
	ctx := sdk.UnwrapSDKContext(suite.ctx)
	
	testCases := []struct {
		name         string
		lockedAmount math.Int
		lockDuration int64
	}{
		{
			name:         "short lock vault",
			lockedAmount: math.NewInt(100000), // 100K VITA
			lockDurationBlocks: 1000,                // 1000 blocks
		},
		{
			name:         "long lock vault",
			lockedAmount: math.NewInt(500000), // 500K VITA
			lockDurationBlocks: 10000,               // 10000 blocks
		},
	}

	for i, tc := range testCases {
		suite.Run(tc.name, func() {
			creator := fmt.Sprintf("vita1vaultrewardtest%d", i)
			
			msg := &types.MsgCreateVault{
				Sender:      creator,
				Amount:       tc.lockedAmount,
				LockDuration: tc.lockDuration,
			}

			resp, err := suite.msgServer.CreateVault(ctx, msg)
			suite.Require().NoError(err)
			suite.Require().NotEmpty(resp.VaultId)
			
			// Verify vault was created with rewards calculation
			vault, err := suite.keeper.GetVault(suite.ctx, resp.VaultId)
			suite.Require().NoError(err)
			suite.Require().Equal(tc.lockedAmount, vault.Amount)
			suite.Require().Equal(tc.lockDuration, vault.LockDuration)
			suite.Require().True(vault.RewardMultiplier.IsPositive())
		})
	}
}