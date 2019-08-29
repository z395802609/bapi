package fetch

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"unicode/utf8"

	"errors"
	"mime"
	"net/http/cookiejar"
	"path"
	"strings"
	"time"

	"github.com/openbiox/butils"
	mpb "github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"

	"github.com/openbiox/butils/log"
)

var pg *mpb.Progress
var gCurCookies []*http.Cookie
var gCurCookieJar *cookiejar.Jar

func setQueryFromEnd(from int, size int, total int) (int, int) {
	if size == -1 {
		size = total + 1
	}
	end := from + size
	if end == -1 || end > total {
		end = total + 1
	}
	if from < 0 {
		from = 0
	} else if from > total {
		from = total
	}
	if end <= from {
		end = from + 1
	}
	return from, end
}

// HTTPDownload can use golang http.Get to query URL with progress bar
func HTTPDownload(url string, destFn string, pg *mpb.Progress, quiet bool, saveLog bool, retries int, timeout int, retSleepTime int) error {
	client := &http.Client{
		CheckRedirect: defaultCheckRedirect,
		Jar:           gCurCookieJar,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: time.Duration(timeout) * time.Second,
			}).Dial,
		},
	}
	var t int
	success := false

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/56.0.2924.87 Safari/537.36")
	if err != nil {
		// handle error
		log.Warn(err)
		return err
	}
	gCurCookies = gCurCookieJar.Cookies(req.URL)

	for t = 0; t < retries; t++ {
		err := downloadWorker(client, req, url, destFn, pg, quiet, saveLog)
		if err == nil {
			success = true
			break
		} else {
			log.Warnf("Failed to retrive on attempt %d... error: %v ... retrying after %d seconds.", t+1, err, retSleepTime)
			time.Sleep(time.Duration(retSleepTime) * time.Second)
		}
	}
	if !success {
		return err
	}
	return nil
}

func downloadWorker(client *http.Client, req *http.Request, url string, destFn string, pg *mpb.Progress, quiet bool, saveLog bool) error {
	resp, err := client.Do(req)
	if err != nil {
		log.Warn(err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if !quiet {
			log.Warnf("Access failed: %s", url)
			fmt.Println("")
		}
		return err
	}
	size := resp.ContentLength

	if hasParDir, _ := butils.PathExists(filepath.Dir(destFn)); !hasParDir {
		err := butils.CreateFileParDir(destFn)
		if err != nil {
			log.Fatal(err)
		}
	}
	// create dest
	destName := filepath.Base(url)
	dest, err := os.Create(destFn)
	if err != nil {
		log.Warnf("Can't create %s: %v\n", destName, err)
		return err
	}
	prefixStr := filepath.Base(destFn)
	prefixStrLen := utf8.RuneCountInString(prefixStr)
	if prefixStrLen > 35 {
		prefixStr = prefixStr[0:31] + "..."
	}
	prefixStr = fmt.Sprintf("%-35s\t", prefixStr)
	if !quiet {
		bar := pg.AddBar(size,
			mpb.BarStyle("[=>-|"),
			mpb.PrependDecorators(
				decor.Name(prefixStr, decor.WC{W: len(prefixStr) + 1, C: decor.DidentRight}),
				decor.CountersKibiByte("% -.1f / % -.1f\t"),
				decor.OnComplete(decor.Percentage(decor.WC{W: 5}), " "+"√"),
			),
			mpb.AppendDecorators(
				decor.EwmaETA(decor.ET_STYLE_MMSS, float64(size)/2048),
				decor.Name(" ] "),
				decor.AverageSpeed(decor.UnitKiB, "% .1f"),
			),
		)
		// create proxy reader
		reader := bar.ProxyReader(resp.Body)
		// and copy from reader, ignoring errors
		_, err := io.Copy(dest, reader)
		if err != nil {
			bar.Abort(true)
			reader.Close()
			log.Warn(err)
			return err
		}
	} else {
		_, err := io.Copy(dest, io.Reader(resp.Body))
		if err != nil {
			log.Warn(err)
			return err
		}
	}
	defer dest.Close()
	return nil
}

func defaultCheckRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 20 {
		return errors.New("stopped after 20 redirects")
	}
	return nil
}

func newHTTPClient(timeout int) *http.Client {
	return &http.Client{
		CheckRedirect: defaultCheckRedirect,
		Jar:           gCurCookieJar,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: time.Duration(timeout) * time.Second,
			}).Dial,
		},
	}
}
func setReqHeader(req *http.Request) {
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/56.0.2924.87 Safari/537.36")
}

func retryClient(client *http.Client, req *http.Request, retries int, retSleepTime int) (resp *http.Response, err error) {
	for t := 0; t < retries; t++ {
		resp, err = client.Do(req)
		if err != nil {
			log.Warnf("Failed to retrieve on attempt %d... error: %v ... retrying after %d seconds.", t+1, err, retSleepTime)
			time.Sleep(time.Duration(retSleepTime) * time.Second)
			continue
		} else if err2 := checkResp(resp); err2 != nil {
			return nil, err2
		} else {
			break
		}
	}
	return resp, err
}

func parseOutfnFromHeader(outfn string, resp *http.Response, useRemoteName bool) string {
	contentDis := resp.Header.Get("Content-Disposition")
	if outfn == "" && contentDis != "" && useRemoteName &&
		strings.Contains(contentDis, "filename") {
		_, params, err := mime.ParseMediaType(contentDis)
		if err != nil {
			log.Warn(err)
		} else {
			outfn = params["filename"]
		}
	}
	return outfn
}

// set of as standout or file
func creatOutStream(outfn string, url string) *os.File {
	var of *os.File
	if outfn == "" {
		of = os.Stdout
	} else {
		var err error
		of, err = os.OpenFile(outfn, os.O_CREATE|os.O_WRONLY, 0664)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		wd, _ := os.Getwd()
		if url != "" {
			log.Infof("Trying %s => %s", url, path.Join(wd, outfn))
		}
	}
	return of
}

func checkResp(resp *http.Response) (err error) {
	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("access failed: %s", resp.Request.URL.String()))
	}
	return nil
}
func init() {
	pg = mpb.New(
		mpb.WithWidth(45),
		mpb.WithRefreshRate(180*time.Millisecond),
	)
	gCurCookies = nil
	//var err error;
	gCurCookieJar, _ = cookiejar.New(nil)
}
