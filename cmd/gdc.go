package cmd

import (
	"io/ioutil"
	"os"

	"github.com/JhuangLab/bquery/fetch"
	"github.com/JhuangLab/butils/log"
	"github.com/spf13/cobra"
)

var endp fetch.GdcEndpoints

var gdcCmd = &cobra.Command{
	Use:   "gdc",
	Short: "Query GDC portal website APIs.",
	Long:  `Query GDC portal APIs. More see here https://github.com/JhuangLab/bquery.`,
	Run: func(cmd *cobra.Command, args []string) {
		gdcCmdRunOptions(cmd)
	},
}

func gdcCmdRunOptions(cmd *cobra.Command) {
	if quiet {
		log.SetOutput(ioutil.Discard)
	} else {
		log.SetOutput(os.Stderr)
	}
	if endp.Status || endp.Projects || endp.Cases || endp.Files || endp.Annotations {
		fetch.Gdc(endp)
		helpFlags = false
	}
	if helpFlags {
		cmd.Help()
	}
}

func init() {
	gdcCmd.Flags().BoolVarP(&endp.Status, "status", "s", false, "Check GDC portal status (https://portal.gdc.cancer.gov/).")
	gdcCmd.Flags().BoolVarP(&endp.Cases, "cases", "c", false, "Retrive cases info from GDC portal.")
	gdcCmd.Flags().BoolVarP(&endp.Files, "files", "f", false, "Retrive files info from GDC portal.")
	gdcCmd.Flags().BoolVarP(&endp.Projects, "projects", "p", false, "Retrive projects meta info from GDC portal.")
	gdcCmd.Flags().BoolVarP(&endp.Annotations, "annotations", "a", false, "Retrive annotations info from GDC portal.")
	gdcCmd.Flags().IntVarP(&endp.ExtraParams.From, "from", "", -1, "Parameters of API control the start item of retrived data.")
	gdcCmd.Flags().IntVarP(&endp.ExtraParams.Size, "size", "", -1, "Parameters of API control the lenth of retrived data.")
	gdcCmd.Flags().BoolVarP(&endp.ExtraParams.Pretty, "json-pretty", "", false, "Parameters of API pretty retrived data.")
	gdcCmd.Flags().BoolVarP(&endp.Json, "json", "", false, "Parameters of API pretty retrived data.")
	gdcCmd.Flags().BoolVarP(&endp.Legacy, "legacy", "l", false, "Use legacy API of GDC portal.")
}
