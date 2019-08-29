package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/openbiox/butils/log"
	"github.com/spf13/cobra"
)

type bapiClisT struct {
	quiet        bool
	helpFlags    bool
	version      string
	ncbiDB       string
	ncbiRetmax   int
	retries      int
	query        string
	format       string
	outfn        string
	email        string
	ncbiXML2json string
	ncbiXMLPaths []string
	ncbiKeywords string
	thread       int
	from         int
	size         int
	remoteName   bool
	timeout      int
	retSleepTime int
	callCor      bool
}

var bapiClis = bapiClisT{}

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
	bapiClis.quiet = false
	bapiClis.helpFlags = true
	bapiClis.version = "v0.1.0"
	rootCmd.AddCommand(ncbiCmd)
	rootCmd.AddCommand(gdcCmd)
	rootCmd.AddCommand(fmtCmd)
	rootCmd.AddCommand(dataset2toolsCmd)
	rootCmd.PersistentFlags().StringVarP(&bapiClis.query, "query", "q", "", "Query specifies the search query for record retrieval (required).")
	rootCmd.PersistentFlags().StringVarP(&bapiClis.format, "format", "", "", "Rettype specifies the format of the returned data (CSV, TSV, JSON for gdc; XML/TEXT for ncbi).")
	rootCmd.PersistentFlags().BoolVarP(&bapiClis.quiet, "quiet", "", false, "No log output.")
	rootCmd.PersistentFlags().IntVarP(&bapiClis.from, "from", "", -1, "Parameters of API control the start item of retrived data.")
	rootCmd.PersistentFlags().IntVarP(&bapiClis.size, "size", "", -1, "Parameters of API control the lenth of retrived data. Default is auto determined.")
	rootCmd.PersistentFlags().StringVarP(&bapiClis.email, "email", "e", "your_email@domain.com", "Email specifies the email address to be sent to the server (NCBI website is required).")
	rootCmd.PersistentFlags().IntVarP(&bapiClis.retries, "retries", "r", 5, "Retry specifies the number of attempts to retrieve the data.")
	rootCmd.PersistentFlags().IntVarP(&bapiClis.timeout, "timeout", "", 35, "Set the timeout of per request.")
	rootCmd.PersistentFlags().IntVarP(&bapiClis.retSleepTime, "retries-sleep-time", "", 5, "Sleep time after one retry.")

	rootCmd.Version = bapiClis.version
}

func rootCmdRunOptions(cmd *cobra.Command) {
	if bapiClis.quiet {
		log.SetOutput(ioutil.Discard)
	} else {
		log.SetOutput(os.Stderr)
	}
	if bapiClis.helpFlags {
		cmd.Help()
	}
}
