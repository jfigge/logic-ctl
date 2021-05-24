package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.td.teradata.com/sandbox/logic-ctl/internal/config"
	"github.td.teradata.com/sandbox/logic-ctl/internal/driver"
	"log"
	"os"
)

var cfgFile string
var romFile string

var rootCmd = &cobra.Command{
	Use:   "logic",
	Short: "logic is logic 1 breadboard cpu driver",
	RunE: func(cmd *cobra.Command, args []string) error {

		// Load 6502 rom
		if config.CLIConfig.RomFile == "" {
			fmt.Printf("No rom specified.  Use -r/--rom <file> to specify")
			os.Exit(1)
		}

		d := driver.New()
		d.Run()
		return nil
	},
}

// Execute bootstraps the viper
func Execute() error {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "configuration file for logic")
	rootCmd.PersistentFlags().StringVarP(&romFile, "rom",    "r", "", "rom file for logic simulation")
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {

	if err := initConfigE(); err != nil {
		log.Fatalf("Failed to load configuration: %s", err)
		return
	}
}

func initConfigE() error {
	defer func() {
		config.CLIConfig.RomFile = romFile
	}()
	return config.NewConfig(cfgFile)
}