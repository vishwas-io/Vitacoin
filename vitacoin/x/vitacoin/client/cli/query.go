package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin/types"
)

// GetQueryCmd returns the CLI query commands for this module
func GetQueryCmd() *cobra.Command {
	vitacoinQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the vitacoin module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	vitacoinQueryCmd.AddCommand(
		GetCmdQueryParams(),
		GetCmdQueryMerchant(),
		GetCmdQueryMerchantAll(),
		GetCmdQueryPayment(),
		GetCmdQueryPaymentAll(),
		GetCmdQueryVault(),
		GetCmdQueryVaultAll(),
		GetCmdQueryRewardPool(),
		GetCmdQueryRewardPoolAll(),
		// Phase 3: Fee & Economics Queries
		GetCmdQueryFeeStatistics(),
		GetCmdQueryBurnStatistics(),
		GetCmdQuerySupplySnapshot(),
		GetCmdQuerySupplySnapshotLatest(),
		GetCmdQueryFeeAccumulator(),
	)

	return vitacoinQueryCmd
}

// GetCmdQueryParams implements the params query command
func GetCmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the current vitacoin module parameters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryMerchant implements the merchant query command
func GetCmdQueryMerchant() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "merchant [address]",
		Short: "Query a merchant by address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Merchant(context.Background(), &types.QueryMerchantRequest{
				Address: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryMerchantAll implements the merchant-all query command
func GetCmdQueryMerchantAll() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "merchants",
		Short: "Query all merchants",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.MerchantAll(context.Background(), &types.QueryAllMerchantRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryPayment implements the payment query command
func GetCmdQueryPayment() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "payment [id]",
		Short: "Query a payment by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Payment(context.Background(), &types.QueryPaymentRequest{
				Id: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryPaymentAll implements the payment-all query command
func GetCmdQueryPaymentAll() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "payments",
		Short: "Query all payments",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.PaymentAll(context.Background(), &types.QueryAllPaymentRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryVault implements the vault query command
func GetCmdQueryVault() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vault [id]",
		Short: "Query a vault by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Vault(context.Background(), &types.QueryVaultRequest{
				Id: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryVaultAll implements the vault-all query command
func GetCmdQueryVaultAll() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vaults",
		Short: "Query all vaults",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.VaultAll(context.Background(), &types.QueryAllVaultRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryRewardPool implements the reward-pool query command
func GetCmdQueryRewardPool() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pool [id]",
		Short: "Query a reward pool by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.RewardPool(context.Background(), &types.QueryRewardPoolRequest{
				Id: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryRewardPoolAll implements the reward-pool-all query command
func GetCmdQueryRewardPoolAll() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pools",
		Short: "Query all reward pools",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.RewardPoolAll(context.Background(), &types.QueryAllRewardPoolRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// Phase 3: Fee & Economics Query Commands

// GetCmdQueryFeeStatistics implements the fee-statistics query command
func GetCmdQueryFeeStatistics() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fee-statistics",
		Short: "Query cumulative fee statistics",
		Long: `Query cumulative fee statistics since genesis including:
- Total fees collected
- Total burned
- Total to validators
- Total to treasury
- Transaction count`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.FeeStatistics(context.Background(), &types.QueryFeeStatisticsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryBurnStatistics implements the burn-statistics query command
func GetCmdQueryBurnStatistics() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "burn-statistics",
		Short: "Query burn mechanism statistics",
		Long: `Query burn mechanism statistics including:
- Total burned
- Burn rate per day
- Current supply
- Burn cap supply
- Remaining to cap
- Burn cap reached status`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.BurnStatistics(context.Background(), &types.QueryBurnStatisticsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQuerySupplySnapshot implements the supply-snapshot query command
func GetCmdQuerySupplySnapshot() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "supply-snapshot [height]",
		Short: "Query supply snapshot at specific height",
		Long: `Query supply snapshot for a specific block height including:
- Total supply
- Circulating supply
- Liquid supply
- Bonded supply
- Burned cumulative`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			height, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid height: %w", err)
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.SupplySnapshot(context.Background(), &types.QuerySupplySnapshotRequest{
				Height: height,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQuerySupplySnapshotLatest implements the supply-snapshot-latest query command
func GetCmdQuerySupplySnapshotLatest() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "supply-snapshot-latest",
		Short: "Query the most recent supply snapshot",
		Long: `Query the most recent supply snapshot including:
- Total supply
- Circulating supply
- Liquid supply
- Bonded supply
- Burned cumulative`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.SupplySnapshotLatest(context.Background(), &types.QuerySupplySnapshotLatestRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryFeeAccumulator implements the fee-accumulator query command
func GetCmdQueryFeeAccumulator() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fee-accumulator",
		Short: "Query the current block's fee accumulator",
		Long: `Query the current block's fee accumulator showing:
- Current block height
- Total fees collected in current block
- Transaction count in current block`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.FeeAccumulator(context.Background(), &types.QueryFeeAccumulatorRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}