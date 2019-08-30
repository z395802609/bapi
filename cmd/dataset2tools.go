package cmd

import (
	"github.com/Miachol/bapi/fetch"
	"github.com/Miachol/bapi/types"
	"github.com/spf13/cobra"
)

var dendp types.Datasets2toolsEndpoints
var dataset2toolsCmd = &cobra.Command{
	Use:   "dts",
	Short: "Query dataset2tools website APIs.",
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

	dataset2toolsCmd.Example = `  bapi dts -a DCA00000060 --json-pretty
  bapi dts -s GSE31106 --json-pretty
  bapi dts --type dataset --json-pretty
  bapi dts -g upregulated --json-pretty`
}
