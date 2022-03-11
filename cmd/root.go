package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var configPath string

var rootCommand = &cobra.Command{

	Use:   "rec-tool",
	Short: "run recover tool",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	rootCommand.PersistentFlags().StringVar(&configPath, "conf-path", "./conf", "system conf path")
	cobra.OnInitialize(initConfig)

	_ = viper.BindPFlag("conf-path", rootCommand.PersistentFlags().Lookup("conf-path"))
}

func initConfig() {

	viper.AddConfigPath(configPath)
	viper.SetConfigType("yaml")
	viper.SetConfigName("conf")

	if err := viper.ReadInConfig(); err != nil {

		fmt.Println()
		os.Exit(1)
	}
}

func Execute() {

	if err := rootCommand.Execute(); err != nil {

		fmt.Println("启动失败: ", err)
		os.Exit(1)
	}
}
