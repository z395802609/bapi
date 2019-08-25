package cmd

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/JhuangLab/butils/log"
	butils "github.com/JhuangLab/butils"
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
	ncbiCmd.Flags().StringVarP(&rettype, "rettype", "t", "XML", "Rettype specifies the format of the returned data.")
	ncbiCmd.Flags().IntVarP(&retmax, "retmax", "m", 500, "Retmax specifies the number of records to be retrieved per request.")
	ncbiCmd.Flags().StringVarP(&outfn, "outfn", "o", "", "Out specifies destination of the returned data (default to stdout).")
	ncbiCmd.Flags().StringVarP(&email, "email", "e", "", "Email specifies the email address to be sent to the server (required).")
	ncbiCmd.Flags().IntVarP(&retries, "retries", "r", 5, "Retry specifies the number of attempts to retrieve the data.")
	ncbiCmd.Flags().BoolVarP(&quiet, "quiet", "", false, "No log output.")

	ncbiCmd.Example = `  bquery ncbi -d pubmed -q B-ALL -t XML -e your_email@domain.com`
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
			io.Copy(buf, io.Reader(strings.NewReader("\n")))
			bn += _bn + 1
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