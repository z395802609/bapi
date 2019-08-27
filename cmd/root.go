package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/JhuangLab/butils/log"
	"github.com/spf13/cobra"
)

type bqueryClisT struct {
	quiet        bool
	helpFlags    bool
	version      string
	ncbiclQuery  string
	ncbiDB       string
	ncbiRetmax   int
	retries      int
	format       string
	outfn        string
	email        string
	ncbiXML2json string
	ncbiXMLPaths []string
	ncbiKeywords string
	ncbiThread   int
	from         int
	size         int
}

var bqueryClis = bqueryClisT{}

var rootCmd = &cobra.Command{
	Use:   "bquery",
	Short: "Query bioinformatics website APIs.",
	Long:  `Query bioinformatics website APIs. More see here https://github.com/JhuangLab/bquery.`,
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
	bqueryClis.quiet = false
	bqueryClis.helpFlags = true
	bqueryClis.version = "v0.1.0"
	rootCmd.AddCommand(ncbiCmd)
	rootCmd.AddCommand(gdcCmd)
	rootCmd.PersistentFlags().StringVarP(&bqueryClis.format, "format", "", "", "Rettype specifies the format of the returned data (CSV, TSV, JSON for gdc; XML/TEXT for ncbi).")
	rootCmd.PersistentFlags().StringVarP(&bqueryClis.outfn, "outfn", "o", "", "Out specifies destination of the returned data (default to stdout).")
	rootCmd.PersistentFlags().BoolVarP(&bqueryClis.quiet, "quiet", "", false, "No log output.")
	rootCmd.PersistentFlags().IntVarP(&bqueryClis.from, "from", "", 0, "Parameters of API control the start item of retrived data.")
	rootCmd.PersistentFlags().IntVarP(&bqueryClis.size, "size", "", -1, "Parameters of API control the lenth of retrived data. Default is auto determined.")
	rootCmd.PersistentFlags().StringVarP(&bqueryClis.email, "email", "e", "your_email@domain.com", "Email specifies the email address to be sent to the server (NCBI website is required).")
	rootCmd.PersistentFlags().IntVarP(&bqueryClis.retries, "retries", "r", 5, "Retry specifies the number of attempts to retrieve the data.")

	rootCmd.Version = bqueryClis.version
}

func rootCmdRunOptions(cmd *cobra.Command) {
	if bqueryClis.quiet {
		log.SetOutput(ioutil.Discard)
	} else {
		log.SetOutput(os.Stderr)
	}
	if bqueryClis.helpFlags {
		cmd.Help()
	}
}
