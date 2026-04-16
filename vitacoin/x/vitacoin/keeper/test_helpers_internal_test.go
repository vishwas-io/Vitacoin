package keeper

// test_helpers_internal_test.go — provides test setup for white-box (package keeper) tests.
// This file is compiled only during testing.

import (
	"context"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"

	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin/types"
)

// setupTestKeeper creates a fresh Keeper + Context for internal (white-box) tests.
func setupTestKeeper(t *testing.T) (Keeper, sdk.Context) {
	t.Helper()

	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("vita", "vitapub")
	config.SetBech32PrefixForValidator("vitavaloper", "vitavaloperpub")
	config.SetBech32PrefixForConsensusNode("vitavalcons", "vitavalconspub")

	ir := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(ir)
	cdc := codec.NewProtoCodec(ir)

	db := dbm.NewMemDB()
	ss := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	sk := storetypes.NewKVStoreKey(types.StoreKey)
	ss.MountStoreWithDB(sk, storetypes.StoreTypeIAVL, db)
	require.NoError(t, ss.LoadLatestVersion())

	ctx := sdk.NewContext(ss, cmtproto.Header{
		Height: 1,
		Time:   time.Unix(1_700_000_000, 0),
	}, false, log.NewNopLogger())

	bk := &internalMockBankKeeper{}
	ak := &internalMockAccountKeeper{}

	k := NewKeeper(
		cdc,
		runtime.NewKVStoreService(sk),
		log.NewNopLogger(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		bk,
		ak,
	)

	p := types.DefaultParams()
	p.MinMerchantStake = math.NewInt(1_000)
	p.MerchantRegistrationFee = math.NewInt(0)
	p.TransactionFeePercent = math.LegacyNewDecWithPrec(1, 1)
	p.MinProtocolFee = math.NewInt(0)
	p.MaxProtocolFee = math.NewInt(1_000_000_000_000_000_000)
	require.NoError(t, k.SetParams(ctx, p))

	return k, ctx
}

// internalMockBankKeeper fully implements types.BankKeeper.
type internalMockBankKeeper struct{}

func (m *internalMockBankKeeper) GetBalance(_ context.Context, _ sdk.AccAddress, denom string) sdk.Coin {
	return sdk.NewCoin(denom, math.NewInt(1_000_000_000))
}
func (m *internalMockBankKeeper) GetAllBalances(_ context.Context, _ sdk.AccAddress) sdk.Coins {
	return sdk.NewCoins(sdk.NewCoin(types.BondDenom, math.NewInt(1_000_000_000)))
}
func (m *internalMockBankKeeper) GetSupply(_ context.Context, denom string) sdk.Coin {
	return sdk.NewCoin(denom, math.NewInt(1_000_000_000_000_000))
}
func (m *internalMockBankKeeper) SendCoins(_ context.Context, _ sdk.AccAddress, _ sdk.AccAddress, _ sdk.Coins) error {
	return nil
}
func (m *internalMockBankKeeper) SendCoinsFromAccountToModule(_ context.Context, _ sdk.AccAddress, _ string, _ sdk.Coins) error {
	return nil
}
func (m *internalMockBankKeeper) SendCoinsFromModuleToAccount(_ context.Context, _ string, _ sdk.AccAddress, _ sdk.Coins) error {
	return nil
}
func (m *internalMockBankKeeper) SendCoinsFromModuleToModule(_ context.Context, _ string, _ string, _ sdk.Coins) error {
	return nil
}
func (m *internalMockBankKeeper) MintCoins(_ context.Context, _ string, _ sdk.Coins) error {
	return nil
}
func (m *internalMockBankKeeper) BurnCoins(_ context.Context, _ string, _ sdk.Coins) error {
	return nil
}
func (m *internalMockBankKeeper) SpendableCoins(_ context.Context, _ sdk.AccAddress) sdk.Coins {
	return sdk.NewCoins(sdk.NewCoin(types.BondDenom, math.NewInt(1_000_000_000)))
}

// internalMockAccountKeeper fully implements types.AccountKeeper.
type internalMockAccountKeeper struct{}

func (m *internalMockAccountKeeper) GetAccount(_ context.Context, _ sdk.AccAddress) sdk.AccountI {
	return nil
}
func (m *internalMockAccountKeeper) GetModuleAddress(name string) sdk.AccAddress {
	return authtypes.NewModuleAddress(name)
}
func (m *internalMockAccountKeeper) GetModuleAccount(_ context.Context, _ string) sdk.ModuleAccountI {
	return nil
}
