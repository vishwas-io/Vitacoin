#!/bin/bash
# VitaCoin node initialization script
# Run on your validator server
set -e

NODE_MONIKER=${1:-"my-validator"}
CHAIN_ID="vitacoin-1"
BINARY="./build/vitacoind"

echo "Initializing VitaCoin node: $NODE_MONIKER"
$BINARY init $NODE_MONIKER --chain-id $CHAIN_ID

echo "Node initialized. Keys are in ~/.vitacoin/config/"
echo "IMPORTANT: Back up priv_validator_key.json and node_key.json SECURELY"
echo "Never share these files. Never commit them to git."
