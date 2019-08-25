package cmd

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/JhuangLab/bquery/parse"
	butils "github.com/JhuangLab/butils"
	"github.com/JhuangLab/butils/log"
	"github.com/biogo/ncbi"
	"github.com/biogo/ncbi/entrez"
	"github.com/spf13/cobra"
)

var clQuery string
var db string
var rettype string
var retmax int
var outfn string
var email string
var retries int
var xml2json bool
var xmlPaths []string
var keywords string
var thread int

var ncbiCmd = &cobra.Command{
	Use:   "ncbi",
	Short: "Query ncbi website APIs.",
	Long:  `Query ncbi website APIs. More see here https://github.com/JhuangLab/bquery.`,
	Run: func(cmd *cobra.Command, args []string) {
		ncbiCmdRunOptions(cmd)
	},
}

func init() {
	ncbiCmd.Flags().StringVarP(&clQuery, "query", "q", "", "Query specifies the search query for record retrieval (required).")
	ncbiCmd.Flags().StringVarP(&db, "db", "d", "pubmed", "Db specifies the database to search")
	ncbiCmd.Flags().StringVarP(&rettype, "rettype", "", "XML", "Rettype specifies the format of the returned data.")
	ncbiCmd.Flags().IntVarP(&retmax, "retmax", "m", 500, "Retmax specifies the number of records to be retrieved per request.")
	ncbiCmd.Flags().StringVarP(&outfn, "outfn", "o", "", "Out specifies destination of the returned data (default to stdout).")
	ncbiCmd.Flags().StringVarP(&email, "email", "e", "your_email@domain.com", "Email specifies the email address to be sent to the server (required).")
	ncbiCmd.Flags().IntVarP(&retries, "retries", "r", 5, "Retry specifies the number of attempts to retrieve the data.")
	ncbiCmd.Flags().BoolVarP(&xml2json, "xml2json", "", false, "Convert XML files to json (Pubmed).")
	ncbiCmd.Flags().IntVarP(&thread, "thread", "t", 2, "Thread to parse XML from local files.")
	ncbiCmd.Flags().StringVarP(&keywords, "keywords", "k", "algorithm, tool, model, pipleline, method, database, workflow, dataset, bioinformatics, sequencing, http, github.com, gitlab.com, bitbucket.org", "Keywords to extracted from abstract.")
	ncbiCmd.Flags().BoolVarP(&quiet, "quiet", "", false, "No log output.")

	ncbiCmd.Example = `  bquery ncbi -d pubmed -q B-ALL -t XML -e your_email@domain.com
  bquery ncbi -q "RNA-seq and bioinformatics[journal]" -e "your_email@domain.com" -m 500 | awk '/<[?]xml version="1.0" [?]>/{close(f); f="abstract.http.XML.tmp" ++c;next} {print>f;}'
  
  k="algorithm, tool, model, pipleline, method, database, workflow, dataset, bioinformatics, sequencing, http, github.com, gitlab.com, bitbucket.org, RNA-Seq, DNA, profile, landscape"
  echo "[" > final.json
  bquery ncbi --xml2json abstract.http.XML.tmp1* -k "${k}"| sed 's/}{/},{/g' >> final.json
  echo "]" >> final.json`
}

func ncbiCmdRunOptions(cmd *cobra.Command) {
	if quiet {
		log.SetOutput(ioutil.Discard)
	} else {
		log.SetOutput(os.Stderr)
	}
	if hasDir, _ := butils.PathExists(outfn); outfn != "" && !hasDir {
		if err := butils.CreateDir(path.Dir(outfn)); err != nil {
			log.FATAL(fmt.Sprintf("Could not to create %s", path.Dir(outfn)))
		}
	}
	if email != "" && clQuery != "" {
		ncbiQuery()
		helpFlags = false
	}
	if xml2json {
		if len(cmd.Flags().Args()) >= 1 {
			xmlPaths = append(xmlPaths, cmd.Flags().Args()...)
			keywordsList := butils.StrSplit(keywords, ", |,", 10000)
			parse.ParsePubmedXML(xmlPaths, outfn, keywordsList, thread)
		}
		helpFlags = false
	}
	if helpFlags {
		cmd.Help()
	}
}

// modified from https://github.com/biogo/ncbi BSD license
func ncbiQuery() {
	ncbi.SetTimeout(0)
	tool := "entrez.example"
	h := entrez.History{}
	s, err := entrez.DoSearch(db, clQuery, nil, &h, tool, email)
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}
	log.Infof("Will retrieve %d records.", s.Count)

	var of *os.File
	if outfn == "" {
		of = os.Stdout
	} else {
		of, err = os.Create(outfn)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		defer of.Close()
	}

	var (
		buf   = &bytes.Buffer{}
		p     = &entrez.Parameters{RetMax: retmax, RetType: rettype, RetMode: "text"}
		bn, n int64
	)
	for p.RetStart = 0; p.RetStart < s.Count; p.RetStart += p.RetMax {
		log.Infof("Attempting to retrieve %d records: %d-%d with %d retries.", p.RetMax, p.RetStart+1, p.RetMax+p.RetStart, retries)
		var t int
		for t = 0; t < retries; t++ {
			buf.Reset()
			var (
				r   io.ReadCloser
				_bn int64
			)
			r, err = entrez.Fetch(db, p, tool, email, &h)
			if err != nil {
				if r != nil {
					r.Close()
				}
				log.Warnf("Failed to retrieve on attempt %d... error: %v ... retrying.", t, err)
				continue
			}
			_bn, err = io.Copy(buf, r)
			bn += _bn
			r.Close()
			if err == nil {
				break
			}
			log.Warnf("Failed to buffer on attempt %d... error: %v ... retrying.", t, err)
		}
		if err != nil {
			os.Exit(1)
		}

		log.Infof("Retrieved records with %d retries... writing out.", t)
		_n, err := io.Copy(of, buf)
		n += _n
		if err != nil {
			log.Fatalf("Error: %v\n", err)
		}
	}
	if bn != n {
		log.Warnf("Writethrough mismatch: %d != %d", bn, n)
	}
}
