package cmd

import (
	"github.com/Miachol/bapi/fetch"
	"github.com/Miachol/bapi/types"
	"github.com/spf13/cobra"
)

var dendp types.Datasets2toolsEndpoints
var dataset2toolsCmd = &cobra.Command{
	Use:   "dta",
	Short: "Query dataset2tools website APIs: datasets (d), tools (t), and canned analysis (a).",
	Long:  `Query dataset2tools APIs. More see here https://github.com/Miachol/bapi.`,
	Run: func(cmd *cobra.Command, args []string) {
		dataset2toolsCmdRunOptions(cmd)
	},
}

func dataset2toolsCmdRunOptions(cmd *cobra.Command) {
	dendp.Query = bapiClis.Query
	dendp.PageSize = bapiClis.Size
	if dendp.ObjectType != "" || dendp.DatasetAccession != "" ||
		dendp.CannedAnalysisAccession != "" ||
		dendp.ToolName != "" || dendp.DiseaseName != "" || dendp.Gneset != "" {
		fetch.Dataset2tools(&dendp, &bapiClis, &fmtClis)
		bapiClis.HelpFlags = false
	}
	if bapiClis.HelpFlags {
		cmd.Help()
	}
}

func init() {
	dataset2toolsCmd.Flags().StringVarP(&dendp.ObjectType, "type", "", "", "Object type [tool, dataset, canned_analysis].")
	dataset2toolsCmd.Flags().StringVarP(&dendp.ToolName, "tool", "t", "", "Tool name, e.g. bwa.")
	dataset2toolsCmd.Flags().StringVarP(&dendp.DiseaseName, "disease", "d", "", "Disease name, e.g. prostate cancer")
	dataset2toolsCmd.Flags().StringVarP(&dendp.DatasetAccession, "dataset-acc", "s", "", "Dataset accession number, e.g. GSE31106.")
	dataset2toolsCmd.Flags().StringVarP(&dendp.CannedAnalysisAccession, "analysis-acc", "a", "", "Canned analysis accession	, e.g. DCA00000060.")
	dataset2toolsCmd.Flags().StringVarP(&dendp.Gneset, "geneset", "g", "", "With dataset accession, e.g. upregulated.")
	dataset2toolsCmd.Flags().StringVarP(&bapiClis.Outfn, "outfn", "o", "", "Out specifies destination of the returned data (default to stdout).")

	dataset2toolsCmd.Example = `  bapi dta -a DCA00000060 
  bapi dta -s GSE31106 | bapi fmt --json-pretty -
  bapi dta --type dataset | bapi fmt --json-pretty --indent 2 -
  bapi dta -g upregulated | json2csv -o out.csv`
}
