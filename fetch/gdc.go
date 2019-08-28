package fetch

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry/bytefmt"
	"github.com/olekukonko/tablewriter"
	mpb "github.com/vbauerster/mpb/v4"

	"github.com/openbiox/butils/log"
)

const gdcAPIHost = "https://api.gdc.cancer.gov"
const gdcAPIHostLegacy = "https://api.gdc.cancer.gov/legacy"

var endpoints = []string{"status", "projects", "cases", "files", "annotations",
	"data", "manifest", "slicing"}

var tables []*tablewriter.Table
var pg *mpb.Progress

// Gdc accesss https://api.gdc.cancer.gov data
func Gdc(endpoint GdcEndpoints, outfn string, retries int, timeout int, retSleepTime int, quite bool) {
	client := &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: time.Duration(timeout) * time.Second,
			}).Dial,
		},
	}
	host := gdcAPIHost
	if endpoint.Legacy {
		host = gdcAPIHostLegacy
	}
	v := reflect.ValueOf(endpoint)
	count := v.NumField()
	var req *http.Request
	var queryFlag string
	for i := 0; i < count; i++ {
		if i > len(endpoints) {
			continue
		}
		queryFlag = ""
		f := v.Field(i)
		if f.Kind() == reflect.String && f.String() != "" {
			req, queryFlag = setGdcReq(host, i, &endpoint)
		} else if f.Kind() == reflect.Bool && f.Bool() {
			req, queryFlag = setGdcReq(host, i, &endpoint)
		}
		if queryFlag == "" {
			continue
		}
		log.Infof("Query GDC portal %s API: %s.", queryFlag, req.URL.String())
		resp, err := client.Do(req)
		if resp.StatusCode != http.StatusOK {
			log.Warnf("Access failed: %s", host+"/"+endpoints[i])
			fmt.Println("")
			return
		}
		contentDis := resp.Header.Get("Content-Disposition")
		if outfn == "" && contentDis != "" && endpoint.ExtraParams.RemoteName &&
			strings.Contains(contentDis, "filename") {
			_, params, err := mime.ParseMediaType(contentDis)
			if err != nil {
				log.Warn(err)
			} else {
				outfn = params["filename"]
			}
		}

		defer resp.Body.Close()

		var of *os.File
		if outfn == "" {
			of = os.Stdout
		} else {
			of, err := os.OpenFile(outfn, os.O_CREATE|os.O_WRONLY, 0664)
			if err != nil {
				log.Fatalf("error: %v", err)
			}
			defer of.Close()
			wd, _ := os.Getwd()
			log.Infof("Trying %s => %s", req.URL.String(), path.Join(wd, outfn))
		}
		if outfn != "" && endpoint.Data || endpoint.Slicing {
			err = HTTPDownload(req.URL.String(), outfn, pg, quite, false, retries, timeout, retSleepTime)
			if err != nil {
				log.Warn(err)
			}
			return
		}
		if endpoint.ExtraParams.Format != "" || endpoint.ExtraParams.Pretty || endpoint.ExtraParams.Query != "" {
			_, err := io.Copy(of, resp.Body)
			if err != nil {
				log.Warn(err)
			}
			return
		}
		postGdcQuery(&queryFlag, resp, &endpoint, of)
	}
}

func setGdcReq(host string, i int, endpoint *GdcEndpoints) (*http.Request, string) {
	queryFlag := endpoints[i]
	suffix := setGdcQuerySuffix(queryFlag, endpoint)
	method := "GET"
	req, err := http.NewRequest(method, host+"/"+endpoints[i]+suffix, nil)
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/56.0.2924.87 Safari/537.36")
	if err != nil {
		log.Warn(err)
	}
	return req, queryFlag
}

func setGdcQuerySuffix(queryFlag string, endpoint *GdcEndpoints) (suffix string) {
	queryStr := ""
	suffixList := []string{}
	if endpoint.ExtraParams.Query != "" {
		if endpoint.Slicing {
			queryStr = "/view/" + endpoint.ExtraParams.Query
		} else if !strings.Contains(queryFlag, "?") {
			queryStr = "/" + endpoint.ExtraParams.Query
		} else {
			queryStr = "?" + endpoint.ExtraParams.Query
		}
	}
	if strings.Contains(endpoint.ExtraParams.Query, ",") && endpoint.Data {
		queryStr = queryStr + "?tarfile"
	} else if endpoint.Manifest {
		suffixList = append(suffixList, "return_type=manifest")
	}
	if queryFlag == "projects" && endpoint.ExtraParams.Size == -1 {
		suffixList = append(suffixList, "size=1000000")
	}
	if endpoint.ExtraParams.Format != "" {
		suffixList = append(suffixList, "format="+endpoint.ExtraParams.Format)
	}
	if endpoint.ExtraParams.Fields != "" {
		suffixList = append(suffixList, "fields="+endpoint.ExtraParams.Fields)
	}
	if endpoint.ExtraParams.Filter != "" {
		suffixList = append(suffixList, "filter="+endpoint.ExtraParams.Filter)
	}
	if endpoint.ExtraParams.Expand != "" {
		suffixList = append(suffixList, "expand="+endpoint.ExtraParams.Expand)
	}
	if endpoint.ExtraParams.Facets != "" {
		suffixList = append(suffixList, "facets="+endpoint.ExtraParams.Facets)
	}
	if endpoint.ExtraParams.Sort != "" {
		suffixList = append(suffixList, "sort="+endpoint.ExtraParams.Sort)
	}
	if endpoint.ExtraParams.From != -1 {
		suffixList = append(suffixList, "from="+strconv.Itoa(endpoint.ExtraParams.From))
	}
	if endpoint.ExtraParams.Size != -1 {
		suffixList = append(suffixList, "size="+strconv.Itoa(endpoint.ExtraParams.Size))
	}
	if endpoint.ExtraParams.Pretty {
		suffixList = append(suffixList, "pretty=true")
	}
	if len(suffixList) > 0 && queryStr != "" {
		if !strings.Contains(queryStr, "?") {
			suffix = queryStr + "?" + strings.Join(suffixList, "&")
		} else {
			suffix = queryStr + "&" + strings.Join(suffixList, "&")
		}
	} else if len(suffixList) > 0 {
		suffix = "?" + strings.Join(suffixList, "&")
	} else if queryStr != "" {
		suffix = queryStr
	}

	return suffix
}

func postGdcQuery(queryFlag *string, resp *http.Response, endpoint *GdcEndpoints, of *os.File) {
	if *queryFlag == "projects" {
		postGdcProj(resp, endpoint, of)
	}
	if *queryFlag == "status" {
		postGdcStatus(resp, endpoint, of)
	}
	if *queryFlag == "cases" {
		postGdcCases(resp, endpoint, of)
	}
	if *queryFlag == "files" {
		postGdcFiles(resp, endpoint, of)
	}
	if *queryFlag == "annotations" {
		postGdcAnnotations(resp, endpoint, of)
	}
	if *queryFlag == "data" {
		postGdcData(resp, endpoint, of)
	}
	if *queryFlag == "manifest" {
		postGdcManifest(resp, endpoint, of)
	}
	if *queryFlag == "slicing" {
		postGdcSlicing(resp, endpoint, of)
	}
}

func postGdcStatus(resp *http.Response, endpoint *GdcEndpoints, of *os.File) {
	var status GdcStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		log.Warn(err)
	}
	log.Infoln("Print GDC portal status table.")
	table := newCmdlineRenderTable([]string{"Commit", "DataRelease", "Status", "Tag", "Version"}, of)
	table.Append([]string{status.Commit, status.DataRelease, status.Status,
		status.Tag, strconv.Itoa(status.Version)})
	table.Render()
}

func postGdcProj(resp *http.Response, endpoint *GdcEndpoints, of *os.File) {
	var projects GdcProjects
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		log.Warn(err)
	}

	table := newCmdlineRenderTable([]string{"ProjectID", "Name", "PrimarySite", "State"}, of)
	log.Infoln("Print GDC portal projects table.")
	for i := range projects.Data.Hits {
		table.Append([]string{projects.Data.Hits[i].ProjectID, projects.Data.Hits[i].Name, strings.Join(projects.Data.Hits[i].PrimarySite, "; "), projects.Data.Hits[i].State})
	}
	table.Render()
	table = newCmdlineRenderTable([]string{"ProjectID", "DiseaseType", "DbgapAccessionNumber", "Releasable", "Released"}, of)
	for i := range projects.Data.Hits {
		table.Append([]string{projects.Data.Hits[i].ProjectID, strings.Join(projects.Data.Hits[i].DiseaseType, "; "), projects.Data.Hits[i].DbgapAccessionNumber, strconv.FormatBool(projects.Data.Hits[i].Releasable), strconv.FormatBool(projects.Data.Hits[i].Released)})
	}
	table.Render()
	log.Infof("From %d to %d GDC portal projects (%d records) done.", projects.Data.Pagination.From, projects.Data.Pagination.From+projects.Data.Pagination.Count, projects.Data.Pagination.Total)
}

func postGdcCases(resp *http.Response, endpoint *GdcEndpoints, of *os.File) {
	var cases GdcCases
	if err := json.NewDecoder(resp.Body).Decode(&cases); err != nil {
		log.Warn(err)
	}
	table := newCmdlineRenderTable([]string{"CaseID", "PrimarySite", "State", "CreatedDatetime"}, of)
	log.Infoln("Print GDC portal cases table.")
	for i := range cases.Data.Hits {
		table.Append([]string{cases.Data.Hits[i].CaseID, cases.Data.Hits[i].PrimarySite, cases.Data.Hits[i].State, cases.Data.Hits[i].CreatedDatetime})
	}
	table.Render()
	table = newCmdlineRenderTable([]string{"CaseID", "SubmitterID", "DiagnosisIds", "SubmitterSampleIds"}, of)
	for i := range cases.Data.Hits {
		table.Append([]string{cases.Data.Hits[i].CaseID, cases.Data.Hits[i].SubmitterID, strings.Join(cases.Data.Hits[i].DiagnosisIds, ";"), strings.Join(cases.Data.Hits[i].SubmitterSampleIds, ";")})
	}
	table.Render()
	log.Infof("From %d to %d GDC portal cases (%d records) done.", cases.Data.Pagination.From, cases.Data.Pagination.From+cases.Data.Pagination.Count, cases.Data.Pagination.Total)
}

func postGdcFiles(resp *http.Response, endpoint *GdcEndpoints, of *os.File) {
	var files GdcFiles
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		log.Warn(err)
	}
	table := newCmdlineRenderTable([]string{"FileID", "DataFormat", "DataType", "Access", "State"}, of)
	for i := range files.Data.Hits {
		table.Append([]string{files.Data.Hits[i].FileID, files.Data.Hits[i].DataFormat, files.Data.Hits[i].DataType, files.Data.Hits[i].Access, files.Data.Hits[i].State})
	}
	log.Infoln("Print GDC portal files table.")
	table.Render()
	table = newCmdlineRenderTable([]string{"FileID", "Md5sum", "FileSize", "UpdatedDatetime"}, of)
	for i := range files.Data.Hits {
		table.Append([]string{files.Data.Hits[i].FileID, files.Data.Hits[i].Md5sum, bytefmt.ByteSize(uint64(files.Data.Hits[i].FileSize)), files.Data.Hits[i].UpdatedDatetime})
	}
	table.Render()
	log.Infof("From %d to %d GDC portal files (%d records) done.", files.Data.Pagination.From, files.Data.Pagination.From+files.Data.Pagination.Count, files.Data.Pagination.Total)
}

func postGdcAnnotations(resp *http.Response, endpoint *GdcEndpoints, of *os.File) {
	var annotations GdcAnnotations
	if err := json.NewDecoder(resp.Body).Decode(&annotations); err != nil {
		log.Warn(err)
	}
	table := newCmdlineRenderTable([]string{"AnnotationID", "CaseID", "Category", "Classification"}, of)
	for i := range annotations.Data.Hits {
		table.Append([]string{annotations.Data.Hits[i].AnnotationID, annotations.Data.Hits[i].CaseID, annotations.Data.Hits[i].Category, annotations.Data.Hits[i].Classification})
	}
	table.Render()
	table = newCmdlineRenderTable([]string{"AnnotationID", "EntityType", "EntityID", "Notes", "State"}, of)
	for i := range annotations.Data.Hits {
		table.Append([]string{annotations.Data.Hits[i].AnnotationID, annotations.Data.Hits[i].EntityType,
			annotations.Data.Hits[i].EntityID, annotations.Data.Hits[i].Notes, annotations.Data.Hits[i].State})
	}
	table.Render()
	log.Infoln("Print GDC portal annotations table.")
	log.Infof("From %d to %d GDC portal annotations (%d records) done.", annotations.Data.Pagination.From, annotations.Data.Pagination.From+annotations.Data.Pagination.Count, annotations.Data.Pagination.Total)
}

func postGdcData(resp *http.Response, endpoint *GdcEndpoints, of *os.File) {
	_, err := io.Copy(of, resp.Body)
	if err != nil {
		log.Warn(err)
	}
}
func postGdcManifest(resp *http.Response, endpoint *GdcEndpoints, of *os.File) {
	_, err := io.Copy(of, resp.Body)
	if err != nil {
		log.Warn(err)
	}
}
func postGdcSlicing(resp *http.Response, endpoint *GdcEndpoints, of *os.File) {
	_, err := io.Copy(of, resp.Body)
	if err != nil {
		log.Warn(err)
	}
}
func newCmdlineRenderTable(header []string, of *os.File) (table *tablewriter.Table) {
	table = tablewriter.NewWriter(of)
	table.SetRowLine(true)
	table.SetRowSeparator("-")
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeader(header)
	return table
}

func init() {
	pg = mpb.New(
		mpb.WithWidth(45),
		mpb.WithRefreshRate(180*time.Millisecond),
	)
}
