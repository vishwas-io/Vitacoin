package types_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/esspron/VITACOIN/vitacoin/vitacoin/x/vitacoin/types"
)

func BenchmarkMsgRegisterMerchant_ValidateBasic(b *testing.B) {
	msg := types.MsgRegisterMerchant{
		Sender:       validAddress1,
		BusinessName: "Test Business Name",
		StakeAmount:  math.NewInt(1e18),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = msg.ValidateBasic()
	}
}

func BenchmarkMsgCreatePayment_ValidateBasic(b *testing.B) {
	msg := types.MsgCreatePayment{
		Sender:          validAddress1,
		MerchantAddress: validAddress2,
		Amount:          math.NewInt(1e18),
		Memo:            "Test payment memo for benchmarking",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = msg.ValidateBasic()
	}
}

func BenchmarkMsgCreateVault_ValidateBasic(b *testing.B) {
	msg := types.MsgCreateVault{
		Sender:       validAddress1,
		Amount:       math.NewInt(1e18),
		LockDuration: 1000,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = msg.ValidateBasic()
	}
}

func BenchmarkMsgCreateRewardPool_ValidateBasic(b *testing.B) {
	msg := types.MsgCreateRewardPool{
		Sender:         validAddress1,
		TotalRewards:   math.NewInt(1e18),
		DurationBlocks: 1000,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = msg.ValidateBasic()
	}
}

func BenchmarkMsgDistributeRewards_ValidateBasic(b *testing.B) {
	recipients := make([]string, 100)
	amounts := make([]math.Int, 100)
	for i := 0; i < 100; i++ {
		recipients[i] = validAddress1
		amounts[i] = math.NewInt(1e16)
	}

	msg := types.MsgDistributeRewards{
		Sender:     validAddress1,
		PoolId:     "test-pool-123",
		Recipients: recipients,
		Amounts:    amounts,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = msg.ValidateBasic()
	}
}

func BenchmarkParams_Validate(b *testing.B) {
	params := types.DefaultParams()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = params.Validate()
	}
}

func BenchmarkMerchant_String(b *testing.B) {
	merchant := types.Merchant{
		Address:            validAddress1,
		BusinessName:       "Test Business Name",
		Tier:               types.MerchantTierGold,
		StakeAmount:        math.NewInt(1e18),
		RegistrationHeight: 100,
		IsActive:           true,
		TotalVolume:        math.NewInt(5e18),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = merchant.String()
	}
}

func BenchmarkPayment_String(b *testing.B) {
	payment := types.Payment{
		Id:               "payment-123",
		FromAddress:      validAddress1,
		ToAddress:        validAddress2,
		Amount:           math.NewInt(1e18),
		Status:           types.PaymentStatusCompleted,
		CreationHeight:   100,
		CompletionHeight: 150,
		Memo:             "Test payment memo for benchmarking",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = payment.String()
	}
}

func BenchmarkRewardPool_String(b *testing.B) {
	pool := types.RewardPool{
		Id:                 "pool-123",
		MerchantAddress:    validAddress1,
		TotalRewards:       math.NewInt(1e18),
		DistributedRewards: math.NewInt(5e17),
		StartHeight:        100,
		EndHeight:          1000,
		IsActive:           true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pool.String()
	}
}

func BenchmarkMsgRegisterMerchant_String(b *testing.B) {
	msg := &types.MsgRegisterMerchant{
		Sender:       validAddress1,
		BusinessName: "Test Business Name",
		StakeAmount:  math.NewInt(1e18),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = msg.String()
	}
}

func BenchmarkMsgDistributeRewards_String_LargeRecipientList(b *testing.B) {
	recipients := make([]string, 1000)
	amounts := make([]math.Int, 1000)
	for i := 0; i < 1000; i++ {
		recipients[i] = validAddress1
		amounts[i] = math.NewInt(1e16)
	}

	msg := &types.MsgDistributeRewards{
		Sender:     validAddress1,
		PoolId:     "test-pool-123",
		Recipients: recipients,
		Amounts:    amounts,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = msg.String()
	}
}