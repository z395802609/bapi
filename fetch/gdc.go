package fetch

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/cloudfoundry/bytefmt"
	"github.com/olekukonko/tablewriter"

	"github.com/JhuangLab/butils/log"
)

const GdcAPIHost = "https://api.gdc.cancer.gov"
const GdcAPIHostLegacy = "https://api.gdc.cancer.gov/legacy"

var endpoints = []string{"status", "projects", "cases", "files", "annotations",
	"data", "manifest", "slicing", "submission"}

var tables []*tablewriter.Table

func Gdc(endpoint GdcEndpoints) {
	client := &http.Client{}
	host := GdcAPIHost
	if endpoint.Legacy {
		host = GdcAPIHostLegacy
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
			req, queryFlag = setGdcReq(host, i)
		} else if f.Kind() == reflect.Bool && f.Bool() {
			req, queryFlag = setGdcReq(host, i)
		}
		if queryFlag == "" {
			continue
		}
		log.Infof("Query GDC portal %s API......", queryFlag)
		resp, err := client.Do(req)
		if err != nil {
			log.Warn(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			log.Warnf("Access failed: %s", host+"/"+endpoints[i])
			fmt.Println("")
			return
		}
		postGdcQuery(&queryFlag, resp, &endpoint)
	}
}

func setGdcReq(host string, i int) (*http.Request, string) {
	queryFlag := endpoints[i]
	suffix := setGdcQuerySuffix(queryFlag)
	req, err := http.NewRequest("GET", host+"/"+endpoints[i]+suffix, nil)
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/56.0.2924.87 Safari/537.36")
	if err != nil {
		log.Warn(err)
	}
	return req, queryFlag
}

func setGdcQuerySuffix(queryFlag string) (suffix string) {
	if queryFlag == "projects" {
		suffix = "?size=1000000"
	}
	return suffix
}

func postGdcQuery(queryFlag *string, resp *http.Response, endpoint *GdcEndpoints) {
	if *queryFlag == "projects" {
		postGdcProj(resp, endpoint)
	}
	if *queryFlag == "status" {
		postGdcStatus(resp, endpoint)
	}
	if *queryFlag == "cases" {
		postGdcCases(resp, endpoint)
	}
	if *queryFlag == "files" {
		postGdcFiles(resp, endpoint)
	}
	if *queryFlag == "annotations" {
		postGdcAnnotations(resp, endpoint)
	}
}

func postGdcStatus(resp *http.Response, endpoint *GdcEndpoints) {
	var status GdcStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		log.Warn(err)
	}
	log.Infoln("Print GDC portal status table.")
	table := newCmdlineRenderTable([]string{"Commit", "DataRelease", "Status", "Tag", "Version"})
	table.Append([]string{status.Commit, status.DataRelease, status.Status,
		status.Tag, strconv.Itoa(status.Version)})
	table.Render()
}

func postGdcProj(resp *http.Response, endpoint *GdcEndpoints) error {
	if endpoint.Json || endpoint.ExtraParams.Pretty {
		_, err := io.Copy(os.Stdout, resp.Body)
		return err
	}
	var projects GdcProjects
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		log.Warn(err)
	}

	table := newCmdlineRenderTable([]string{"ProjectID", "Name", "PrimarySite", "State"})
	log.Infoln("Print GDC portal projects table.")
	for i := range projects.Data.Hits {
		table.Append([]string{projects.Data.Hits[i].ProjectID, projects.Data.Hits[i].Name, strings.Join(projects.Data.Hits[i].PrimarySite, "; "), projects.Data.Hits[i].State})
	}
	table.Render()
	table = newCmdlineRenderTable([]string{"ProjectID", "DiseaseType", "DbgapAccessionNumber", "Releasable", "Released"})
	for i := range projects.Data.Hits {
		table.Append([]string{projects.Data.Hits[i].ProjectID, strings.Join(projects.Data.Hits[i].DiseaseType, "; "), projects.Data.Hits[i].DbgapAccessionNumber, strconv.FormatBool(projects.Data.Hits[i].Releasable), strconv.FormatBool(projects.Data.Hits[i].Released)})
	}
	table.Render()
	log.Infof("%d/%d GDC portal projects done.", len(projects.Data.Hits), projects.Data.Pagination.Total)
	return nil
}

func postGdcCases(resp *http.Response, endpoint *GdcEndpoints) {
	var cases GdcCases
	if err := json.NewDecoder(resp.Body).Decode(&cases); err != nil {
		log.Warn(err)
	}
	table := newCmdlineRenderTable([]string{"CaseID", "PrimarySite", "State", "CreatedDatetime"})
	log.Infoln("Print GDC portal cases table.")
	for i := range cases.Data.Hits {
		table.Append([]string{cases.Data.Hits[i].CaseID, cases.Data.Hits[i].PrimarySite, cases.Data.Hits[i].State, cases.Data.Hits[i].CreatedDatetime})
	}
	table.Render()
	table = newCmdlineRenderTable([]string{"CaseID", "SubmitterID", "DiagnosisIds", "SubmitterSampleIds"})
	for i := range cases.Data.Hits {
		table.Append([]string{cases.Data.Hits[i].CaseID, cases.Data.Hits[i].SubmitterID, strings.Join(cases.Data.Hits[i].DiagnosisIds, ";"), strings.Join(cases.Data.Hits[i].SubmitterSampleIds, ";")})
	}
	table.Render()
	log.Infof("%d/%d GDC portal cases done.", len(cases.Data.Hits), cases.Data.Pagination.Total)
}

func postGdcFiles(resp *http.Response, endpoint *GdcEndpoints) {
	var files GdcFiles
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		log.Warn(err)
	}
	table := newCmdlineRenderTable([]string{"FileID", "DataFormat", "DataType", "Access", "State"})
	for i := range files.Data.Hits {
		table.Append([]string{files.Data.Hits[i].FileID, files.Data.Hits[i].DataFormat, files.Data.Hits[i].DataType, files.Data.Hits[i].Access, files.Data.Hits[i].State})
	}
	log.Infoln("Print GDC portal files table.")
	table.Render()
	table = newCmdlineRenderTable([]string{"FileID", "Md5sum", "FileSize", "UpdatedDatetime"})
	for i := range files.Data.Hits {
		table.Append([]string{files.Data.Hits[i].FileID, files.Data.Hits[i].Md5sum, bytefmt.ByteSize(uint64(files.Data.Hits[i].FileSize)), files.Data.Hits[i].UpdatedDatetime})
	}
	table.Render()
	log.Infof("%d/%d GDC portal files done.", len(files.Data.Hits), files.Data.Pagination.Total)
}

func postGdcAnnotations(resp *http.Response, endpoint *GdcEndpoints) {
	var annotations GdcAnnotations
	if err := json.NewDecoder(resp.Body).Decode(&annotations); err != nil {
		log.Warn(err)
	}
	table := newCmdlineRenderTable([]string{"AnnotationID", "CaseID", "Category", "Classification"})
	for i := range annotations.Data.Hits {
		table.Append([]string{annotations.Data.Hits[i].AnnotationID, annotations.Data.Hits[i].CaseID, annotations.Data.Hits[i].Category, annotations.Data.Hits[i].Classification})
	}
	table.Render()
	table = newCmdlineRenderTable([]string{"AnnotationID", "EntityType", "EntityID", "Notes", "State"})
	for i := range annotations.Data.Hits {
		table.Append([]string{annotations.Data.Hits[i].AnnotationID, annotations.Data.Hits[i].EntityType,
			annotations.Data.Hits[i].EntityID, annotations.Data.Hits[i].Notes, annotations.Data.Hits[i].State})
	}
	table.Render()
	log.Infoln("Print GDC portal annotations table.")
	log.Infof("%d/%d GDC portal annotations done.", len(annotations.Data.Hits), annotations.Data.Pagination.Total)
}

func newCmdlineRenderTable(header []string) (table *tablewriter.Table) {
	table = tablewriter.NewWriter(os.Stdout)
	table.SetRowLine(true)
	table.SetRowSeparator("-")
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeader(header)
	return table
}
