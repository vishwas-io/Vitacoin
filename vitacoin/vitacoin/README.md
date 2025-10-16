# 🪙 VITACOIN - The Cryptocurrency Blockchain

This folder contains the **VITACOIN blockchain** - the cryptocurrency that powers the entire ecosystem.

## What is This?

**VITACOIN** is a Cosmos SDK-based blockchain that provides:
- **VITA Token**: The native cryptocurrency
- **Validators**: Network security through Proof-of-Stake
- **Governance**: On-chain decision making
- **Staking**: Earn rewards for securing the network
- **IBC**: Cross-chain interoperability

**Think of it as**: The "money layer" - like Bitcoin or Ethereum, but optimized for payments.

## Relationship to VITAPAY

```
VITAPAY (payment network)
    ↓ uses
VITACOIN (this blockchain)
```

VITAPAY is the user-friendly payment app built ON TOP OF this blockchain. This folder contains the underlying blockchain infrastructure.

## What's In This Folder?

```
vitacoin/
├── README.md          # This file
├── app/              # Blockchain application setup
├── cmd/              # vitacoind binary (node software)
├── proto/            # Protocol buffer definitions
├── x/                # Custom Cosmos SDK modules
├── testutil/         # Testing utilities
├── build/            # Compiled binaries
├── Makefile          # Build commands
├── config.yml        # Chain configuration
└── buf.yaml          # Protobuf build config
```

## Quick Start

### Prerequisites
- Go 1.21 or higher
- Make
- Protocol Buffers compiler

### Build the Blockchain

```bash
# Install dependencies
make install

# Build the node software
make build

# The binary will be in ./build/vitacoind
```

### Run a Local Node

```bash
# Initialize a local node
./build/vitacoind init mynode --chain-id vitacoin-local

# Start the node
./build/vitacoind start
```

### Run Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage
```

## Key Commands

### Node Operations
```bash
# Start node
vitacoind start

# Check node status
vitacoind status

# View logs
vitacoind logs
```

### Wallet Operations
```bash
# Create a wallet
vitacoind keys add mywallet

# Check balance
vitacoind query bank balances <address>

# Send tokens
vitacoind tx bank send <from> <to> <amount>uvita --chain-id vitacoin-1
```

### Staking Operations
```bash
# View validators
vitacoind query staking validators

# Delegate to validator
vitacoind tx staking delegate <validator-addr> <amount>uvita --from <key>

# Check delegation
vitacoind query staking delegations <delegator-addr>
```

### Governance Operations
```bash
# Submit proposal
vitacoind tx gov submit-proposal [proposal-file] --from <key>

# Vote on proposal
vitacoind tx gov vote <proposal-id> yes --from <key>

# Check proposal status
vitacoind query gov proposal <proposal-id>
```

## Architecture

### Cosmos SDK Modules

**Standard Modules:**
- `auth` - Account management
- `bank` - Token transfers
- `staking` - Validator staking
- `distribution` - Reward distribution
- `gov` - Governance proposals
- `slashing` - Validator penalties
- `ibc` - Cross-chain communication

**Custom Modules:**
- `x/vitacoin` - VITACOIN-specific features (see `x/vitacoin/README.md`)

### Consensus

- **Algorithm**: CometBFT (formerly Tendermint)
- **Type**: Proof-of-Stake (PoS)
- **Block Time**: ~6 seconds
- **Finality**: Instant (no reorgs)

### Network Specifications

| Property | Value |
|----------|-------|
| Chain ID | vitacoin-1 |
| Token Denom | uvita |
| Token Symbol | VITA |
| Decimals | 18 |
| Total Supply | 1,000,000,000 VITA |
| Block Time | ~6 seconds |
| Unbonding Period | 21 days |

## Development

### Project Structure

```
app/
├── ante.go           # Transaction preprocessing
├── app.go           # Main application setup
├── encoding.go      # Encoding configuration
├── genesis.go       # Genesis state
└── params.go        # Network parameters

cmd/vitacoind/
├── main.go          # Entry point
└── cmd/             # CLI commands

proto/vitacoin/
└── v1/              # Protobuf definitions

x/vitacoin/
├── keeper/          # State management
├── types/           # Data types
└── module.go        # Module definition
```

### Adding Features

1. **Define Protobuf Messages**: Add to `proto/vitacoin/v1/`
2. **Generate Code**: Run `make proto-gen`
3. **Implement Logic**: Add to `x/vitacoin/keeper/`
4. **Add Tests**: Create in `x/vitacoin/keeper/*_test.go`
5. **Update Module**: Modify `x/vitacoin/module.go`

### Testing

```bash
# Unit tests
make test-unit

# Integration tests
make test-integration

# All tests
make test

# With coverage report
make test-coverage
```

## Token Economics

### Supply
- **Total Supply**: 1,000,000,000 VITA
- **Genesis Allocation**: 300M VITA
- **Staking Rewards**: 400M VITA (over 10 years)
- **Ecosystem Fund**: 200M VITA
- **Governance Reserve**: 100M VITA

### Fees
- **Transaction Fee**: 0.1% of amount
- **Distribution**: 
  - 50% to validators
  - 25% burned (deflationary)
  - 25% to treasury

### Inflation
- **Annual Rate**: 3% - 10% (dynamic)
- **Target Bonded**: 67%
- **Purpose**: Staking rewards

## Running a Validator

### Requirements
- Dedicated server (24/7 uptime)
- 4+ CPU cores
- 16GB+ RAM
- 500GB+ SSD storage
- 100Mbps+ network

### Setup Steps

1. **Install Software**
   ```bash
   git clone https://github.com/esspron/vitacoin
   cd vitacoin/vitacoin
   make install
   ```

2. **Initialize Node**
   ```bash
   vitacoind init myvalidator --chain-id vitacoin-1
   ```

3. **Configure**
   ```bash
   # Download genesis
   wget -O ~/.vitacoin/config/genesis.json https://raw.githubusercontent.com/esspron/vitacoin/main/genesis.json
   
   # Set seeds/peers
   # Edit ~/.vitacoin/config/config.toml
   ```

4. **Create Validator**
   ```bash
   vitacoind tx staking create-validator \
     --amount=1000000uvita \
     --pubkey=$(vitacoind tendermint show-validator) \
     --moniker="My Validator" \
     --commission-rate="0.05" \
     --commission-max-rate="0.10" \
     --commission-max-change-rate="0.01" \
     --min-self-delegation="1" \
     --from=mykey
   ```

[Full Validator Guide →](../docs/architecture/VALIDATOR_GUIDE.md) (Coming soon)

## API Documentation

### gRPC
- **Endpoint**: `grpc.vitacoin.network:9090`
- **TLS**: Yes

### REST API
- **Endpoint**: `https://api.vitacoin.network`
- **Docs**: [API Reference](../docs/api/README.md) (Coming soon)

### WebSocket
- **Endpoint**: `wss://ws.vitacoin.network`
- **Purpose**: Real-time events

## Resources

### Documentation
- [Main Documentation](../docs/README.md) - Complete docs index
- [VITACOIN Guide](../docs/VITACOIN.md) - Cryptocurrency overview
- [Architecture](../docs/architecture/ARCHITECTURE.md) - System design
- [Getting Started](../docs/development/GETTING_STARTED.md) - Developer guide

### External Resources
- [Cosmos SDK Docs](https://docs.cosmos.network/)
- [CometBFT Docs](https://docs.cometbft.com/)
- [CosmJS Docs](https://cosmos.github.io/cosmjs/)

### Community
- **GitHub**: [github.com/esspron/vitacoin](https://github.com/esspron/vitacoin)
- **Discord**: Coming soon
- **Forum**: Coming soon

## Support

- **Technical Issues**: Open an issue on GitHub
- **Security**: security@vitacoin.network
- **General**: contact@vitacoin.network

## License

Apache 2.0 - See [LICENSE](../LICENSE)

---

**This is the VITACOIN blockchain.** For the payment network, see [../vitapay/](../vitapay/)
