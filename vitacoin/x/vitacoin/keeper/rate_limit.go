package keeper

import (
	"context"
	"encoding/binary"

	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin/types"
)

const (
	// DefaultMinBlocksBetweenTx is 0 (disabled by default — no rate limit)
	DefaultMinBlocksBetweenTx uint64 = 0
)

// SetLastTxBlock stores the last transaction block height for an address.
func (k Keeper) SetLastTxBlock(ctx context.Context, address string, blockHeight int64) error {
	store := k.storeService.OpenKVStore(ctx)

	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(blockHeight))

	return store.Set(types.GetRateLimitKey(address), buf[:])
}

// GetLastTxBlock retrieves the last transaction block height for an address.
// Returns 0 and no error if the address has never transacted (first-time allowed).
func (k Keeper) GetLastTxBlock(ctx context.Context, address string) (int64, error) {
	store := k.storeService.OpenKVStore(ctx)

	bz, err := store.Get(types.GetRateLimitKey(address))
	if err != nil {
		return 0, err
	}
	if bz == nil {
		return 0, nil // never transacted
	}
	if len(bz) != 8 {
		return 0, nil // corrupt / old data → treat as fresh
	}

	return int64(binary.BigEndian.Uint64(bz)), nil
}

// SetMinBlocksBetweenTx stores the global rate-limit config in the KV store.
func (k Keeper) SetMinBlocksBetweenTx(ctx context.Context, minBlocks uint64) error {
	store := k.storeService.OpenKVStore(ctx)

	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], minBlocks)

	return store.Set(types.RateLimitConfigKey, buf[:])
}

// GetMinBlocksBetweenTx retrieves the global rate-limit config.
// Falls back to DefaultMinBlocksBetweenTx (0 = disabled) if not set.
func (k Keeper) GetMinBlocksBetweenTx(ctx context.Context) (uint64, error) {
	store := k.storeService.OpenKVStore(ctx)

	bz, err := store.Get(types.RateLimitConfigKey)
	if err != nil {
		return DefaultMinBlocksBetweenTx, err
	}
	if bz == nil {
		return DefaultMinBlocksBetweenTx, nil
	}
	if len(bz) != 8 {
		return DefaultMinBlocksBetweenTx, nil
	}

	return binary.BigEndian.Uint64(bz), nil
}
