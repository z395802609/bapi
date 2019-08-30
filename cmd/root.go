package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/Miachol/bapi/types"
	cio "github.com/openbiox/butils/io"
	"github.com/openbiox/butils/log"
	"github.com/spf13/cobra"
)

var bapiClis = types.BapiClisT{}

var rootCmd = &cobra.Command{
	Use:   "bapi",
	Short: "Query bioinformatics website APIs.",
	Long:  `Query bioinformatics website APIs. More see here https://github.com/Miachol/bapi.`,
	Run: func(cmd *cobra.Command, args []string) {
		rootCmdRunOptions(cmd)
	},
}

// Execute main interface of butils
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
	bapiClis.Quiet = false
	bapiClis.HelpFlags = true
	bapiClis.Version = "v0.1.0"
	rootCmd.AddCommand(ncbiCmd)
	rootCmd.AddCommand(gdcCmd)
	rootCmd.AddCommand(fmtCmd)
	rootCmd.AddCommand(dataset2toolsCmd)
	rootCmd.PersistentFlags().StringVarP(&bapiClis.Query, "query", "q", "", "Query specifies the search query for record retrieval (required).")
	rootCmd.PersistentFlags().StringVarP(&bapiClis.Format, "format", "", "", "Rettype specifies the format of the returned data (CSV, TSV, JSON for gdc; XML/TEXT for ncbi).")
	rootCmd.PersistentFlags().BoolVarP(&bapiClis.Quiet, "quiet", "", false, "No log output.")
	rootCmd.PersistentFlags().IntVarP(&bapiClis.From, "from", "", -1, "Parameters of API control the start item of retrived data.")
	rootCmd.PersistentFlags().IntVarP(&bapiClis.Size, "size", "", -1, "Parameters of API control the lenth of retrived data. Default is auto determined.")
	rootCmd.PersistentFlags().StringVarP(&bapiClis.Email, "email", "e", "your_email@domain.com", "Email specifies the email address to be sent to the server (NCBI website is required).")
	rootCmd.PersistentFlags().IntVarP(&bapiClis.Retries, "retries", "r", 5, "Retry specifies the number of attempts to retrieve the data.")
	rootCmd.PersistentFlags().IntVarP(&bapiClis.Timeout, "timeout", "", 35, "Set the timeout of per request.")
	rootCmd.PersistentFlags().IntVarP(&bapiClis.RetSleepTime, "retries-sleep-time", "", 5, "Sleep time after one retry.")

	rootCmd.Version = bapiClis.Version
}

func rootCmdRunOptions(cmd *cobra.Command) {
	if bapiClis.Quiet {
		log.SetOutput(ioutil.Discard)
	} else {
		log.SetOutput(os.Stderr)
	}
	if hasDir, _ := cio.PathExists(bapiClis.Outfn); bapiClis.Outfn != "" && !hasDir {
		if err := cio.CreateDir(path.Dir(bapiClis.Outfn)); err != nil {
			log.FATAL(fmt.Sprintf("Could not to create %s", path.Dir(bapiClis.Outfn)))
		}
	}
	if bapiClis.HelpFlags {
		cmd.Help()
	}
}
