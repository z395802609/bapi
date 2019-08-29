package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/Miachol/bapi/fetch"
	"github.com/Miachol/bapi/parse"
	butils "github.com/openbiox/butils"
	"github.com/openbiox/butils/log"
	"github.com/spf13/cobra"
)

var ncbiCmd = &cobra.Command{
	Use:   "ncbi",
	Short: "Query ncbi website APIs.",
	Long:  `Query ncbi website APIs. More see here https://github.com/Miachol/bapi.`,
	Run: func(cmd *cobra.Command, args []string) {
		ncbiCmdRunOptions(cmd)
	},
}

func ncbiCmdRunOptions(cmd *cobra.Command) {
	if bapiClis.quiet {
		log.SetOutput(ioutil.Discard)
	} else {
		log.SetOutput(os.Stderr)
	}
	if bapiClis.format == "" {
		bapiClis.format = "XML"
	}
	if hasDir, _ := butils.PathExists(bapiClis.outfn); bapiClis.outfn != "" && !hasDir {
		if err := butils.CreateDir(path.Dir(bapiClis.outfn)); err != nil {
			log.FATAL(fmt.Sprintf("Could not to create %s", path.Dir(bapiClis.outfn)))
		}
	}
	if bapiClis.email != "" && bapiClis.query != "" {
		fetch.Ncbi(bapiClis.ncbiDB, bapiClis.query, bapiClis.from, bapiClis.size, bapiClis.email, bapiClis.outfn, bapiClis.format, bapiClis.ncbiRetmax, bapiClis.retries, bapiClis.timeout, bapiClis.retSleepTime)
		bapiClis.helpFlags = false
	}
	if bapiClis.ncbiXML2json == "pubmed" {
		if len(cmd.Flags().Args()) >= 1 {
			bapiClis.ncbiXMLPaths = append(bapiClis.ncbiXMLPaths, cmd.Flags().Args()...)
			keywordsList := butils.StrSplit(bapiClis.ncbiKeywords, ", |,", 10000)
			parse.ParsePubmedXML(bapiClis.ncbiXMLPaths, bapiClis.outfn, keywordsList, bapiClis.thread, bapiClis.callCor)
		}
		bapiClis.helpFlags = false
	}
	if bapiClis.helpFlags {
		cmd.Help()
	}
}

func init() {
	ncbiCmd.Flags().StringVarP(&bapiClis.ncbiDB, "db", "d", "pubmed", "Db specifies the database to search")
	ncbiCmd.Flags().IntVarP(&bapiClis.ncbiRetmax, "per-size", "m", 100, "Retmax specifies the number of records to be retrieved per request.")
	ncbiCmd.Flags().StringVarP(&bapiClis.ncbiXML2json, "xml2json", "", "", "Convert XML files to json [e.g. pubmed].")
	ncbiCmd.Flags().IntVarP(&bapiClis.thread, "thread", "t", 2, "Thread to process.")
	ncbiCmd.Flags().StringVarP(&bapiClis.ncbiKeywords, "keywords", "k", "algorithm, tool, model, pipleline, method, database, workflow, dataset, bioinformatics, sequencing, http, github.com, gitlab.com, bitbucket.org", "Keywords to extracted from abstract.")
	ncbiCmd.Flags().BoolVarP(&bapiClis.quiet, "quiet", "", false, "No log output.")
	ncbiCmd.Flags().BoolVarP(&bapiClis.callCor, "call-cor", "", false, "Wheather to calculate the corelated keywords, and return the sentence contains >=2 keywords.")
	ncbiCmd.Flags().StringVarP(&bapiClis.outfn, "outfn", "o", "", "Out specifies destination of the returned data (default to stdout).")

	ncbiCmd.Example = `  bapi ncbi -d pubmed -q B-ALL --format XML -e your_email@domain.com
  bapi ncbi -q "RNA-seq and bioinformatics[journal]" -e "your_email@domain.com" -m 100 | awk '/<[?]xml version="1.0" [?]>/{close(f); f="abstract.http.XML.tmp" ++c;next} {print>f;}'
  
  k="algorithm, tool, model, pipleline, method, database, workflow, dataset, bioinformatics, sequencing, http, github.com, gitlab.com, bitbucket.org, RNA-Seq, DNA, profile, landscape"
  bapi ncbi --xml2json pubmed abstract.http.XML.tmp* -k "${k}" --call-cor | sed 's;}{;,;g' > final.json

  bapi ncbi -q "Galectins control MTOR and AMPK in response to lysosomal damage to induce autophagy OR MTOR-independent autophagy induced by interrupted endoplasmic reticulum-mitochondrial Ca2+ communication: a dead end in cancer cells. OR The PARK10 gene USP24 is a negative regulator of autophagy and ULK1 protein stability OR Coordinate regulation of autophagy and the ubiquitin proteasome system by MTOR." -o titleSearch.XML
  bapi ncbi --xml2json pubmed titleSearch.XML -k "${k}" --call-cor | sed 's;}{;,;g' > final.json
  bapi fmt --json-to-slice final.json
  json2csv -i final.slice.json -o final.csv`
}
