package keeper_test

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MockBankKeeper is a minimal mock for testing
type MockBankKeeper struct {
	// ModuleBalances tracks coins sent to module accounts: moduleName -> denom -> amount
	ModuleBalances map[string]map[string]math.Int
}

func NewMockBankKeeper() *MockBankKeeper {
	return &MockBankKeeper{
		ModuleBalances: make(map[string]map[string]math.Int),
	}
}

func (m *MockBankKeeper) GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return sdk.NewCoin(denom, math.ZeroInt())
}
func (m *MockBankKeeper) GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins {
	return sdk.NewCoins()
}
func (m *MockBankKeeper) GetSupply(ctx context.Context, denom string) sdk.Coin {
	return sdk.NewCoin(denom, math.ZeroInt())
}
func (m *MockBankKeeper) SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	return nil
}
func (m *MockBankKeeper) SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	if m.ModuleBalances == nil {
		m.ModuleBalances = make(map[string]map[string]math.Int)
	}
	if m.ModuleBalances[recipientModule] == nil {
		m.ModuleBalances[recipientModule] = make(map[string]math.Int)
	}
	for _, coin := range amt {
		existing, ok := m.ModuleBalances[recipientModule][coin.Denom]
		if !ok {
			existing = math.ZeroInt()
		}
		m.ModuleBalances[recipientModule][coin.Denom] = existing.Add(coin.Amount)
	}
	return nil
}
func (m *MockBankKeeper) SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	return nil
}
func (m *MockBankKeeper) SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error {
	return nil
}
func (m *MockBankKeeper) MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error {
	return nil
}
func (m *MockBankKeeper) BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error {
	return nil
}
func (m *MockBankKeeper) SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins {
	return sdk.NewCoins()
}

// GetModuleBalance returns the tracked balance for a module account and denom
func (m *MockBankKeeper) GetModuleBalance(moduleName, denom string) math.Int {
	if m.ModuleBalances == nil {
		return math.ZeroInt()
	}
	if m.ModuleBalances[moduleName] == nil {
		return math.ZeroInt()
	}
	bal, ok := m.ModuleBalances[moduleName][denom]
	if !ok {
		return math.ZeroInt()
	}
	return bal
}

// MockAccountKeeper is a minimal mock for testing
type MockAccountKeeper struct{}

func (m *MockAccountKeeper) GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI {
	return nil
}
func (m *MockAccountKeeper) GetModuleAddress(name string) sdk.AccAddress {
	return sdk.AccAddress{}
}
func (m *MockAccountKeeper) GetModuleAccount(ctx context.Context, name string) sdk.ModuleAccountI {
	return nil
}
