package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/JhuangLab/bquery/fetch"
	"github.com/JhuangLab/bquery/parse"
	butils "github.com/JhuangLab/butils"
	"github.com/JhuangLab/butils/log"
	"github.com/spf13/cobra"
)

var ncbiCmd = &cobra.Command{
	Use:   "ncbi",
	Short: "Query ncbi website APIs.",
	Long:  `Query ncbi website APIs. More see here https://github.com/JhuangLab/bquery.`,
	Run: func(cmd *cobra.Command, args []string) {
		ncbiCmdRunOptions(cmd)
	},
}

func ncbiCmdRunOptions(cmd *cobra.Command) {
	if bqueryClis.quiet {
		log.SetOutput(ioutil.Discard)
	} else {
		log.SetOutput(os.Stderr)
	}
	if bqueryClis.format == "" {
		bqueryClis.format = "XML"
	}
	if hasDir, _ := butils.PathExists(bqueryClis.outfn); bqueryClis.outfn != "" && !hasDir {
		if err := butils.CreateDir(path.Dir(bqueryClis.outfn)); err != nil {
			log.FATAL(fmt.Sprintf("Could not to create %s", path.Dir(bqueryClis.outfn)))
		}
	}
	if bqueryClis.email != "" && bqueryClis.ncbiclQuery != "" {
		fetch.Ncbi(bqueryClis.ncbiDB, bqueryClis.ncbiclQuery, bqueryClis.from, bqueryClis.size, bqueryClis.email, bqueryClis.outfn, bqueryClis.format, bqueryClis.ncbiRetmax, bqueryClis.retries)
		bqueryClis.helpFlags = false
	}
	if bqueryClis.ncbiXML2json == "pubmed" {
		if len(cmd.Flags().Args()) >= 1 {
			bqueryClis.ncbiXMLPaths = append(bqueryClis.ncbiXMLPaths, cmd.Flags().Args()...)
			keywordsList := butils.StrSplit(bqueryClis.ncbiKeywords, ", |,", 10000)
			parse.ParsePubmedXML(bqueryClis.ncbiXMLPaths, bqueryClis.outfn, keywordsList, bqueryClis.ncbiThread)
		}
		bqueryClis.helpFlags = false
	}
	if bqueryClis.helpFlags {
		cmd.Help()
	}
}

func init() {
	ncbiCmd.Flags().StringVarP(&bqueryClis.ncbiclQuery, "query", "q", "", "Query specifies the search query for record retrieval (required).")
	ncbiCmd.Flags().StringVarP(&bqueryClis.ncbiDB, "db", "d", "pubmed", "Db specifies the database to search")
	ncbiCmd.Flags().IntVarP(&bqueryClis.ncbiRetmax, "per-size", "m", 100, "Retmax specifies the number of records to be retrieved per request.")
	ncbiCmd.Flags().StringVarP(&bqueryClis.ncbiXML2json, "xml2json", "", "", "Convert XML files to json [e.g. pubmed].")
	ncbiCmd.Flags().IntVarP(&bqueryClis.ncbiThread, "thread", "t", 2, "Thread to parse XML from local files.")
	ncbiCmd.Flags().StringVarP(&bqueryClis.ncbiKeywords, "keywords", "k", "algorithm, tool, model, pipleline, method, database, workflow, dataset, bioinformatics, sequencing, http, github.com, gitlab.com, bitbucket.org", "Keywords to extracted from abstract.")
	ncbiCmd.Flags().BoolVarP(&bqueryClis.quiet, "quiet", "", false, "No log output.")

	ncbiCmd.Example = `  bquery ncbi -d pubmed -q B-ALL --format XML -e your_email@domain.com
  bquery ncbi -q "RNA-seq and bioinformatics[journal]" -e "your_email@domain.com" -m 100 | awk '/<[?]xml version="1.0" [?]>/{close(f); f="abstract.http.XML.tmp" ++c;next} {print>f;}'
  
  k="algorithm, tool, model, pipleline, method, database, workflow, dataset, bioinformatics, sequencing, http, github.com, gitlab.com, bitbucket.org, RNA-Seq, DNA, profile, landscape"
  echo "[" > final.json
  bquery ncbi --xml2json pubmed abstract.http.XML.tmp* -k "${k}"| sed 's/}{/},{/g' >> final.json
  echo "]" >> final.json`
}
