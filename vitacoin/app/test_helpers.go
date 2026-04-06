package app

import (
	"encoding/json"
	"math/rand"
	"testing"
	"time"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// Setup initializes a new VitacoinApp for testing purposes
func Setup(t *testing.T, isCheckTx bool) *VitacoinApp {
	t.Helper()

	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err)

	// Create validator set with single validator
	validator := cmttypes.NewValidator(pubKey, 1)
	valSet := cmttypes.NewValidatorSet([]*cmttypes.Validator{validator})

	// Generate genesis account
	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100000000000000))),
	}

	app := SetupWithGenesisValSet(t, valSet, []authtypes.GenesisAccount{acc}, balance)

	return app
}

// SetupWithGenesisValSet initializes a new VitacoinApp with a validator set and genesis accounts
func SetupWithGenesisValSet(t *testing.T, valSet *cmttypes.ValidatorSet, genAccs []authtypes.GenesisAccount, balances ...banktypes.Balance) *VitacoinApp {
	t.Helper()

	db := dbm.NewMemDB()
	logger := log.NewTestLogger(t)
	
	app := NewVitacoinApp(logger, db, nil, true, simtestutil.NewAppOptionsWithFlagHome(t.TempDir()))

	genesisState := NewDefaultGenesisState()

	// Set up auth genesis state
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
	genesisState[authtypes.ModuleName] = app.AppCodec().MustMarshalJSON(authGenesis)

	// Set up bank genesis state
	totalSupply := sdk.NewCoins()
	for _, b := range balances {
		totalSupply = totalSupply.Add(b.Coins...)
	}

	bankGenesis := banktypes.NewGenesisState(
		banktypes.DefaultGenesisState().Params,
		balances,
		totalSupply,
		[]banktypes.Metadata{},
		[]banktypes.SendEnabled{},
	)
	genesisState[banktypes.ModuleName] = app.AppCodec().MustMarshalJSON(bankGenesis)

	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	require.NoError(t, err)

	// Initialize the chain
	app.InitChain(
		&abci.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: simtestutil.DefaultConsensusParams,
			AppStateBytes:   stateBytes,
		},
	)

	// Commit genesis changes
	app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:             app.LastBlockHeight() + 1,
		Time:               time.Now(),
	})

	return app
}

// SetupWithGenesisAccounts initializes a new VitacoinApp with genesis accounts
func SetupWithGenesisAccounts(t *testing.T, genAccs []authtypes.GenesisAccount, balances ...banktypes.Balance) *VitacoinApp {
	t.Helper()

	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err)

	validator := cmttypes.NewValidator(pubKey, 1)
	valSet := cmttypes.NewValidatorSet([]*cmttypes.Validator{validator})

	return SetupWithGenesisValSet(t, valSet, genAccs, balances...)
}

// NewTestApp creates a new VitacoinApp instance for testing
func NewTestApp(t *testing.T) *VitacoinApp {
	t.Helper()

	db := dbm.NewMemDB()
	logger := log.NewTestLogger(t)
	
	return NewVitacoinApp(logger, db, nil, true, simtestutil.NewAppOptionsWithFlagHome(t.TempDir()))
}

// MakeTestEncodingConfig creates an EncodingConfig for testing
func MakeTestEncodingConfig() EncodingConfig {
	encodingConfig := MakeEncodingConfig()
	std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	ModuleBasics.RegisterLegacyAminoCodec(encodingConfig.Amino)
	ModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	return encodingConfig
}

// CreateTestContext creates a test context for unit testing
func CreateTestContext(app *VitacoinApp) sdk.Context {
	header := cmtproto.Header{
		Height: app.LastBlockHeight() + 1,
		Time:   time.Now(),
	}
	return app.BaseApp.NewContext(false).WithBlockHeader(header)
}

// CreateRandomAccounts creates a specified number of random test accounts
func CreateRandomAccounts(n int) []sdk.AccAddress {
	accounts := make([]sdk.AccAddress, n)
	for i := 0; i < n; i++ {
		pk := secp256k1.GenPrivKey().PubKey()
		accounts[i] = sdk.AccAddress(pk.Address())
	}
	return accounts
}

// FundAccount funds an account with the specified coins
func FundAccount(t *testing.T, app *VitacoinApp, ctx sdk.Context, addr sdk.AccAddress, amounts sdk.Coins) {
	t.Helper()

	err := app.BankKeeper.MintCoins(ctx, "mint", amounts)
	require.NoError(t, err)

	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, "mint", addr, amounts)
	require.NoError(t, err)
}

// GenesisStateWithValSet returns a new genesis state with the validator set
func GenesisStateWithValSet(codec codec.Codec, genesisState map[string]json.RawMessage,
	valSet *cmttypes.ValidatorSet, genAccs []authtypes.GenesisAccount,
	balances ...banktypes.Balance) map[string]json.RawMessage {
	
	// Set genesis accounts
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
	genesisState[authtypes.ModuleName] = codec.MustMarshalJSON(authGenesis)

	totalSupply := sdk.NewCoins()
	for _, b := range balances {
		// Add coins to total supply
		totalSupply = totalSupply.Add(b.Coins...)
	}

	// Set genesis bank state
	bankGenesis := banktypes.NewGenesisState(
		banktypes.DefaultGenesisState().Params,
		balances,
		totalSupply,
		[]banktypes.Metadata{},
		[]banktypes.SendEnabled{},
	)
	genesisState[banktypes.ModuleName] = codec.MustMarshalJSON(bankGenesis)

	return genesisState
}

// SignAndDeliver signs and delivers a transaction
func SignAndDeliver(
	t *testing.T, txCfg client.TxConfig, app *VitacoinApp, header cmtproto.Header, msgs []sdk.Msg,
	chainID string, accNums, accSeqs []uint64, expPass bool, priv ...cryptotypes.PrivKey,
) (sdk.GasInfo, *sdk.Result, error) {
	t.Helper()

	tx, err := simtestutil.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		txCfg,
		msgs,
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)},
		simtestutil.DefaultGenTxGas,
		chainID,
		accNums,
		accSeqs,
		priv...,
	)
	require.NoError(t, err)

	_, err = txCfg.TxEncoder()(tx)
	require.NoError(t, err)

	// Simulate a sending a transaction
	gInfo, res, err := app.SimDeliver(txCfg.TxEncoder(), tx)

	if expPass {
		require.NoError(t, err)
		require.NotNil(t, res)
	} else {
		require.Error(t, err)
		require.Nil(t, res)
	}

	return gInfo, res, err
}

// CheckBalance checks that an account has the expected balance
func CheckBalance(t *testing.T, app *VitacoinApp, ctx sdk.Context, addr sdk.AccAddress, expected sdk.Coins) {
	t.Helper()

	balances := app.BankKeeper.GetAllBalances(ctx, addr)
	require.True(t, balances.Equal(expected), "expected %s, got %s", expected, balances)
}

// GetAccountBalance returns the balance of an account
func GetAccountBalance(app *VitacoinApp, ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return app.BankKeeper.GetAllBalances(ctx, addr)
}
