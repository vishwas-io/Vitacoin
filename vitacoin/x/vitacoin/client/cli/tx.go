package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"cosmossdk.io/math"

	"github.com/vitacoin/vitacoin/vitacoin/x/vitacoin/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdRegisterMerchant(),
		CmdCreatePayment(),
		CmdCompletePayment(),
		CmdRefundPayment(),
	)

	return cmd
}

func CmdRegisterMerchant() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-merchant [business-name] [stake-amount]",
		Short: "Register as a merchant",
		Long:  "Register as a VITAPAY merchant with a business name and initial stake amount (in uvita)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			stakeAmount, ok := math.NewIntFromString(args[1])
			if !ok {
				return fmt.Errorf("invalid stake amount: %s", args[1])
			}

			msg := &types.MsgRegisterMerchant{
				Sender:       clientCtx.GetFromAddress().String(),
				BusinessName: args[0],
				StakeAmount:  stakeAmount,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdCreatePayment() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-payment [merchant-address] [amount] [memo]",
		Short: "Create a payment to a merchant",
		Long:  "Create a VITAPAY payment to a registered merchant. Amount in uvita. Memo is optional.",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			_, err = sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return fmt.Errorf("invalid merchant address: %w", err)
			}

			amount, ok := math.NewIntFromString(args[1])
			if !ok {
				return fmt.Errorf("invalid amount: %s", args[1])
			}

			memo := ""
			if len(args) > 2 {
				memo = strings.Join(args[2:], " ")
			}

			msg := &types.MsgCreatePayment{
				Sender:          clientCtx.GetFromAddress().String(),
				MerchantAddress: args[0],
				Amount:          amount,
				Memo:            memo,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdCompletePayment() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "complete-payment [payment-id]",
		Short: "Complete a pending payment (merchant only)",
		Long:  "Complete a VITAPAY payment. Only the merchant who received the payment can complete it. This triggers fee distribution (40% burn, 40% validators, 20% treasury).",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgCompletePayment{
				Sender:    clientCtx.GetFromAddress().String(),
				PaymentId: args[0],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdRefundPayment() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "refund-payment [payment-id]",
		Short: "Refund a pending payment (merchant only)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgRefundPayment{
				Sender:    clientCtx.GetFromAddress().String(),
				PaymentId: args[0],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
