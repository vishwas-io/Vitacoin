package vitacoin

// ibc_module.go — IBCModule interface implementation for AppModule.
//
// The vitacoin AppModule implements the types.IBCModule interface so it can be
// wired into an IBC router.  When ibc-go is added as a Go dependency, swap the
// local types.IBCModule interface for porttypes.IBCModule from ibc-go.

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin/types"
)

// compile-time assertion
var _ types.IBCModule = AppModule{}

// ─── Channel Lifecycle ────────────────────────────────────────────────────────

// OnChanOpenInit validates the proposed channel version during channel
// initialisation on the initiating chain.  VitaCoin only accepts IBCVersion.
func (am AppModule) OnChanOpenInit(
	order types.ChannelOrder,
	connectionHops []string,
	portID string,
	channelID string,
	counterpartyPortID string,
	counterpartyChannelID string,
	version string,
) (string, error) {
	if version != "" && version != types.IBCVersion {
		return "", fmt.Errorf("invalid IBC version %q; expected %q", version, types.IBCVersion)
	}
	return types.IBCVersion, nil
}

// OnChanOpenTry validates the counterparty version during channel handshake
// (the try step on the receiving chain).
func (am AppModule) OnChanOpenTry(
	order types.ChannelOrder,
	connectionHops []string,
	portID string,
	channelID string,
	counterpartyPortID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) (string, error) {
	if counterpartyVersion != types.IBCVersion {
		return "", fmt.Errorf("invalid counterparty IBC version %q; expected %q", counterpartyVersion, types.IBCVersion)
	}
	return types.IBCVersion, nil
}

// OnChanOpenAck is called when the initiating chain receives the channel ACK.
// We emit an event and accept the counterparty version.
func (am AppModule) OnChanOpenAck(portID, channelID, counterpartyChannelID, counterpartyVersion string) error {
	if counterpartyVersion != types.IBCVersion {
		return fmt.Errorf("invalid counterparty IBC version %q; expected %q", counterpartyVersion, types.IBCVersion)
	}
	return nil
}

// OnChanOpenConfirm is called on the receiving chain after ACK is relayed.
func (am AppModule) OnChanOpenConfirm(portID, channelID string) error {
	return nil
}

// OnChanCloseInit is called when the channel close is initiated locally.
// VitaCoin does not allow unilateral channel closure.
func (am AppModule) OnChanCloseInit(portID, channelID string) error {
	return fmt.Errorf("user cannot close channel")
}

// OnChanCloseConfirm is called when the channel closure is confirmed.
func (am AppModule) OnChanCloseConfirm(portID, channelID string) error {
	return nil
}

// ─── Packet Callbacks ─────────────────────────────────────────────────────────

// OnRecvPacket processes an inbound IBC packet on the destination chain.
// It unmarshals the VITAPacketData, calls the keeper handler, and returns
// the acknowledgement bytes.
func (am AppModule) OnRecvPacket(packet types.IBCPacket) ([]byte, error) {
	var packetData types.VITAPacketData
	if err := json.Unmarshal(packet.Data, &packetData); err != nil {
		ack := types.VITAAcknowledgement{Success: false, Error: fmt.Sprintf("failed to unmarshal packet data: %s", err)}
		return mustMarshalAck(ack), nil
	}

	ack, err := am.keeper.OnReceivePacket(sdk.Context{}, packetData)
	if err != nil {
		ack.Success = false
		if ack.Error == "" {
			ack.Error = err.Error()
		}
	}
	return mustMarshalAck(ack), nil
}

// OnAcknowledgementPacket processes the acknowledgement returned by the
// destination chain after it processed (or failed to process) an outbound packet.
func (am AppModule) OnAcknowledgementPacket(packet types.IBCPacket, acknowledgement []byte) error {
	var packetData types.VITAPacketData
	if err := json.Unmarshal(packet.Data, &packetData); err != nil {
		return fmt.Errorf("failed to unmarshal packet data: %w", err)
	}

	var ack types.VITAAcknowledgement
	if err := json.Unmarshal(acknowledgement, &ack); err != nil {
		return fmt.Errorf("failed to unmarshal acknowledgement: %w", err)
	}

	return am.keeper.OnAcknowledgePacket(sdk.Context{}, packetData, ack)
}

// OnTimeoutPacket handles the case where a sent packet was not received before
// the timeout height/timestamp.  Escrowed coins are returned to the sender.
func (am AppModule) OnTimeoutPacket(packet types.IBCPacket) error {
	var packetData types.VITAPacketData
	if err := json.Unmarshal(packet.Data, &packetData); err != nil {
		return fmt.Errorf("failed to unmarshal packet data: %w", err)
	}
	return am.keeper.OnTimeoutPacket(sdk.Context{}, packetData)
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// mustMarshalAck marshals a VITAAcknowledgement to JSON, panicking on error.
// Panics should never happen here since VITAAcknowledgement has no complex fields.
func mustMarshalAck(ack types.VITAAcknowledgement) []byte {
	bz, err := json.Marshal(ack)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal VITAAcknowledgement: %s", err))
	}
	return bz
}
