package parse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	// "bytes"

	"strings"

	// "reflect"

	// "github.com/gocolly/colly"
	"github.com/PuerkitoBio/goquery"
)

type PubmedFields struct {
	Pmid, Doi, Title, Abs, Journal, Issue, Volume, Date, Issn string
	Keywords                                                  []string
}

// ParsePubmedXML convert Pubmed XML to json
func ParsePubmedXML(xmlPaths []string, outfn string, keywords []string, thread int) {
	sem := make(chan bool, thread)

	//|os.O_APPEND
	var of *os.File
	if outfn == "" {
		of = os.Stdout
	} else {
		of, err := os.OpenFile(outfn, os.O_CREATE|os.O_WRONLY, 0664)
		if err != nil {
			log.Fatal(err)
		}
		defer of.Close()
	}

	var buf = &bytes.Buffer{}
	for _, xmlPath := range xmlPaths {
		sem <- true
		go func(xmlPath string) {
			defer func() {
				<-sem
			}()
			xml, err := os.Open(xmlPath)
			if err != nil {
				log.Fatal(err)
			}
			defer xml.Close()
			htmlDoc, err := goquery.NewDocumentFromReader(xml)
			if err != nil {
				log.Fatal(err)
			}
			htmlDoc.Find("PubmedArticle").Each(func(i int, s *goquery.Selection) {
				json := getPubmedFields(keywords, s)
				io.Copy(buf, bytes.NewBuffer(json))
				// fmt.Printf("%s, %s\n%s\n%v\n", pmid, doi, abs, key)
			})
		}(xmlPath)
	}
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
	if _, err := io.Copy(of, buf); err != nil {
		log.Fatal(err)
	}
}

func getPubmedFields(keywords []string, s *goquery.Selection) []byte {
	year := s.Find("PubmedArticle MedlineCitation Article Journal JournalIssue PubDate > Year").Text()
	month := s.Find("PubmedArticle MedlineCitation Article Journal JournalIssue PubDate > Month").Text()
	day := s.Find("PubmedArticle MedlineCitation Article Journal JournalIssue PubDate > Day").Text()
	date := fmt.Sprintf("%s %s %s", year, month, day)
	issue := s.Find("PubmedArticle MedlineCitation Article Journal JournalIssue Issue").Text()
	volume := s.Find("PubmedArticle MedlineCitation Article Journal JournalIssue Volume").Text()
	journal := s.Find("PubmedArticle MedlineCitation Article Journal ISOAbbreviation").Text()
	issn := s.Find("PubmedArticle MedlineCitation Article Journal ISSN").Text()
	pmid := s.Find("PubmedArticle PubmedData > ArticleIdList > ArticleId[IdType=pubmed]").Text()
	doi := s.Find("PubmedArticle PubmedData > ArticleIdList > ArticleId[IdType=doi]").Text()
	abs := s.Find("PubmedArticle MedlineCitation Article AbstractText").Text()
	title := s.Find("PubmedArticle MedlineCitation Article ArticleTitle").Text()

	var key []string
	for _, item := range keywords {
		if strings.Contains(title, item) || strings.Contains(abs, item) {
			key = append(key, item)
		}
	}
	json, _ := json.MarshalIndent(
		PubmedFields{
			Pmid:     pmid,
			Doi:      doi,
			Title:    title,
			Abs:      abs,
			Journal:  journal,
			Issn:     issn,
			Date:     date,
			Issue:    issue,
			Volume:   volume,
			Keywords: key,
		}, "", "   ")
	return json
}
