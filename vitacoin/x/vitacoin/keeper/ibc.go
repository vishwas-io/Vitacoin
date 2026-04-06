package keeper

// Phase 6: IBC Integration — Channel + Packet Types
//
// This file implements cross-chain VITA transfer logic using KV-based pending
// packet tracking. The IBC channel router hookup (channelKeeper) is a separate
// step; for now packets are escrowed and recorded in KV so they survive restarts.

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin/types"
)

// ─── IBC Sequence Counter Key ────────────────────────────────────────────────

// ibcSequenceKey is a dedicated KV key for the monotonic packet sequence counter.
var ibcSequenceKey = []byte{0x41}

// getNextIBCSequence atomically increments and returns the next packet sequence.
func (k Keeper) getNextIBCSequence(ctx context.Context) (uint64, error) {
	store := k.storeService.OpenKVStore(ctx)

	bz, err := store.Get(ibcSequenceKey)
	if err != nil {
		return 0, fmt.Errorf("failed to read IBC sequence: %w", err)
	}

	var seq uint64
	if bz != nil {
		seq = binary.BigEndian.Uint64(bz)
	}

	seq++

	next := make([]byte, 8)
	binary.BigEndian.PutUint64(next, seq)
	if err := store.Set(ibcSequenceKey, next); err != nil {
		return 0, fmt.Errorf("failed to store IBC sequence: %w", err)
	}

	return seq, nil
}

// ─── IBCSendVITA ─────────────────────────────────────────────────────────────

// IBCSendVITA initiates a cross-chain VITA transfer from the local chain.
//
// Steps:
//  1. Validate all inputs.
//  2. Escrow coins: sender → vitacoin module account.
//  3. Build a VITAPacketData and assign a monotonic sequence number.
//  4. Persist the pending packet to KV (channelKeeper hookup is Phase 6 Job 2).
//  5. Emit EventTypeIBCSend.
func (k Keeper) IBCSendVITA(
	ctx context.Context,
	sender, receiver, sourceChannel string,
	amount sdk.Coin,
	memo string,
) error {
	// 1. Validate inputs
	if receiver == "" {
		return types.ErrInvalidReceiver
	}
	if !amount.Amount.IsPositive() {
		return types.ErrInvalidAmount
	}
	if len(memo) > types.MaxIBCMemo {
		return types.ErrMemoTooLong
	}
	if sender == "" {
		return types.ErrInvalidSender
	}

	senderAddr, err := sdk.AccAddressFromBech32(sender)
	if err != nil {
		return fmt.Errorf("%w: %s", types.ErrInvalidSender, err)
	}

	// 2. Escrow coins into the module account
	if err := k.bankKeeper.SendCoinsFromAccountToModule(
		ctx,
		senderAddr,
		types.ModuleName,
		sdk.NewCoins(amount),
	); err != nil {
		return fmt.Errorf("failed to escrow IBC coins: %w", err)
	}

	// 3. Assign sequence and build packet
	seq, err := k.getNextIBCSequence(ctx)
	if err != nil {
		return err
	}

	packet := types.VITAPacketData{
		Sender:        sender,
		Receiver:      receiver,
		Amount:        amount.Amount,
		Denom:         amount.Denom,
		Memo:          memo,
		Sequence:      seq,
		SourceChannel: sourceChannel,
	}

	// 4. Persist pending packet
	if err := k.setIBCPendingPacket(ctx, packet); err != nil {
		return fmt.Errorf("failed to store pending IBC packet: %w", err)
	}

	// 5. Emit event
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeIBCSend,
		sdk.NewAttribute(types.AttributeKeyIBCSender, sender),
		sdk.NewAttribute(types.AttributeKeyIBCReceiver, receiver),
		sdk.NewAttribute(types.AttributeKeyIBCChannel, sourceChannel),
		sdk.NewAttribute(types.AttributeKeyIBCSequence, fmt.Sprintf("%d", seq)),
		sdk.NewAttribute(types.AttributeKeyIBCAmount, amount.Amount.String()),
		sdk.NewAttribute(types.AttributeKeyIBCDenom, amount.Denom),
	))

	k.logger.Info("IBC VITA send queued",
		"sender", sender,
		"receiver", receiver,
		"channel", sourceChannel,
		"amount", amount,
		"sequence", seq,
	)

	return nil
}

// ─── OnReceivePacket ─────────────────────────────────────────────────────────

// OnReceivePacket processes an inbound IBC packet on the destination chain.
//
// Steps:
//  1. Validate packet fields.
//  2. Mint equivalent VITA to the receiver.
//  3. Emit EventTypeIBCReceive.
//  4. Return a success acknowledgement.
func (k Keeper) OnReceivePacket(ctx context.Context, packet types.VITAPacketData) (types.VITAAcknowledgement, error) {
	// 1. Validate
	if err := packet.Validate(); err != nil {
		return types.VITAAcknowledgement{Success: false, Error: err.Error()}, err
	}

	receiverAddr, err := sdk.AccAddressFromBech32(packet.Receiver)
	if err != nil {
		errMsg := fmt.Sprintf("invalid receiver address: %s", err)
		return types.VITAAcknowledgement{Success: false, Error: errMsg},
			fmt.Errorf("%w: %s", types.ErrInvalidReceiver, err)
	}

	// 2. Mint VITA to module then send to receiver
	coins := sdk.NewCoins(sdk.NewCoin(packet.Denom, packet.Amount))

	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
		errMsg := fmt.Sprintf("failed to mint coins: %s", err)
		return types.VITAAcknowledgement{Success: false, Error: errMsg}, err
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiverAddr, coins); err != nil {
		errMsg := fmt.Sprintf("failed to send coins to receiver: %s", err)
		return types.VITAAcknowledgement{Success: false, Error: errMsg}, err
	}

	// 3. Emit event
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeIBCReceive,
		sdk.NewAttribute(types.AttributeKeyIBCSender, packet.Sender),
		sdk.NewAttribute(types.AttributeKeyIBCReceiver, packet.Receiver),
		sdk.NewAttribute(types.AttributeKeyIBCSequence, fmt.Sprintf("%d", packet.Sequence)),
		sdk.NewAttribute(types.AttributeKeyIBCAmount, packet.Amount.String()),
		sdk.NewAttribute(types.AttributeKeyIBCDenom, packet.Denom),
	))

	k.logger.Info("IBC VITA received",
		"sender", packet.Sender,
		"receiver", packet.Receiver,
		"amount", packet.Amount,
		"denom", packet.Denom,
		"sequence", packet.Sequence,
	)

	return types.VITAAcknowledgement{Success: true}, nil
}

// ─── OnAcknowledgePacket ─────────────────────────────────────────────────────

// OnAcknowledgePacket handles the acknowledgement returned by the destination chain.
//
//   - Success → delete pending packet record, emit success event.
//   - Failure → refund escrowed coins to sender, emit refund event.
func (k Keeper) OnAcknowledgePacket(ctx context.Context, packetData types.VITAPacketData, ack types.VITAAcknowledgement) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if ack.Success {
		// Delete the pending packet record — transfer is complete.
		if err := k.deleteIBCPendingPacket(ctx, packetData.Sequence); err != nil {
			k.logger.Error("failed to delete pending IBC packet after success", "sequence", packetData.Sequence, "error", err)
			// Non-fatal: the transfer already succeeded on the remote chain.
		}

		sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeIBCAck,
			sdk.NewAttribute(types.AttributeKeyIBCSender, packetData.Sender),
			sdk.NewAttribute(types.AttributeKeyIBCSequence, fmt.Sprintf("%d", packetData.Sequence)),
			sdk.NewAttribute(types.AttributeKeyIBCAckSuccess, "true"),
		))

		k.logger.Info("IBC VITA ack — success", "sequence", packetData.Sequence, "sender", packetData.Sender)
		return nil
	}

	// Acknowledgement failure — refund sender.
	if err := k.ibcRefundSender(ctx, packetData); err != nil {
		return fmt.Errorf("failed to refund sender on ack failure: %w", err)
	}

	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeIBCAck,
		sdk.NewAttribute(types.AttributeKeyIBCSender, packetData.Sender),
		sdk.NewAttribute(types.AttributeKeyIBCSequence, fmt.Sprintf("%d", packetData.Sequence)),
		sdk.NewAttribute(types.AttributeKeyIBCAckSuccess, "false"),
		sdk.NewAttribute(types.AttributeKeyIBCAckError, ack.Error),
	))

	k.logger.Info("IBC VITA ack — failure, refunded", "sequence", packetData.Sequence, "sender", packetData.Sender, "error", ack.Error)
	return nil
}

// ─── OnTimeoutPacket ─────────────────────────────────────────────────────────

// OnTimeoutPacket handles the case where a sent IBC packet timed out before
// being received on the destination chain. Escrowed coins are returned to sender.
func (k Keeper) OnTimeoutPacket(ctx context.Context, packetData types.VITAPacketData) error {
	if err := k.ibcRefundSender(ctx, packetData); err != nil {
		return fmt.Errorf("failed to refund sender on timeout: %w", err)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeIBCTimeout,
		sdk.NewAttribute(types.AttributeKeyIBCSender, packetData.Sender),
		sdk.NewAttribute(types.AttributeKeyIBCSequence, fmt.Sprintf("%d", packetData.Sequence)),
		sdk.NewAttribute(types.AttributeKeyIBCAmount, packetData.Amount.String()),
		sdk.NewAttribute(types.AttributeKeyIBCDenom, packetData.Denom),
	))

	k.logger.Info("IBC VITA timeout — refunded", "sequence", packetData.Sequence, "sender", packetData.Sender)
	return nil
}

// ─── GetIBCPendingPackets ─────────────────────────────────────────────────────

// GetIBCPendingPackets returns all pending outgoing IBC packets stored in KV.
func (k Keeper) GetIBCPendingPackets(ctx context.Context) ([]types.VITAPacketData, error) {
	store := k.storeService.OpenKVStore(ctx)
	var packets []types.VITAPacketData

	iter, err := store.Iterator(
		types.KeyPrefixIBCPacket,
		storetypes.PrefixEndBytes(types.KeyPrefixIBCPacket),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open IBC packet iterator: %w", err)
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var p types.VITAPacketData
		if err := json.Unmarshal(iter.Value(), &p); err != nil {
			k.logger.Error("failed to unmarshal pending IBC packet", "error", err)
			continue
		}
		packets = append(packets, p)
	}

	return packets, nil
}

// ─── Internal helpers ─────────────────────────────────────────────────────────

// setIBCPendingPacket stores a packet record in KV.
func (k Keeper) setIBCPendingPacket(ctx context.Context, packet types.VITAPacketData) error {
	bz, err := json.Marshal(packet)
	if err != nil {
		return fmt.Errorf("failed to marshal IBC packet: %w", err)
	}
	store := k.storeService.OpenKVStore(ctx)
	return store.Set(types.GetIBCPacketKey(packet.Sequence), bz)
}

// deleteIBCPendingPacket removes a packet record from KV.
func (k Keeper) deleteIBCPendingPacket(ctx context.Context, sequence uint64) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Delete(types.GetIBCPacketKey(sequence))
}

// ibcRefundSender releases escrowed coins back to the original sender and removes
// the pending packet record.
func (k Keeper) ibcRefundSender(ctx context.Context, packet types.VITAPacketData) error {
	senderAddr, err := sdk.AccAddressFromBech32(packet.Sender)
	if err != nil {
		return fmt.Errorf("%w: %s", types.ErrInvalidSender, err)
	}

	coins := sdk.NewCoins(sdk.NewCoin(packet.Denom, packet.Amount))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, senderAddr, coins); err != nil {
		return fmt.Errorf("failed to refund escrowed coins: %w", err)
	}

	// Best-effort deletion of the pending record.
	if err := k.deleteIBCPendingPacket(ctx, packet.Sequence); err != nil {
		k.logger.Error("failed to delete pending IBC packet on refund", "sequence", packet.Sequence, "error", err)
	}

	return nil
}
