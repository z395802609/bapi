package cmd

import (
	"bufio"
	"io/ioutil"
	"os"

	"github.com/Miachol/bapi/fetch"
	"github.com/Miachol/bapi/parse"
	"github.com/Miachol/bapi/types"
	"github.com/openbiox/butils/log"
	"github.com/openbiox/butils/stringo"
	"github.com/spf13/cobra"
)

var ncbiClis types.NcbiClisT

var ncbiCmd = &cobra.Command{
	Use:   "ncbi",
	Short: "Query ncbi website APIs.",
	Long:  `Query ncbi website APIs. More see here https://github.com/Miachol/bapi.`,
	Run: func(cmd *cobra.Command, args []string) {
		ncbiCmdRunOptions(cmd)
	},
}

func ncbiCmdRunOptions(cmd *cobra.Command) {
	cleanArgs := []string{}
	var stdin []byte
	var err error
	hasStdin := false
	if cleanArgs, hasStdin = checkStdInFlag(cmd); hasStdin {
		reader := bufio.NewReader(os.Stdin)
		stdin, err = ioutil.ReadAll(reader)
		if err != nil {
			log.Fatal(err)
		}
	}
	if bapiClis.Format == "" {
		bapiClis.Format = "XML"
	}
	if bapiClis.Email != "" && bapiClis.Query != "" {
		fetch.Ncbi(&bapiClis, &ncbiClis)
		bapiClis.HelpFlags = false
	}
	if ncbiClis.NcbiXMLToJSON == "pubmed" {
		if len(cleanArgs) >= 1 || len(stdin) > 0 {
			ncbiClis.NcbiXMLPaths = append(ncbiClis.NcbiXMLPaths, cleanArgs...)
			keywordsList := stringo.StrSplit(ncbiClis.NcbiKeywords, ", |,", 10000)
			parse.ParsePubmedXML(ncbiClis.NcbiXMLPaths, stdin, bapiClis.Outfn, keywordsList, bapiClis.Thread, bapiClis.CallCor)
		}
		bapiClis.HelpFlags = false
	}
	if bapiClis.HelpFlags {
		cmd.Help()
	}
}

func init() {
	ncbiCmd.Flags().StringVarP(&ncbiClis.NcbiDB, "db", "d", "pubmed", "Db specifies the database to search")
	ncbiCmd.Flags().IntVarP(&ncbiClis.NcbiRetmax, "per-size", "m", 100, "Retmax specifies the number of records to be retrieved per request.")
	ncbiCmd.Flags().StringVarP(&ncbiClis.NcbiXMLToJSON, "xml2json", "", "", "Convert XML files to json [e.g. pubmed].")
	ncbiCmd.Flags().StringVarP(&ncbiClis.NcbiKeywords, "keywords", "k", "algorithm, tool, model, pipleline, method, database, workflow, dataset, bioinformatics, sequencing, http, github.com, gitlab.com, bitbucket.org", "Keywords to extracted from abstract.")
	ncbiCmd.Flags().IntVarP(&bapiClis.Thread, "thread", "t", 2, "Thread to process.")
	ncbiCmd.Flags().BoolVarP(&bapiClis.Quiet, "quiet", "", false, "No log output.")
	ncbiCmd.Flags().BoolVarP(&bapiClis.CallCor, "call-cor", "", false, "Wheather to calculate the corelated keywords, and return the sentence contains >=2 keywords.")
	ncbiCmd.Flags().StringVarP(&bapiClis.Outfn, "outfn", "o", "", "Out specifies destination of the returned data (default to stdout).")

	ncbiCmd.Example = `  bapi ncbi -d pubmed -q B-ALL --format XML -e your_email@domain.com
  bapi ncbi -q "RNA-seq and bioinformatics[journal]" -e "your_email@domain.com" -m 100 | awk '/<[?]xml version="1.0" [?]>/{close(f); f="abstract.http.XML.tmp" ++c;next} {print>f;}'
  
  k="algorithm, tool, model, pipleline, method, database, workflow, dataset, bioinformatics, sequencing, http, github.com, gitlab.com, bitbucket.org, RNA-Seq, DNA, profile, landscape"
  bapi ncbi --xml2json pubmed abstract.http.XML.tmp* -k "${k}" --call-cor | sed 's;}{;,;g' > final.json

  bapi ncbi -q "Galectins control MTOR and AMPK in response to lysosomal damage to induce autophagy OR MTOR-independent autophagy induced by interrupted endoplasmic reticulum-mitochondrial Ca2+ communication: a dead end in cancer cells. OR The PARK10 gene USP24 is a negative regulator of autophagy and ULK1 protein stability OR Coordinate regulation of autophagy and the ubiquitin proteasome system by MTOR." -o titleSearch.XML
  bapi ncbi --xml2json pubmed titleSearch.XML -k "${k}" --call-cor | sed 's;}{;,;g' | bapi fmt --json-to-slice - > final.json
  json2csv -i final.json -o final.csv`
}
