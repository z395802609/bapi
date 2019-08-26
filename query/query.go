package query

import (
	"bytes"
	"io"
	"os"

	"github.com/JhuangLab/butils/log"
	"github.com/biogo/ncbi"
	"github.com/biogo/ncbi/entrez"
)

// Ncbi modified from https://github.com/biogo/ncbi BSD license
func Ncbi(db string, clQuery string, start int, end int, email string, outfn string, rettype string, retmax int, retries int) {
	ncbi.SetTimeout(0)
	tool := "entrez.example"
	h := entrez.History{}
	parms := entrez.Parameters{
		APIKey: "193124979d2e7f360c150dadc5b1e3bfec09",
	}
	s, err := entrez.DoSearch(db, clQuery, &parms, &h, tool, email)
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}
	log.Infof("Available retrieve %d records.", s.Count)
	if end == -1 || end > s.Count {
		end = s.Count
	}
	if start < 1 {
		start = 1
	} else if start > s.Count {
		start = s.Count
	}
	if end < start {
		end = start
	}
	log.Infof("Will retrieve %d records, from %d to %d.", end-start+1, start, end)

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
	if p.RetMax > end-start {
		p.RetMax = end - start + 1
	}
	for p.RetStart = start - 1; p.RetStart < end; p.RetStart += p.RetMax {
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
			r.Close()
			if err == nil {
				bn += _bn
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
