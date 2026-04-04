package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// UpdateParams implements the Msg/UpdateParams method
// Updates module parameters, can only be called by governance
func (ms msgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	// Validate authority
	if err := ms.Keeper.ValidateAuthority(msg.Authority); err != nil {
		return nil, err
	}

	// Validate params
	if err := msg.Params.Validate(); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	// Set params
	if err := ms.Keeper.SetParams(ctx, msg.Params); err != nil {
		return nil, fmt.Errorf("failed to set params: %w", err)
	}

	ms.Keeper.Logger().Info("params updated by governance",
		"authority", msg.Authority,
	)

	return &types.MsgUpdateParamsResponse{}, nil
}

// RegisterMerchant implements the Msg/RegisterMerchant method
// Registers a new merchant after collecting registration fee and initial stake
func (ms msgServer) RegisterMerchant(ctx context.Context, msg *types.MsgRegisterMerchant) (*types.MsgRegisterMerchantResponse, error) {
	// Validate sender address
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, fmt.Errorf("invalid sender address: %w", err)
	}

	// Check if merchant already exists
	exists, err := ms.Keeper.HasMerchant(ctx, msg.Sender)
	if err != nil {
		return nil, fmt.Errorf("failed to check merchant existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("merchant already registered: %s", msg.Sender)
	}

	// Validate business name
	if msg.BusinessName == "" {
		return nil, fmt.Errorf("business name cannot be empty")
	}
	if len(msg.BusinessName) > 100 {
		return nil, fmt.Errorf("invalid business name: too long (max 100 chars)")
	}
	// Get params
	params, err := ms.Keeper.GetParams(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get params: %w", err)
	}

	// Validate stake amount meets minimum
	if msg.StakeAmount.LT(params.MinMerchantStake) {
		return nil, fmt.Errorf("stake amount %s is less than minimum required %s", 
			msg.StakeAmount.String(), params.MinMerchantStake.String())
	}

	// Phase 3: Collect registration fee from sender (if non-zero)
	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, fmt.Errorf("invalid sender address for fee collection: %w", err)
	}

	if !params.MerchantRegistrationFee.IsZero() {
		feeCoins := sdk.NewCoins(sdk.NewCoin("avita", params.MerchantRegistrationFee))
		if err := ms.Keeper.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, types.ModuleName, feeCoins); err != nil {
			return nil, fmt.Errorf("failed to collect merchant registration fee: %w", err)
		}
	}

	// Phase 3: Collect initial stake from sender
	if !msg.StakeAmount.IsZero() {
		stakeCoins := sdk.NewCoins(sdk.NewCoin("avita", msg.StakeAmount))
		if err := ms.Keeper.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, types.ModuleName, stakeCoins); err != nil {
			return nil, fmt.Errorf("failed to collect merchant stake: %w", err)
		}
	}

	// Calculate merchant tier based on stake amount
	tier := ms.Keeper.calculateMerchantTier(msg.StakeAmount)

	// Create merchant
	merchant := types.Merchant{
		Address:            msg.Sender,
		BusinessName:       msg.BusinessName,
		Tier:               tier, // Automatically calculated from stake
		StakeAmount:        msg.StakeAmount,
		RegistrationHeight: sdk.UnwrapSDKContext(ctx).BlockHeight(),
		IsActive:           true,
		TotalTransactions:  0,
		TotalVolume:        math.ZeroInt(),
		// TODO: Add timestamp fields to Merchant struct
		// RegistrationTime: sdk.UnwrapSDKContext(ctx).BlockTime().Unix(),
		// LastActivityTime: sdk.UnwrapSDKContext(ctx).BlockTime().Unix(),
	}

	// Store merchant
	if err := ms.Keeper.SetMerchant(ctx, merchant); err != nil {
		return nil, fmt.Errorf("failed to store merchant: %w", err)
	}

	ms.Keeper.Logger().Info("merchant registered",
		"address", msg.Sender,
		"business_name", msg.BusinessName,
		"stake", msg.StakeAmount.String(),
	)

	return &types.MsgRegisterMerchantResponse{
		MerchantId: msg.Sender,
	}, nil
}

// UpdateMerchant implements the Msg/UpdateMerchant method
// Updates an existing merchant's information
func (ms msgServer) UpdateMerchant(ctx context.Context, msg *types.MsgUpdateMerchant) (*types.MsgUpdateMerchantResponse, error) {
	// Validate sender address
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, fmt.Errorf("invalid sender address: %w", err)
	}

	// Get existing merchant
	merchant, err := ms.Keeper.GetMerchant(ctx, msg.Sender)
	if err != nil {
		return nil, fmt.Errorf("merchant not found: %w", err)
	}

	// Update business name if provided (empty means no change, but explicitly validate if given)
	if msg.BusinessName != "" {
		if len(msg.BusinessName) > 100 {
			return nil, fmt.Errorf("invalid business name: too long (max 100 chars)")
		}
		merchant.BusinessName = msg.BusinessName
	}
	// Note: empty BusinessName means "no change" (not an error)

	// Add additional stake if provided
	if !msg.AdditionalStake.IsZero() {
		if msg.AdditionalStake.IsNegative() {
			return nil, fmt.Errorf("additional stake cannot be negative")
		}
		params, err := ms.Keeper.GetParams(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get params: %w", err)
		}
		newStake := merchant.StakeAmount.Add(msg.AdditionalStake)
		if newStake.LT(params.MinMerchantStake) {
			return nil, fmt.Errorf("insufficient stake amount: total stake %s is less than minimum required %s",
				newStake, params.MinMerchantStake)
		}
		// Phase 3: Collect additional stake from sender
		senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
		if err != nil {
			return nil, fmt.Errorf("invalid sender address for stake collection: %w", err)
		}
		stakeCoins := sdk.NewCoins(sdk.NewCoin("avita", msg.AdditionalStake))
		if err := ms.Keeper.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, types.ModuleName, stakeCoins); err != nil {
			return nil, fmt.Errorf("failed to collect additional stake: %w", err)
		}
		merchant.StakeAmount = newStake
	}
	
	// Always recalculate tier based on current stake amount (whether updated or not)
	merchant.Tier = ms.Keeper.calculateMerchantTier(merchant.StakeAmount)

	// Update last activity time
	// TODO: Add LastActivityTime field to Merchant struct
	// merchant.LastActivityTime = sdk.UnwrapSDKContext(ctx).BlockTime().Unix()

	// Store updated merchant
	if err := ms.Keeper.SetMerchant(ctx, merchant); err != nil {
		return nil, fmt.Errorf("failed to update merchant: %w", err)
	}

	ms.Keeper.Logger().Info("merchant updated",
		"address", msg.Sender,
		"new_stake", merchant.StakeAmount.String(),
		"tier", merchant.Tier.String(),
	)

	return &types.MsgUpdateMerchantResponse{}, nil
}

// CreatePayment implements the Msg/CreatePayment method
// Creates a new payment transaction
func (ms msgServer) CreatePayment(ctx context.Context, msg *types.MsgCreatePayment) (*types.MsgCreatePaymentResponse, error) {
	// Validate addresses
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, fmt.Errorf("invalid sender address: %w", err)
	}

	_, err = sdk.AccAddressFromBech32(msg.MerchantAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid merchant address: %w", err)
	}

	// Verify merchant exists and is active
	merchant, err := ms.Keeper.GetMerchant(ctx, msg.MerchantAddress)
	if err != nil {
		return nil, fmt.Errorf("merchant not found: %w", err)
	}
	if !merchant.IsActive {
		return nil, fmt.Errorf("merchant is not active")
	}

	// Validate amount
	if msg.Amount.IsZero() || msg.Amount.IsNegative() {
		return nil, fmt.Errorf("invalid payment amount: %s", msg.Amount.String())
	}

	// Check max transaction amount if set
	params, err := ms.Keeper.GetParams(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get params: %w", err)
	}
	if !params.MaxTransactionAmount.IsZero() && msg.Amount.GT(params.MaxTransactionAmount) {
		return nil, fmt.Errorf("amount %s exceeds max transaction amount %s", 
			msg.Amount.String(), params.MaxTransactionAmount.String())
	}

	// Generate payment ID (using block height + sender + nonce for uniqueness)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	paymentID := fmt.Sprintf("payment-%d-%s-%d", 
		sdkCtx.BlockHeight(), 
		msg.Sender, 
		sdkCtx.BlockTime().Unix())

	// Create payment
	payment := types.Payment{
		Id:               paymentID,
		FromAddress:      msg.Sender,
		ToAddress:        msg.MerchantAddress,
		Amount:           msg.Amount,
		// TODO: Add Fee field to Payment struct
		// Fee:             ms.calculateFee(msg.Amount, params.TransactionFeePercent),
		Status:           types.PaymentStatusPending,
		Memo:             msg.Memo,
		CreationHeight:   sdkCtx.BlockHeight(),
		// TODO: Add timestamp fields to Payment struct
		// CreatedAt:       sdkCtx.BlockTime().Unix(),
		// UpdatedAt:       sdkCtx.BlockTime().Unix(),
		// ExpiresAt:       sdkCtx.BlockHeight() + int64(params.PaymentTimeoutBlocks),
	}

	// Store payment
	if err := ms.Keeper.SetPayment(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to store payment: %w", err)
	}

	// Phase 3: Escrow payment amount from sender to module account
	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, fmt.Errorf("invalid sender address for escrow: %w", err)
	}
	
	if err := ms.Keeper.EscrowPaymentFunds(ctx, senderAddr, msg.Amount); err != nil {
		return nil, fmt.Errorf("failed to escrow payment funds: %w", err)
	}

	// Emit payment created event
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypePaymentCreated,
			sdk.NewAttribute(types.AttributeKeyPaymentID, paymentID),
			sdk.NewAttribute(types.AttributeKeyPayer, msg.Sender),
			sdk.NewAttribute(types.AttributeKeyMerchant, msg.MerchantAddress),
			sdk.NewAttribute(types.AttributeKeyAmount, msg.Amount.String()),
		),
	)

	ms.Keeper.Logger().Info("payment created and escrowed",
		"id", paymentID,
		"payer", msg.Sender,
		"merchant", msg.MerchantAddress,
		"amount", msg.Amount.String(),
	)

	return &types.MsgCreatePaymentResponse{
		PaymentId: paymentID,
	}, nil
}

// CompletePayment implements the Msg/CompletePayment method
// Completes a pending payment and releases funds to merchant
func (ms msgServer) CompletePayment(ctx context.Context, msg *types.MsgCompletePayment) (*types.MsgCompletePaymentResponse, error) {
	// Validate sender
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, fmt.Errorf("invalid sender address: %w", err)
	}

	// Get payment
	payment, err := ms.Keeper.GetPayment(ctx, msg.PaymentId)
	if err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}

	// Verify sender is the merchant
	if payment.ToAddress != msg.Sender {
		return nil, fmt.Errorf("only merchant can complete payment")
	}

	// Check payment status
	if payment.Status != types.PaymentStatusPending {
		return nil, fmt.Errorf("payment is not pending, current status: %s", payment.Status.String())
	}

	// TODO: Check if payment has expired when ExpiresAt field is added
	// sdkCtx := sdk.UnwrapSDKContext(ctx)
	// if sdkCtx.BlockHeight() > payment.ExpiresAt {
	//	return nil, fmt.Errorf("payment has expired at block %d", payment.ExpiresAt)
	// }

	// Update payment status
	payment.Status = types.PaymentStatusCompleted
	payment.CompletionHeight = sdk.UnwrapSDKContext(ctx).BlockHeight()
	// TODO: Add UpdatedAt and CompletedAt fields when available
	// payment.UpdatedAt = sdkCtx.BlockTime().Unix()
	// payment.CompletedAt = sdkCtx.BlockTime().Unix()

	// Store updated payment
	if err := ms.Keeper.SetPayment(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	// Phase 3: Release escrowed funds to merchant (minus protocol fees)
	merchantAddr, err := sdk.AccAddressFromBech32(payment.ToAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid merchant address: %w", err)
	}
	
	feeAmount, netAmount, err := ms.Keeper.ReleasePaymentFunds(ctx, merchantAddr, payment.Amount, msg.PaymentId)
	if err != nil {
		return nil, fmt.Errorf("failed to release payment funds: %w", err)
	}

	// Update merchant stats with gross amount (before fees)
	merchant, err := ms.Keeper.GetMerchant(ctx, payment.ToAddress)
	if err == nil {
		merchant.TotalTransactions++
		merchant.TotalVolume = merchant.TotalVolume.Add(payment.Amount) // Track gross volume
		// TODO: Add LastActivityTime field to Merchant struct
		// merchant.LastActivityTime = sdkCtx.BlockTime().Unix()
		ms.Keeper.SetMerchant(ctx, merchant)
	}

	ms.Keeper.Logger().Info("payment completed with fees",
		"id", msg.PaymentId,
		"merchant", msg.Sender,
		"gross_amount", payment.Amount.String(),
		"protocol_fee", feeAmount.String(),
		"net_amount", netAmount.String(),
	)

	return &types.MsgCompletePaymentResponse{}, nil
}

// RefundPayment implements the Msg/RefundPayment method
// Refunds a completed payment back to the payer
func (ms msgServer) RefundPayment(ctx context.Context, msg *types.MsgRefundPayment) (*types.MsgRefundPaymentResponse, error) {
	// Validate sender
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, fmt.Errorf("invalid sender address: %w", err)
	}

	// Get payment
	payment, err := ms.Keeper.GetPayment(ctx, msg.PaymentId)
	if err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}

	// Verify sender is the merchant
	if payment.ToAddress != msg.Sender {
		return nil, fmt.Errorf("only the merchant can refund payments")
	}

	// Check payment status (can only refund completed payments)
	if payment.Status != types.PaymentStatusCompleted {
		return nil, fmt.Errorf("can only refund completed payments, current status: %s", payment.Status.String())
	}

	// Validate refund reason
	if msg.Reason == "" {
		return nil, fmt.Errorf("refund reason is required")
	}

	// Update payment status
	payment.Status = types.PaymentStatusRefunded
	// TODO: Add UpdatedAt field to Payment struct
	// TODO: Add RefundReason field to Payment struct

	// Store updated payment
	if err := ms.Keeper.SetPayment(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	// Update merchant stats (decrease counts)
	merchant, err := ms.Keeper.GetMerchant(ctx, payment.ToAddress)
	if err == nil {
		if merchant.TotalTransactions > 0 {
			merchant.TotalTransactions--
		}
		merchant.TotalVolume = merchant.TotalVolume.Sub(payment.Amount)
		if merchant.TotalVolume.IsNegative() {
			merchant.TotalVolume = math.ZeroInt()
		}
		// TODO: Add LastActivityTime field to Merchant struct
		// merchant.LastActivityTime = sdkCtx.BlockTime().Unix()
		ms.Keeper.SetMerchant(ctx, merchant)
	}

	// Phase 3: NOTE - Refund implementation
	// For completed payments, funds have already been settled (merchant received net, fees distributed)
	// Refunds should be a separate merchant-to-customer transfer, not unwinding the original payment
	// Merchant needs to have sufficient balance to refund the customer
	// This ensures protocol fees are not reversed (fees were earned for processing the original payment)
	
	// Future enhancement: Implement actual refund transfer
	// merchantAddr, _ := sdk.AccAddressFromBech32(payment.ToAddress)
	// payerAddr, _ := sdk.AccAddressFromBech32(payment.FromAddress)
	// refundCoins := sdk.NewCoins(sdk.NewCoin("avita", payment.Amount))
	// err = ms.Keeper.bankKeeper.SendCoins(ctx, merchantAddr, payerAddr, refundCoins)

	ms.Keeper.Logger().Info("payment marked as refunded",
		"id", msg.PaymentId,
		"merchant", msg.Sender,
		"reason", msg.Reason,
		"note", "refund transfer should be processed by merchant directly",
	)

	return &types.MsgRefundPaymentResponse{}, nil
}

// CreateVault implements the Msg/CreateVault method
// Creates a time-locked vault for staking
func (ms msgServer) CreateVault(ctx context.Context, msg *types.MsgCreateVault) (*types.MsgCreateVaultResponse, error) {
	// Validate sender
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, fmt.Errorf("invalid sender address: %w", err)
	}

	// Validate amount
	if msg.Amount.IsZero() || msg.Amount.IsNegative() {
		return nil, fmt.Errorf("invalid vault amount: %s", msg.Amount.String())
	}

	// Validate lock duration
	if msg.LockDuration == 0 {
		return nil, fmt.Errorf("lock duration must be positive")
	}

	// Generate vault ID
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	vaultID := fmt.Sprintf("vault-%d-%s-%d", 
		sdkCtx.BlockHeight(), 
		msg.Sender, 
		sdkCtx.BlockTime().Unix())

	unlockHeight := sdkCtx.BlockHeight() + int64(msg.LockDuration)

	// Create vault
	vault := types.Vault{
		Id:               vaultID,
		Owner:            msg.Sender,
		Amount:           msg.Amount,
		LockDuration:     msg.LockDuration,
		CreationHeight:   sdkCtx.BlockHeight(),
		UnlockHeight:     unlockHeight,
		RewardMultiplier: ms.Keeper.calculateRewardMultiplier(msg.LockDuration),
		// TODO: Regenerate proto to include IsActive field for Vault
		// IsActive:         true, // Vault is active when created
		// TODO: Add fields for tracking withdrawal status and timestamps
		// IsWithdrawn:   false,
		// CreatedAt:     sdkCtx.BlockTime().Unix(),
	}

	// Store vault
	if err := ms.Keeper.SetVault(ctx, vault); err != nil {
		return nil, fmt.Errorf("failed to store vault: %w", err)
	}

	// TODO Phase 4: Lock tokens from sender

	ms.Keeper.Logger().Info("vault created",
		"id", vaultID,
		"owner", msg.Sender,
		"amount", msg.Amount.String(),
		"unlock_height", unlockHeight,
	)

	return &types.MsgCreateVaultResponse{
		VaultId:      vaultID,
		UnlockHeight: unlockHeight,
	}, nil
}

// WithdrawVault implements the Msg/WithdrawVault method
// Withdraws from an unlocked vault
func (ms msgServer) WithdrawVault(ctx context.Context, msg *types.MsgWithdrawVault) (*types.MsgWithdrawVaultResponse, error) {
	// Validate sender
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, fmt.Errorf("invalid sender address: %w", err)
	}

	// Get vault
	vault, err := ms.Keeper.GetVault(ctx, msg.VaultId)
	if err != nil {
		return nil, fmt.Errorf("vault not found: %w", err)
	}

	// Verify sender is the vault owner
	if vault.Owner != msg.Sender {
		return nil, fmt.Errorf("only vault owner can withdraw")
	}

	// TODO: Add IsActive field check when proto is regenerated
	// Check if vault is still active
	// if !vault.IsActive {
	//     return nil, fmt.Errorf("vault is not active")
	// }

	// Check if vault is unlocked
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if sdkCtx.BlockHeight() < vault.UnlockHeight {
		return nil, fmt.Errorf("vault is still locked until block %d (current: %d)", 
			vault.UnlockHeight, sdkCtx.BlockHeight())
	}

	// Calculate rewards (simplified - TODO: implement proper reward calculation in Phase 4)
	lockDuration := vault.UnlockHeight - vault.CreationHeight
	rewards := ms.Keeper.calculateVaultRewards(vault.Amount, lockDuration)
	
	// TODO: Deactivate vault after withdrawal when proto is regenerated
	// vault.IsActive = false

	// Store updated vault
	if err := ms.Keeper.SetVault(ctx, vault); err != nil {
		return nil, fmt.Errorf("failed to update vault: %w", err)
	}

	// TODO Phase 4: Transfer locked amount + rewards to owner

	ms.Keeper.Logger().Info("vault withdrawn",
		"id", msg.VaultId,
		"owner", msg.Sender,
		"amount", vault.Amount.String(),
		"rewards", rewards.String(),
	)

	return &types.MsgWithdrawVaultResponse{
		AmountWithdrawn: vault.Amount,
		RewardsEarned:   rewards,
	}, nil
}

// CreateRewardPool implements the Msg/CreateRewardPool method
// Creates a merchant reward pool for customer loyalty
func (ms msgServer) CreateRewardPool(ctx context.Context, msg *types.MsgCreateRewardPool) (*types.MsgCreateRewardPoolResponse, error) {
	// Validate sender
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, fmt.Errorf("invalid sender address: %w", err)
	}

	// Verify sender is a merchant
	merchant, err := ms.Keeper.GetMerchant(ctx, msg.Sender)
	if err != nil {
		return nil, fmt.Errorf("sender is not a registered merchant: %w", err)
	}
	if !merchant.IsActive {
		return nil, fmt.Errorf("merchant is not active")
	}

	// Validate total rewards
	if msg.TotalRewards.IsZero() || msg.TotalRewards.IsNegative() {
		return nil, fmt.Errorf("invalid total rewards: %s", msg.TotalRewards.String())
	}

	// Generate pool ID
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	poolID := fmt.Sprintf("pool-%d-%s-%d", 
		sdkCtx.BlockHeight(), 
		msg.Sender, 
		sdkCtx.BlockTime().Unix())

	// Calculate end height if duration is specified
	endHeight := int64(0) // 0 means no end
	if msg.DurationBlocks > 0 {
		endHeight = sdkCtx.BlockHeight() + int64(msg.DurationBlocks)
	}

	// Create reward pool
	pool := types.RewardPool{
		Id:                poolID,
		MerchantAddress:   msg.Sender,
		TotalRewards:      msg.TotalRewards,
		// TODO: Add RemainingRewards field to RewardPool struct
		// RemainingRewards:  msg.TotalRewards,
		DistributedRewards: math.ZeroInt(),
		StartHeight:       sdkCtx.BlockHeight(),
		EndHeight:         endHeight,
		IsActive:          true,
		// TODO: Add CreatedAt and ExpiresAt fields to RewardPool struct
		// CreatedAt:         sdkCtx.BlockTime().Unix(),
		// ExpiresAt:         expiresAt,
	}

	// Store pool
	if err := ms.Keeper.SetRewardPool(ctx, pool); err != nil {
		return nil, fmt.Errorf("failed to store reward pool: %w", err)
	}

	// TODO Phase 3: Lock reward tokens from merchant

	ms.Keeper.Logger().Info("reward pool created",
		"id", poolID,
		"merchant", msg.Sender,
		"total_rewards", msg.TotalRewards.String(),
	)

	return &types.MsgCreateRewardPoolResponse{
		PoolId: poolID,
	}, nil
}

// DistributeRewards implements the Msg/DistributeRewards method
// Distributes rewards to customers from a merchant's pool
func (ms msgServer) DistributeRewards(ctx context.Context, msg *types.MsgDistributeRewards) (*types.MsgDistributeRewardsResponse, error) {
	// Validate sender
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, fmt.Errorf("invalid sender address: %w", err)
	}

	// Get reward pool
	pool, err := ms.Keeper.GetRewardPool(ctx, msg.PoolId)
	if err != nil {
		return nil, fmt.Errorf("reward pool not found: %w", err)
	}

	// Verify sender is the pool merchant
	if pool.MerchantAddress != msg.Sender {
		return nil, fmt.Errorf("only pool merchant can distribute rewards")
	}

	// Check if pool is active
	if !pool.IsActive {
		return nil, fmt.Errorf("reward pool is not active")
	}

	// Check if pool has expired
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if pool.EndHeight > 0 && sdkCtx.BlockHeight() > pool.EndHeight {
		return nil, fmt.Errorf("reward pool has expired")
	}

	// Validate recipients and amounts match
	if len(msg.Recipients) == 0 {
		return nil, fmt.Errorf("no recipients provided")
	}
	if len(msg.Recipients) != len(msg.Amounts) {
		return nil, fmt.Errorf("recipients and amounts length mismatch")
	}

	// Calculate total distribution
	totalDistributed := math.ZeroInt()
	for i, amount := range msg.Amounts {
		// Validate recipient address
		if _, err := sdk.AccAddressFromBech32(msg.Recipients[i]); err != nil {
			return nil, fmt.Errorf("invalid recipient address at index %d: %w", i, err)
		}

		// Validate amount
		if amount.IsZero() || amount.IsNegative() {
			return nil, fmt.Errorf("invalid amount at index %d: %s", i, amount.String())
		}

		totalDistributed = totalDistributed.Add(amount)
	}

	// Check if enough rewards remain
	remainingRewards := pool.TotalRewards.Sub(pool.DistributedRewards)
	if totalDistributed.GT(remainingRewards) {
		return nil, fmt.Errorf("insufficient rewards: need %s, have %s", 
			totalDistributed.String(), remainingRewards.String())
	}

	// Update pool distributed rewards
	pool.DistributedRewards = pool.DistributedRewards.Add(totalDistributed)

	// Deactivate if all rewards distributed
	if pool.DistributedRewards.Equal(pool.TotalRewards) {
		pool.IsActive = false
	}

	// Store updated pool
	if err := ms.Keeper.SetRewardPool(ctx, pool); err != nil {
		return nil, fmt.Errorf("failed to update reward pool: %w", err)
	}

	// TODO Phase 3: Transfer rewards to recipients

	ms.Keeper.Logger().Info("rewards distributed",
		"pool_id", msg.PoolId,
		"merchant", msg.Sender,
		"recipients", len(msg.Recipients),
		"total_distributed", totalDistributed.String(),
	)

	return &types.MsgDistributeRewardsResponse{
		TotalDistributed: totalDistributed,
	}, nil
}

// Helper functions

// calculateFee calculates transaction fee based on amount and fee percentage
func (ms msgServer) calculateFee(amount math.Int, feePercent math.LegacyDec) math.Int {
	// Fee = amount * (feePercent / 100)
	amountDec := math.LegacyNewDecFromInt(amount)
	feeDec := amountDec.Mul(feePercent).Quo(math.LegacyNewDec(100))
	return feeDec.TruncateInt()
}

// calculateMerchantTier determines merchant tier based on stake amount
func (ms msgServer) calculateMerchantTier(stakeAmount math.Int) types.MerchantTier {
	// Simple tier calculation - can be made more sophisticated
	// These thresholds can be moved to params in Phase 3
	platinumThreshold, _ := math.NewIntFromString("1000000000000000000000000") // 1M tokens (1e24)
	goldThreshold, _ := math.NewIntFromString("100000000000000000000000")      // 100K tokens (1e23)
	silverThreshold, _ := math.NewIntFromString("10000000000000000000000")     // 10K tokens (1e22)
	
	if stakeAmount.GTE(platinumThreshold) {
		return types.MerchantTierPlatinum
	} else if stakeAmount.GTE(goldThreshold) {
		return types.MerchantTierGold
	} else if stakeAmount.GTE(silverThreshold) {
		return types.MerchantTierSilver
	}
	return types.MerchantTierBronze
}

// calculateVaultRewards calculates rewards for a vault based on locked amount and duration
func (ms msgServer) calculateVaultRewards(lockedAmount math.Int, lockDuration int64) math.Int {
	// Simple reward calculation: 0.1% per 10000 blocks (~7 days assuming 6s blocks)
	// This is simplified - actual implementation in Phase 4 will use proper APY calculations
	rewardRate := math.LegacyNewDecWithPrec(1, 3) // 0.1%
	periods := math.LegacyNewDec(lockDuration).Quo(math.LegacyNewDec(10000))
	
	amountDec := math.LegacyNewDecFromInt(lockedAmount)
	rewardDec := amountDec.Mul(rewardRate).Mul(periods)
	
	return rewardDec.TruncateInt()
}
