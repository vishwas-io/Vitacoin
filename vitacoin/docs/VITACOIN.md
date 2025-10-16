# 🪙 VITACOIN - The Cryptocurrency

## What is VITACOIN?

**VITACOIN (VITA)** is a decentralized blockchain cryptocurrency built on the Cosmos SDK. It's the actual digital money - the token that people send, receive, store, and trade.

Think of VITACOIN as:
- **Like Bitcoin or Ethereum**: A cryptocurrency token
- **Not a company**: A decentralized blockchain network
- **The money itself**: Not the payment app (that's VITAPAY)

---

## 🎯 Purpose

VITACOIN exists to be:
1. **A Store of Value**: Hold it like digital gold
2. **A Medium of Exchange**: Use it for transactions
3. **A Unit of Account**: Price things in VITA
4. **A Network Token**: Pay for transactions, stake for security

---

## ⚙️ How It Works

### Blockchain Technology
```
┌────────────────────────────────────────────┐
│         VITACOIN BLOCKCHAIN                │
├────────────────────────────────────────────┤
│                                            │
│  Consensus: Proof-of-Stake (PoS)          │
│  ├─ Validators secure the network         │
│  ├─ Stake VITA to become validator        │
│  └─ Earn rewards for validation           │
│                                            │
│  Transactions:                             │
│  ├─ Send/receive VITA tokens              │
│  ├─ Fee: 0.1% per transaction             │
│  ├─ Speed: ~5 second finality             │
│  └─ Transparent on-chain                  │
│                                            │
│  Governance:                               │
│  ├─ Token holders vote on proposals       │
│  ├─ On-chain governance system            │
│  └─ Community-driven development          │
│                                            │
└────────────────────────────────────────────┘
```

### Key Technology Components

**Cosmos SDK**
- Industry-standard blockchain framework
- Battle-tested and secure
- Modular and extensible
- Used by 100+ blockchains

**CometBFT Consensus**
- Byzantine Fault Tolerant consensus
- Instant finality
- Energy efficient (not Proof-of-Work)
- Scalable and fast

**IBC Protocol**
- Inter-Blockchain Communication
- Connect to other Cosmos chains
- Cross-chain asset transfers
- Future-proof interoperability

---

## 💰 Token Economics

### Token Specifications

| Property | Value |
|----------|-------|
| **Name** | VITACOIN |
| **Symbol** | VITA |
| **Total Supply** | 1,000,000,000 (1 Billion) |
| **Decimals** | 18 |
| **Smallest Unit** | uvita |
| **1 VITA** | 1,000,000,000,000,000,000 uvita |

### Supply Distribution

```
Total Supply: 1,000,000,000 VITA

├─ 400,000,000 VITA (40%) - Staking Rewards
│  └─ Released over 10 years
│  └─ Rewards validators and delegators
│
├─ 300,000,000 VITA (30%) - Genesis Allocation
│  ├─ Team & Advisors: 100M (10%) - 4 year vesting
│  ├─ Early Investors: 50M (5%) - 2 year vesting
│  ├─ Foundation: 100M (10%) - Immediate
│  └─ Community Airdrop: 50M (5%) - Launch event
│
├─ 200,000,000 VITA (20%) - Ecosystem Development
│  └─ Developer grants
│  └─ Partnerships
│  └─ Liquidity provision
│  └─ VITAPAY development
│
└─ 100,000,000 VITA (10%) - Governance Reserve
   └─ Community-controlled treasury
   └─ Voted on by token holders
```

### Inflation & Staking

**Annual Inflation**: 3% - 10% (dynamic)
- **Purpose**: Reward validators and stakers
- **Target**: 67% of supply staked
- **Mechanism**: 
  - If <67% staked → inflation increases (up to 10%)
  - If >67% staked → inflation decreases (down to 3%)
  - Incentivizes network security

**Staking APR**: ~7% average
- Earn rewards for securing the network
- Delegators can stake without running validator
- Flexible staking (no minimum lock period)
- Auto-compounding available

### Deflationary Mechanism

**Fee Burning**: 25% of all transaction fees are burned
- Reduces circulating supply
- Creates scarcity
- Balances inflationary staking rewards
- Transparent on-chain tracking

**Example**:
- 1M daily transactions × 0.1% fee × 25% burned
- ~250 VITA burned daily
- ~91,250 VITA burned annually
- Long-term deflationary pressure

---

## 💸 Transaction Fees

### Fee Structure

**Rate**: 0.1% of transaction amount

**Example Fees**:
| Transaction Amount | Fee (0.1%) |
|-------------------|------------|
| 10 VITA | 0.01 VITA |
| 100 VITA | 0.1 VITA |
| 1,000 VITA | 1 VITA |
| 10,000 VITA | 10 VITA |

### Fee Distribution (Transparent & On-Chain)

Every fee is split three ways:

```
Transaction Fee: 0.1 VITA

├─ 0.05 VITA (50%) → Validators
│  └─ Rewards nodes securing network
│  └─ Distributed proportionally by stake
│
├─ 0.025 VITA (25%) → Burned Forever
│  └─ Sent to burn address (0x000...dead)
│  └─ Reduces total supply
│  └─ Increases scarcity
│
└─ 0.025 VITA (25%) → Treasury
   └─ Development fund
   └─ Ecosystem growth
   └─ Governance controlled
```

### Why This Split?

**50% to Validators**: 
- Incentivizes network security
- Rewards infrastructure operators
- Ensures decentralization

**25% Burned**:
- Creates deflationary pressure
- Benefits all token holders
- Increases value over time

**25% to Treasury**:
- Funds ongoing development
- Community grants and bounties
- Ecosystem partnerships
- Controlled by governance

### Fee Transparency

All fees are tracked on-chain:
- View in block explorer
- Real-time fee statistics
- Total burned supply visible
- Treasury balance public
- Validator earnings tracked

---

## 🔒 Staking

### What is Staking?

Staking is locking your VITA tokens to help secure the network and earn rewards.

### How to Stake

**Option 1: Delegate to a Validator**
```bash
# Find validators
vitacoind query staking validators

# Delegate your VITA
vitacoind tx staking delegate <validator-address> 1000000000000000000000uvita --from mykey

# That's 1000 VITA staked
```

**Option 2: Run Your Own Validator**
- Requires technical expertise
- Need minimum self-delegation
- Run validator node 24/7
- Earn higher rewards

### Staking Rewards

**Base APR**: ~7% annually
- Variable based on total staked
- Paid in VITA tokens
- Claimed automatically or manually
- Compounding available

**Reward Calculation**:
```
Your Rewards = (Your Stake / Total Staked) × Annual Inflation × Validator Commission
```

**Example**:
- You stake: 10,000 VITA
- Validator commission: 5%
- Total staked: 670M VITA (67%)
- Annual inflation: 7%
- Your annual rewards: ~665 VITA (6.65% APR)

### Validator Selection

Choose validators based on:
- **Commission Rate**: 5-10% typical
- **Uptime**: >99% is good
- **Voting Participation**: Active in governance
- **Community Reputation**: Trusted operators

### Unbonding Period

**21 Days**: Standard unbonding period
- Prevents "nothing at stake" attacks
- Ensures network security
- Plan accordingly for liquidity

### Slashing

Validators can be penalized for:
- **Downtime**: 0.01% slash for extended offline
- **Double Signing**: 5% slash for malicious behavior

**As a Delegator**: Your stake is slashed if your validator misbehaves
- **Choose validators carefully**
- **Diversify across multiple validators**

---

## 🏛️ Governance

### On-Chain Governance

VITA token holders control the network through voting.

### Proposal Types

**1. Text Proposals**
- General governance
- Strategic decisions
- Community initiatives

**2. Parameter Change Proposals**
- Modify blockchain parameters
- Adjust fees, inflation, etc.
- Network upgrades

**3. Community Pool Spend**
- Allocate treasury funds
- Grant programs
- Ecosystem development

**4. Software Upgrade Proposals**
- Network upgrades
- New features
- Bug fixes

### Voting Process

```
1. Proposal Submission
   ├─ Anyone can submit
   ├─ Deposit required: 10,000 VITA
   └─ Refunded if proposal passes or gets enough votes

2. Deposit Period (7 days)
   ├─ Community adds deposits
   └─ Reaches minimum deposit to proceed

3. Voting Period (14 days)
   ├─ All stakers can vote
   ├─ Voting power = staked amount
   └─ Options: Yes, No, NoWithVeto, Abstain

4. Results
   ├─ Passes if: >50% Yes, <33.4% NoWithVeto, >40% turnout
   └─ Executed automatically if passed
```

### Voting Power

- **1 VITA staked = 1 vote**
- Validators vote on behalf of delegators
- Delegators can override validator's vote
- Abstain doesn't count as voting power

### Why Governance Matters

- **Decentralized Control**: Community owns the network
- **Transparent Decisions**: All on-chain
- **Aligned Incentives**: Stakeholders decide future
- **No Central Authority**: True decentralization

---

## 🌐 Cross-Chain (IBC)

### Inter-Blockchain Communication

VITACOIN supports IBC - the "internet of blockchains"

### What IBC Enables

**Token Transfers**:
- Send VITA to other Cosmos chains
- Receive tokens from other chains
- Atomic swaps

**Cross-Chain DeFi**:
- Use VITA in other chain's DeFi protocols
- Liquidity pools across chains
- Multi-chain yield farming

**Interoperability**:
- Connect to 100+ Cosmos chains
- Bridge to Ethereum (via Axelar/Gravity)
- Future-proof connectivity

### Supported Chains (Future)

- Cosmos Hub (ATOM)
- Osmosis (OSMO)
- Juno (JUNO)
- Akash (AKT)
- And more...

---

## 🔐 Security

### Consensus Security

**Byzantine Fault Tolerance**:
- Tolerates up to 1/3 malicious validators
- Instant finality (no reorgs)
- Proven secure algorithm

**Economic Security**:
- Validators have "skin in the game"
- Slashing for misbehavior
- High cost to attack network

### Transaction Security

**Cryptographic Signatures**:
- Every transaction signed with private key
- Ed25519 signature algorithm
- Impossible to forge

**Address Format**:
- Bech32 encoding
- Checksums prevent typos
- Format: `vita1abc123...xyz`

### Network Security

**Distributed Validators**:
- 100+ validators target
- Geographic distribution
- No single point of failure

**Open Source**:
- Code publicly audited
- Community reviewed
- Transparent development

---

## 🛠️ For Developers

### Building on VITACOIN

**Cosmos SDK Modules**:
- `bank`: Token transfers
- `staking`: Validator operations
- `gov`: Governance proposals
- `distribution`: Reward distribution
- `slashing`: Validator penalties
- `ibc`: Cross-chain transfers

**Custom Modules**:
- Build your own modules
- Extend functionality
- Integrate with VITACOIN

### APIs

**gRPC API**:
```
grpc.vitacoin.network:9090
```

**REST API**:
```
https://api.vitacoin.network
```

**WebSocket**:
```
wss://ws.vitacoin.network
```

### SDKs

**CosmJS** (JavaScript/TypeScript):
```typescript
import { SigningStargateClient } from "@cosmjs/stargate";

const client = await SigningStargateClient.connect(
  "https://rpc.vitacoin.network"
);

// Send transaction
const result = await client.sendTokens(
  sender,
  recipient,
  [{ denom: "uvita", amount: "1000000000000000000000" }],
  "auto"
);
```

**Cosmos SDK** (Go):
```go
import "github.com/cosmos/cosmos-sdk/types"

// Create transaction
msg := banktypes.NewMsgSend(
    from,
    to,
    sdk.NewCoins(sdk.NewInt64Coin("uvita", 1000)),
)
```

### Running a Node

**Full Node**:
```bash
# Install
go install ./cmd/vitacoind

# Initialize
vitacoind init mynode --chain-id vitacoin-1

# Start
vitacoind start
```

**Validator Node**:
```bash
# Create validator
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

---

## 📊 Network Statistics (Live Stats Coming Soon)

### Current Metrics
- **Total Supply**: 1,000,000,000 VITA
- **Circulating Supply**: TBD at launch
- **Total Staked**: Target 67%
- **Number of Validators**: Target 100+
- **Average Block Time**: ~6 seconds
- **Transaction Throughput**: 1000+ TPS
- **Total Transactions**: Tracked on launch
- **Total Burned**: Tracked on launch

### Block Explorer

View all blockchain activity:
- **Block Explorer**: https://explorer.vitacoin.network (Coming soon)
- See transactions in real-time
- Track addresses
- View validator status
- Monitor governance proposals

---

## 🎯 Use Cases

### 1. Store of Value
- Hold VITA as digital asset
- Long-term investment
- Portfolio diversification

### 2. Medium of Exchange
- Send/receive payments via VITAPAY
- E-commerce transactions
- Peer-to-peer transfers
- Remittances

### 3. Network Utility
- Pay transaction fees
- Stake for validator rewards
- Participate in governance
- Access future services

### 4. DeFi Applications (Future)
- Liquidity pools
- Lending/borrowing
- Yield farming
- Synthetic assets

---

## 🚀 Getting Started with VITACOIN

### For Holders

**Option 1: Use VITAPAY Wallet**
- Download VITAPAY app (easiest)
- Create wallet
- Buy VITA tokens
- Start using immediately

**Option 2: Use Command Line**
```bash
# Create wallet
vitacoind keys add mywallet

# Check balance
vitacoind query bank balances <your-address>

# Send VITA
vitacoind tx bank send mywallet <recipient> 1000uvita
```

### For Validators

1. **Setup Infrastructure**
   - Dedicated server
   - High uptime
   - Good network

2. **Install Software**
   ```bash
   git clone https://github.com/esspron/vitacoin
   cd vitacoin/vitacoin
   make install
   ```

3. **Sync Node**
   ```bash
   vitacoind start
   ```

4. **Create Validator**
   ```bash
   vitacoind tx staking create-validator ...
   ```

[Full Validator Guide →](./architecture/VALIDATOR_GUIDE.md) (Coming soon)

### For Developers

1. **Clone Repository**
   ```bash
   git clone https://github.com/esspron/vitacoin
   ```

2. **Build Blockchain**
   ```bash
   cd vitacoin/vitacoin
   make build
   ```

3. **Start Local Node**
   ```bash
   ./scripts/localnet.sh
   ```

4. **Start Building**
   - Read documentation
   - Explore modules
   - Build custom functionality

[Developer Guide →](./development/GETTING_STARTED.md)

---

## 🆚 VITACOIN vs Other Cryptocurrencies

### vs Bitcoin (BTC)
| Feature | Bitcoin | VITACOIN |
|---------|---------|----------|
| Consensus | Proof-of-Work (energy intensive) | Proof-of-Stake (energy efficient) |
| Transaction Speed | 10-60 minutes | 5 seconds |
| Transaction Fee | $1-$50 (variable) | 0.1% (predictable) |
| Smart Contracts | Limited | Planned (CosmWasm) |
| Staking Rewards | No | Yes (~7% APR) |
| Governance | Off-chain | On-chain |
| Purpose | Store of value | Payment + utility |

### vs Ethereum (ETH)
| Feature | Ethereum | VITACOIN |
|---------|----------|----------|
| Consensus | Proof-of-Stake | Proof-of-Stake |
| Transaction Speed | 12-15 seconds | 5 seconds |
| Transaction Fee | $0.50-$100 (gas) | 0.1% (predictable) |
| Smart Contracts | EVM (Solidity) | CosmWasm (Rust) planned |
| Scalability | Layer 2 needed | Built-in fast |
| Primary Use | General smart contracts | Payments first |

### vs Traditional Cosmos Chains
| Feature | Typical Cosmos Chain | VITACOIN |
|---------|---------------------|----------|
| Framework | Cosmos SDK | Cosmos SDK |
| IBC Support | Yes | Yes |
| Focus | General purpose | Payment-focused |
| Ecosystem | Standalone | Integrated with VITAPAY |
| User Experience | Technical | User-friendly (via VITAPAY) |

---

## 📞 Support & Resources

### Documentation
- [Main README](../README.md) - Ecosystem overview
- [VITAPAY Docs](./project/VITAPAY.md) - Payment network
- [Architecture](./architecture/ARCHITECTURE.md) - Technical details
- [Developer Guide](./development/GETTING_STARTED.md) - Start building

### Community
- **Discord**: [Join community](#) (Coming soon)
- **Forum**: [Discuss ideas](#) (Coming soon)
- **Twitter**: [@VitacoinNetwork](#) (Coming soon)

### Developer Support
- **GitHub**: https://github.com/esspron/vitacoin
- **Documentation**: https://docs.vitacoin.network (Coming soon)
- **API Reference**: https://api.vitacoin.network/docs (Coming soon)

### Contact
- **General**: contact@vitacoin.network
- **Security**: security@vitacoin.network
- **Partnerships**: partners@vitacoin.network

---

## ❓ FAQ

**Q: What's the difference between VITACOIN and VITAPAY?**  
A: VITACOIN is the cryptocurrency (the money). VITAPAY is the payment network (the app to use that money).

**Q: How do I buy VITA tokens?**  
A: Initially through token sales. Later on exchanges. Stay tuned for announcements.

**Q: Can I mine VITACOIN?**  
A: No, VITACOIN uses Proof-of-Stake, not Proof-of-Work. You can stake to earn rewards instead.

**Q: What's the minimum to stake?**  
A: No minimum! Stake any amount.

**Q: Is my VITA safe?**  
A: Yes, if you control your private keys. Use the VITAPAY wallet or secure your keys properly.

**Q: Can I run a validator?**  
A: Yes! Anyone can run a validator node. Requires technical knowledge and infrastructure.

**Q: When mainnet launch?**  
A: Target Q2 2026. Follow our channels for updates.

**Q: Is the code open source?**  
A: Yes, fully open source under Apache 2.0 license.

---

**Last Updated**: October 16, 2025  
**Version**: 1.0.0  
**Status**: In Development - Pre-Launch
