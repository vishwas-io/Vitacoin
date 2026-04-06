package types

// IBCModule defines the interface that the vitacoin AppModule implements to handle
// IBC channel lifecycle and packet callbacks.  The signatures mirror the
// porttypes.IBCModule interface from ibc-go so that wiring the real router in a
// future job is a drop-in change.
//
// This local definition lets the module compile and be tested without importing
// ibc-go today.  When ibc-go is added as a dependency, replace this file with
// a blank import and the var-check:
//
//	var _ porttypes.IBCModule = AppModule{}
type IBCModule interface {
	// Channel lifecycle

	OnChanOpenInit(
		order ChannelOrder,
		connectionHops []string,
		portID string,
		channelID string,
		counterpartyPortID string,
		counterpartyChannelID string,
		version string,
	) (string, error)

	OnChanOpenTry(
		order ChannelOrder,
		connectionHops []string,
		portID string,
		channelID string,
		counterpartyPortID string,
		counterpartyChannelID string,
		counterpartyVersion string,
	) (string, error)

	OnChanOpenAck(portID, channelID, counterpartyChannelID, counterpartyVersion string) error
	OnChanOpenConfirm(portID, channelID string) error
	OnChanCloseInit(portID, channelID string) error
	OnChanCloseConfirm(portID, channelID string) error

	// Packet lifecycle

	OnRecvPacket(packet IBCPacket) ([]byte, error)
	OnAcknowledgementPacket(packet IBCPacket, acknowledgement []byte) error
	OnTimeoutPacket(packet IBCPacket) error
}

// ChannelOrder mirrors the IBC channel ordering enum.
type ChannelOrder int32

const (
	ChannelOrderNone      ChannelOrder = 0
	ChannelOrderUnordered ChannelOrder = 1
	ChannelOrderOrdered   ChannelOrder = 2
)

// IBCPacket is a minimal representation of an IBC packet sufficient for
// VitaCoin's cross-chain transfer logic.  When ibc-go is imported, replace
// usages with channeltypes.Packet.
type IBCPacket struct {
	Sequence           uint64
	SourcePort         string
	SourceChannel      string
	DestinationPort    string
	DestinationChannel string
	Data               []byte // JSON-encoded VITAPacketData
	TimeoutHeight      uint64
	TimeoutTimestamp   uint64
}
