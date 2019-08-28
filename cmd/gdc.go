package cmd

import (
	"io/ioutil"
	"os"

	"github.com/Miachol/bapi/fetch"
	"github.com/openbiox/butils/log"
	"github.com/spf13/cobra"
)

var endp fetch.GdcEndpoints

var gdcCmd = &cobra.Command{
	Use:   "gdc",
	Short: "Query GDC portal website APIs.",
	Long:  `Query GDC portal APIs. More see here https://github.com/Miachol/bapi.`,
	Run: func(cmd *cobra.Command, args []string) {
		gdcCmdRunOptions(cmd)
	},
}

func gdcCmdRunOptions(cmd *cobra.Command) {
	if bapiClis.quiet {
		log.SetOutput(ioutil.Discard)
	} else {
		log.SetOutput(os.Stderr)
	}
	endp.ExtraParams.From = bapiClis.from
	endp.ExtraParams.Size = bapiClis.size
	endp.ExtraParams.Format = bapiClis.format
	endp.ExtraParams.Query = bapiClis.query
	if endp.ExtraParams.JSON {
		endp.ExtraParams.Format = "json"
	}
	if endp.Status || endp.Projects || endp.Cases || endp.Files || endp.Annotations || endp.Data || endp.Manifest || endp.Slicing {
		fetch.Gdc(endp, bapiClis.outfn, bapiClis.retries, bapiClis.timeout, bapiClis.retSleepTime, bapiClis.quiet)
		bapiClis.helpFlags = false
	}
	if bapiClis.helpFlags {
		cmd.Help()
	}
}

func init() {
	gdcCmd.Flags().BoolVarP(&endp.ExtraParams.RemoteName, "remote-name", "n", false, "Use remote defined filename.")
	gdcCmd.Flags().BoolVarP(&endp.Status, "status", "s", false, "Check GDC portal status (https://portal.gdc.cancer.gov/).")
	gdcCmd.Flags().BoolVarP(&endp.Cases, "cases", "c", false, "Retrive cases info from GDC portal.")
	gdcCmd.Flags().BoolVarP(&endp.Files, "files", "f", false, "Retrive files info from GDC portal.")
	gdcCmd.Flags().BoolVarP(&endp.Projects, "projects", "p", false, "Retrive projects meta info from GDC portal.")
	gdcCmd.Flags().BoolVarP(&endp.Annotations, "annotations", "a", false, "Retrive annotations info from GDC portal.")
	gdcCmd.Flags().BoolVarP(&endp.Data, "data", "d", false, "Retrive /data from GDC portal.")
	gdcCmd.Flags().BoolVarP(&endp.Manifest, "manifest", "m", false, "Retrive /manifest data from GDC portal.")
	gdcCmd.Flags().BoolVarP(&endp.Slicing, "slicing", "", false, "Retrive BAM slicing from GDC portal.")
	gdcCmd.Flags().BoolVarP(&endp.ExtraParams.Pretty, "json-pretty", "", false, "Retrive pretty JSON data.")
	gdcCmd.Flags().BoolVarP(&endp.ExtraParams.JSON, "json", "", false, "Retrive JSON data.")
	gdcCmd.Flags().StringVarP(&endp.ExtraParams.Filter, "filter", "", "", "Retrive data with GDC filter.")
	gdcCmd.Flags().BoolVarP(&endp.Legacy, "legacy", "l", false, "Use legacy API of GDC portal.")
	gdcCmd.Flags().StringVarP(&endp.ExtraParams.Token, "token", "", "", "Token to access GDC.")
	gdcCmd.Flags().StringVarP(&endp.ExtraParams.Sort, "sort", "", "", "Sort parameters.")
	gdcCmd.Flags().StringVarP(&endp.ExtraParams.Fields, "fields", "", "", "Fields parameters.")
	gdcCmd.Flags().StringVarP(&bapiClis.outfn, "outfn", "o", "", "Out specifies destination of the returned data (default to stdout), for gdc and ncbi module.")
	gdcCmd.Example = `  bapi gdc -p
  bapi gdc -p --json-pretty
  bapi gdc -p -q TARGET-NBL --json-pretty
  bapi gdc -p --format TSV > tcga_projects.tsv
  bapi gdc -p --format CSV > tcga_projects.csv
  bapi gdc -p --from 1 --szie 2
  bapi gdc -s
  bapi gdc -c
  bapi gdc -f
  bapi gdc -a

  // Download manifest for gdc-client
  bapi gdc -m -q "5b2974ad-f932-499b-90a3-93577a9f0573,556e5e3f-0ab9-4b6c-aa62-c42f6a6cf20c" -o my_manifest.txt 
  bapi gdc -m -q "5b2974ad-f932-499b-90a3-93577a9f0573,556e5e3f-0ab9-4b6c-aa62-c42f6a6cf20c" > my_manifest.txt
  bapi gdc -m -q "5b2974ad-f932-499b-90a3-93577a9f0573,556e5e3f-0ab9-4b6c-aa62-c42f6a6cf20c" -n
	
  // Download data
  bapi gdc -d -q "5b2974ad-f932-499b-90a3-93577a9f0573" -n`
}
