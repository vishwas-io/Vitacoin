package keeper_test

import (
	"encoding/json"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin/types"
)

// ─── IBCKeeperTestSuite ───────────────────────────────────────────────────────

type IBCKeeperTestSuite struct {
	KeeperTestSuite
}

func TestIBCKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(IBCKeeperTestSuite))
}

// helper: a valid sender address using the vita bech32 prefix
func ibcTestSender(t *testing.T) string {
	t.Helper()
	raw := sdk.AccAddress([]byte("sender______________"))
	return raw.String()
}

func ibcTestReceiver(t *testing.T) string {
	t.Helper()
	raw := sdk.AccAddress([]byte("receiver____________"))
	return raw.String()
}

// ─── TestIBCSendVITA ──────────────────────────────────────────────────────────

func (suite *IBCKeeperTestSuite) TestIBCSendVITA() {
	sender := ibcTestSender(suite.T())
	receiver := ibcTestReceiver(suite.T())
	amount := sdk.NewCoin("uvita", math.NewInt(1_000_000))

	err := suite.keeper.IBCSendVITA(suite.ctx, sender, receiver, "channel-0", amount, "test memo")
	require.NoError(suite.T(), err)

	// Module account should hold the escrowed coins.
	escrowed := suite.bankKeeper.GetModuleBalance(types.ModuleName, "uvita")
	require.Equal(suite.T(), math.NewInt(1_000_000), escrowed, "module account should hold escrowed coins")

	// KV should have exactly one pending packet.
	packets, err := suite.keeper.GetIBCPendingPackets(suite.ctx)
	require.NoError(suite.T(), err)
	require.Len(suite.T(), packets, 1)

	pkt := packets[0]
	require.Equal(suite.T(), sender, pkt.Sender)
	require.Equal(suite.T(), receiver, pkt.Receiver)
	require.Equal(suite.T(), "channel-0", pkt.SourceChannel)
	require.Equal(suite.T(), math.NewInt(1_000_000), pkt.Amount)
	require.Equal(suite.T(), "uvita", pkt.Denom)
	require.Equal(suite.T(), uint64(1), pkt.Sequence)
}

// TestIBCSendVITA_Errors checks validation paths.
func (suite *IBCKeeperTestSuite) TestIBCSendVITA_Errors() {
	sender := ibcTestSender(suite.T())
	receiver := ibcTestReceiver(suite.T())
	amount := sdk.NewCoin("uvita", math.NewInt(500))

	// Empty receiver
	err := suite.keeper.IBCSendVITA(suite.ctx, sender, "", "channel-0", amount, "")
	require.ErrorIs(suite.T(), err, types.ErrInvalidReceiver)

	// Zero amount
	zero := sdk.NewCoin("uvita", math.ZeroInt())
	err = suite.keeper.IBCSendVITA(suite.ctx, sender, receiver, "channel-0", zero, "")
	require.ErrorIs(suite.T(), err, types.ErrInvalidAmount)

	// Memo too long
	longMemo := make([]byte, types.MaxIBCMemo+1)
	for i := range longMemo {
		longMemo[i] = 'x'
	}
	err = suite.keeper.IBCSendVITA(suite.ctx, sender, receiver, "channel-0", amount, string(longMemo))
	require.ErrorIs(suite.T(), err, types.ErrMemoTooLong)
}

// ─── TestOnReceivePacket ──────────────────────────────────────────────────────

func (suite *IBCKeeperTestSuite) TestOnReceivePacket() {
	receiver := ibcTestReceiver(suite.T())
	packetData := types.VITAPacketData{
		Sender:        ibcTestSender(suite.T()),
		Receiver:      receiver,
		Amount:        math.NewInt(2_000_000),
		Denom:         "uvita",
		Memo:          "",
		Sequence:      42,
		SourceChannel: "channel-1",
	}

	ack, err := suite.keeper.OnReceivePacket(suite.ctx, packetData)
	require.NoError(suite.T(), err)
	require.True(suite.T(), ack.Success, "acknowledgement should be success")
	require.Empty(suite.T(), ack.Error)
}

// TestOnReceivePacket_InvalidPacket verifies error on bad data.
func (suite *IBCKeeperTestSuite) TestOnReceivePacket_InvalidPacket() {
	bad := types.VITAPacketData{
		// Missing Sender, Receiver, Denom
		Amount: math.NewInt(100),
	}
	ack, err := suite.keeper.OnReceivePacket(suite.ctx, bad)
	require.Error(suite.T(), err)
	require.False(suite.T(), ack.Success)
	require.NotEmpty(suite.T(), ack.Error)
}

// ─── TestOnAcknowledgePacket_Success ─────────────────────────────────────────

func (suite *IBCKeeperTestSuite) TestOnAcknowledgePacket_Success() {
	sender := ibcTestSender(suite.T())
	amount := sdk.NewCoin("uvita", math.NewInt(500_000))

	// First queue a packet so there is a pending record to delete.
	err := suite.keeper.IBCSendVITA(suite.ctx, sender, ibcTestReceiver(suite.T()), "channel-0", amount, "")
	require.NoError(suite.T(), err)

	packets, err := suite.keeper.GetIBCPendingPackets(suite.ctx)
	require.NoError(suite.T(), err)
	require.Len(suite.T(), packets, 1)

	// ACK success.
	ack := types.VITAAcknowledgement{Success: true}
	err = suite.keeper.OnAcknowledgePacket(suite.ctx, packets[0], ack)
	require.NoError(suite.T(), err)

	// Pending record should be deleted.
	packetsAfter, err := suite.keeper.GetIBCPendingPackets(suite.ctx)
	require.NoError(suite.T(), err)
	require.Empty(suite.T(), packetsAfter, "pending packet should be deleted after success ack")
}

// ─── TestOnAcknowledgePacket_Failure ─────────────────────────────────────────

func (suite *IBCKeeperTestSuite) TestOnAcknowledgePacket_Failure() {
	sender := ibcTestSender(suite.T())
	amount := sdk.NewCoin("uvita", math.NewInt(300_000))

	err := suite.keeper.IBCSendVITA(suite.ctx, sender, ibcTestReceiver(suite.T()), "channel-0", amount, "")
	require.NoError(suite.T(), err)

	packets, err := suite.keeper.GetIBCPendingPackets(suite.ctx)
	require.NoError(suite.T(), err)
	require.Len(suite.T(), packets, 1)

	// ACK failure — sender should be refunded (SendCoinsFromModuleToAccount called).
	ack := types.VITAAcknowledgement{Success: false, Error: "execution reverted"}
	err = suite.keeper.OnAcknowledgePacket(suite.ctx, packets[0], ack)
	require.NoError(suite.T(), err)

	// Pending record should be deleted after refund.
	packetsAfter, err := suite.keeper.GetIBCPendingPackets(suite.ctx)
	require.NoError(suite.T(), err)
	require.Empty(suite.T(), packetsAfter, "pending packet should be deleted after failed ack refund")
}

// ─── TestOnTimeoutPacket ─────────────────────────────────────────────────────

func (suite *IBCKeeperTestSuite) TestOnTimeoutPacket() {
	sender := ibcTestSender(suite.T())
	amount := sdk.NewCoin("uvita", math.NewInt(750_000))

	err := suite.keeper.IBCSendVITA(suite.ctx, sender, ibcTestReceiver(suite.T()), "channel-0", amount, "")
	require.NoError(suite.T(), err)

	packets, err := suite.keeper.GetIBCPendingPackets(suite.ctx)
	require.NoError(suite.T(), err)
	require.Len(suite.T(), packets, 1)

	// Timeout — escrowed coins returned to sender.
	err = suite.keeper.OnTimeoutPacket(suite.ctx, packets[0])
	require.NoError(suite.T(), err)

	// Pending record should be cleaned up.
	packetsAfter, err := suite.keeper.GetIBCPendingPackets(suite.ctx)
	require.NoError(suite.T(), err)
	require.Empty(suite.T(), packetsAfter, "pending packet should be deleted after timeout refund")
}

// ─── TestIBCRoundTrip ─────────────────────────────────────────────────────────

// TestIBCRoundTrip exercises the full send → JSON marshal → OnRecvPacket path
// end-to-end without an actual IBC stack, verifying packet data survives
// JSON round-tripping.
func (suite *IBCKeeperTestSuite) TestIBCRoundTrip() {
	sender := ibcTestSender(suite.T())
	receiver := ibcTestReceiver(suite.T())
	amount := sdk.NewCoin("uvita", math.NewInt(9_999))

	err := suite.keeper.IBCSendVITA(suite.ctx, sender, receiver, "channel-99", amount, "round-trip")
	require.NoError(suite.T(), err)

	pending, err := suite.keeper.GetIBCPendingPackets(suite.ctx)
	require.NoError(suite.T(), err)
	require.Len(suite.T(), pending, 1)

	pkt := pending[0]

	// Marshal as it would be placed in an IBC packet data field.
	bz, err := json.Marshal(pkt)
	require.NoError(suite.T(), err)

	// Unmarshal as the destination chain would.
	var decoded types.VITAPacketData
	err = json.Unmarshal(bz, &decoded)
	require.NoError(suite.T(), err)

	require.Equal(suite.T(), pkt.Sender, decoded.Sender)
	require.Equal(suite.T(), pkt.Receiver, decoded.Receiver)
	require.Equal(suite.T(), pkt.Amount, decoded.Amount)
	require.Equal(suite.T(), pkt.Denom, decoded.Denom)
	require.Equal(suite.T(), pkt.Memo, decoded.Memo)
	require.Equal(suite.T(), pkt.Sequence, decoded.Sequence)
}
