package fetch

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/openbiox/butils/log"
)

const Dataset2toolsHost = "http://amp.pharm.mssm.edu/datasets2tools/api/search?"

// Dataset2tools access http://amp.pharm.mssm.edu/datasets2tools/ API
func Dataset2tools(endpoints *Datasets2toolsEndpoints, outfn string, retries int, timeout int, retSleepTime int, quite bool) {
	url := Dataset2toolsHost + setDatasets2toolsQuerySuffix(endpoints)
	client := newHTTPClient(timeout)
	method := "GET"
	req, err := http.NewRequest(method, url, nil)
	setReqHeader(req)
	if err != nil {
		log.Warn(err)
	}
	log.Infof("Query datasets2tools API: %s.", url)
	resp, err := retryClient(client, req, retries, retSleepTime)
	if err != nil {
		return
	}
	if resp != nil {
		defer resp.Body.Close()
	} else {
		return
	}
	of := creatOutStream(outfn, req.URL.String())
	_, err = io.Copy(of, resp.Body)
	if err != nil {
		log.Warn(err)
	}
	defer resp.Body.Close()
	defer of.Close()

	return
}

func setDatasets2toolsQuerySuffix(endpoints *Datasets2toolsEndpoints) (suffix string) {
	suffixList := []string{}
	if endpoints.ObjectType != "" {
		suffixList = append(suffixList, "object_type="+endpoints.ObjectType)
	}
	if endpoints.DatasetAccession != "" {
		suffixList = append(suffixList, "dataset_accession="+endpoints.DatasetAccession)
	}
	if endpoints.CannedAnalysisAccession != "" {
		suffixList = append(suffixList, "canned_analysis_accession="+endpoints.CannedAnalysisAccession)
	}
	if endpoints.Query != "" {
		suffixList = append(suffixList, "q="+endpoints.Query)
	}
	if endpoints.ToolName != "" {
		suffixList = append(suffixList, "tool_name="+endpoints.ToolName)
	}
	if endpoints.DiseaseName != "" {
		suffixList = append(suffixList, "disease_name="+endpoints.DiseaseName)
	}
	if endpoints.Gneset != "" {
		suffixList = append(suffixList, "geneset="+endpoints.Gneset)
	}
	if endpoints.PageSize != -1 {
		suffixList = append(suffixList, "page_size="+strconv.Itoa(endpoints.PageSize))
	}
	if len(suffixList) > 0 {
		suffix = strings.Join(suffixList, "&")
	}
	return suffix
}
