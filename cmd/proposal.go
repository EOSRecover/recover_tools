package cmd

import (
	"github.com/spf13/cobra"
	"os"
	"recover_tool/app"
	"recover_tool/logger"
)

func init() {

	rootCommand.AddCommand(createProposalCmd())
}

func createProposalCmd() *cobra.Command {

	return &cobra.Command{

		Use:   "proposal",
		Short: "run bank api service",
		Run: func(cmd *cobra.Command, args []string) {

			// init logger
			err := logger.InitLogger("proposal")
			if err != nil {

				os.Exit(1)
				return
			}

			// init eos api
			if err = app.InitAPI(); err != nil {

				logger.Instance().Error("init eos api error -> ", err)
				os.Exit(1)
				return
			}

			// send proposal by read conf
			if err = app.SendProposal(); err != nil {

				logger.Instance().Error("send proposal err -> ", err)
				os.Exit(1)
				return
			}
		},
	}
}
