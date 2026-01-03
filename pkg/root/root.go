package root

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "laravel",
	Short: "Laravel Go Services CLI",
	Long:  `A centralized command line tool for high-performance Laravel background services.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func GetRoot() *cobra.Command {
	return rootCmd
}

// SetInfo allows the user to configure the root command's metadata.
func SetInfo(use, short, long string) {
	rootCmd.Use = use
	rootCmd.Short = short
	rootCmd.Long = long
}
