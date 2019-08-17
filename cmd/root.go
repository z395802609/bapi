package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/JhuangLab/butils/log"
	"github.com/spf13/cobra"
)

var quiet bool
var helpFlags = true
var version = "v0.1.0"

var rootCmd = &cobra.Command{
	Use:   "bquery",
	Short: "Query bioinformatics website APIs.",
	Long:  `Query bioinformatics website APIs. More see here https://github.com/JhuangLab/bquery.`,
	Run: func(cmd *cobra.Command, args []string) {
		rootCmdRunOptions(cmd)
	},
}

// Execute main interface of bget
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if !rootCmd.HasFlags() && !rootCmd.HasSubCommands() {
			rootCmd.Help()
		} else {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func init() {
	rootCmd.AddCommand(ncbiCmd)
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "", false, "No log output.")

	rootCmd.Example = `  bquery -d pubmed -q B-ALL -t XML -e your_email@domain.com`

	rootCmd.Version = version
}

func rootCmdRunOptions(cmd *cobra.Command) {
	if quiet {
		log.SetOutput(ioutil.Discard)
	} else {
		log.SetOutput(os.Stderr)
	}
	if helpFlags {
		cmd.Help()
	}
}
