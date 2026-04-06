#!/bin/bash
# Validates a genesis.json before launch
BINARY="./build/vitacoind"
$BINARY genesis validate-genesis
echo "Genesis valid: $?"
