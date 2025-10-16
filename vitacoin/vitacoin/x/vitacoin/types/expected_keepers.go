package types

type ExpectedKeepers struct {
    AccountKeeper AccountKeeper
    BankKeeper    BankKeeper
    StakingKeeper StakingKeeper
}

type AccountKeeper interface {
    // Define methods for account management
}

type BankKeeper interface {
    // Define methods for bank operations
}

type StakingKeeper interface {
    // Define methods for staking operations
}