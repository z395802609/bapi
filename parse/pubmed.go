package parse

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/openbiox/butils"
	"github.com/openbiox/butils/log"
	"github.com/PuerkitoBio/goquery"
	jsoniter "github.com/json-iterator/go"
	prose "gopkg.in/jdkato/prose.v2"
	xurls "mvdan.cc/xurls/v2"
)

type PubmedFields struct {
	Pmid, Doi, Title, Abs, Journal, Issue, Volume, Date, Issn string
	Corelations                                               map[string]string
	URLs                                                      []string
	Keywords                                                  []string
}

// ParsePubmedXML convert Pubmed XML to json
func ParsePubmedXML(xmlPaths []string, outfn string, keywords []string, thread int) {
	if len(xmlPaths) == 1 {
		thread = 1
	}
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

func getPubmedFields(keywords []string, s *goquery.Selection) (jsonData []byte) {
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
	abs := s.Find("PubmedArticle MedlineCitation Article Abstract").Text()
	abs = butils.StrReplaceAll(abs, "\n  *", "\n")
	abs = butils.StrReplaceAll(abs, "(<[/]AbstractText.*>)|(^\n)|(\n$)", "")
	title := s.Find("PubmedArticle MedlineCitation Article ArticleTitle").Text()
	titleAbs := title + "\n" + abs
	urls := xurls.Relaxed().FindAllString(titleAbs, -1)
	keywordsPat := strings.Join(keywords, "|")
	key := butils.StrExtract(titleAbs, keywordsPat, 1000000)
	key = butils.RemoveRepeatEle(key)

	doc, err := prose.NewDocument(titleAbs)
	corela := make(map[string]string)
	if len(key) > 2 {
		getKeywordsCorleations(doc, keywordsPat, &corela)
	}
	if err != nil {
		log.Warn(err)
	} else {

	}
	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	jsonData, _ = json.MarshalIndent(
		PubmedFields{
			Pmid:        pmid,
			Doi:         doi,
			Title:       title,
			Abs:         abs,
			Journal:     journal,
			Issn:        issn,
			Date:        date,
			Issue:       issue,
			Volume:      volume,
			Corelations: corela,
			URLs:        urls,
			Keywords:    key,
		}, "", "   ")
	return jsonData
}

func getKeywordsCorleations(doc *prose.Document, keywordsPat string, corela *map[string]string) {
	for _, sent := range doc.Sentences() {
		kStr := butils.StrExtract(sent.Text, keywordsPat, 1000000)
		kStr = butils.RemoveRepeatEle(kStr)
		if len(kStr) >= 2 {
			(*corela)[strings.Join(kStr, "+")] = sent.Text
		}
	}
}
