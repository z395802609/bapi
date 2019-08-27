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
	if bqueryClis.quiet {
		log.SetOutput(ioutil.Discard)
	} else {
		log.SetOutput(os.Stderr)
	}
	endp.ExtraParams.From = bqueryClis.from
	endp.ExtraParams.Size = bqueryClis.size
	endp.ExtraParams.Format = bqueryClis.format
	endp.ExtraParams.Query = bqueryClis.query
	if endp.ExtraParams.JSON {
		endp.ExtraParams.Format = "json"
	}
	if endp.Status || endp.Projects || endp.Cases || endp.Files || endp.Annotations || endp.Data || endp.Manifest || endp.Slicing {
		fetch.Gdc(endp, bqueryClis.outfn, bqueryClis.retries, bqueryClis.quiet)
		bqueryClis.helpFlags = false
	}
	if bqueryClis.helpFlags {
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
	gdcCmd.Example = `  bquery gdc -p
  bquery gdc -p --json-pretty
  bquery gdc -p -q TARGET-NBL --json-pretty
  bquery gdc -p --format TSV > tcga_projects.tsv
  bquery gdc -p --format CSV > tcga_projects.csv
  bquery gdc -p --from 1 --szie 2
  bquery gdc -s
  bquery gdc -c
  bquery gdc -f
  bquery gdc -a

  // Download manifest for gdc-client
  bquery gdc -m -q "5b2974ad-f932-499b-90a3-93577a9f0573,556e5e3f-0ab9-4b6c-aa62-c42f6a6cf20c" -o my_manifest.txt 
  bquery gdc -m -q "5b2974ad-f932-499b-90a3-93577a9f0573,556e5e3f-0ab9-4b6c-aa62-c42f6a6cf20c" > my_manifest.txt
  bquery gdc -m -q "5b2974ad-f932-499b-90a3-93577a9f0573,556e5e3f-0ab9-4b6c-aa62-c42f6a6cf20c" -n
	
  // Download data
  bquery gdc -d -q "5b2974ad-f932-499b-90a3-93577a9f0573" -n`
}
