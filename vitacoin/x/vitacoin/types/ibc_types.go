package types

import (
	"cosmossdk.io/math"
)

// IBC port and version constants for VitaCoin cross-chain transfers.
const (
	// IBCPortID is the IBC port identifier for the vitacoin module.
	IBCPortID = "vitacoin"

	// IBCVersion is the IBC app version negotiated during channel handshake.
	IBCVersion = "vitacoin-1"

	// MaxIBCMemo is the maximum allowed length (bytes) for the transfer memo field.
	MaxIBCMemo = 256

	// Event type constants for IBC operations.
	EventTypeIBCSend    = "ibc_vita_send"
	EventTypeIBCReceive = "ibc_vita_receive"
	EventTypeIBCAck     = "ibc_vita_acknowledge"
	EventTypeIBCTimeout = "ibc_vita_timeout"

	// Attribute keys for IBC events.
	AttributeKeyIBCSender        = "sender"
	AttributeKeyIBCReceiver      = "receiver"
	AttributeKeyIBCChannel       = "source_channel"
	AttributeKeyIBCSequence      = "sequence"
	AttributeKeyIBCAmount        = "amount"
	AttributeKeyIBCDenom         = "denom"
	AttributeKeyIBCAckSuccess    = "ack_success"
	AttributeKeyIBCAckError      = "ack_error"
)

// VITAPacketData is the payload carried inside an IBC packet for cross-chain VITA transfers.
// It is JSON-marshalled and stored in the IBC packet data field.
type VITAPacketData struct {
	// Sender is the bech32 address of the sender on the source chain.
	Sender string `json:"sender"`

	// Receiver is the bech32 address of the recipient on the destination chain.
	Receiver string `json:"receiver"`

	// Amount is the quantity of VITA being transferred (in base denomination).
	Amount math.Int `json:"amount"`

	// Denom is the token denomination (e.g. "avita").
	Denom string `json:"denom"`

	// Memo is an optional human-readable note (max MaxIBCMemo bytes).
	Memo string `json:"memo,omitempty"`

	// Sequence is the IBC packet sequence number set when the packet is queued.
	Sequence uint64 `json:"sequence,omitempty"`

	// SourceChannel is the channel identifier on the source chain.
	SourceChannel string `json:"source_channel,omitempty"`
}

// VITAAcknowledgement is the acknowledgement payload returned by the destination chain
// after processing a VITAPacketData packet.
type VITAAcknowledgement struct {
	// Success indicates whether the packet was processed successfully.
	Success bool `json:"success"`

	// Error contains a human-readable error description when Success is false.
	Error string `json:"error,omitempty"`
}

// Validate performs basic sanity checks on a VITAPacketData.
func (p VITAPacketData) Validate() error {
	if p.Sender == "" {
		return ErrInvalidSender
	}
	if p.Receiver == "" {
		return ErrInvalidReceiver
	}
	if p.Denom == "" {
		return ErrInvalidDenom
	}
	if !p.Amount.IsPositive() {
		return ErrInvalidAmount
	}
	if len(p.Memo) > MaxIBCMemo {
		return ErrMemoTooLong
	}
	return nil
}
