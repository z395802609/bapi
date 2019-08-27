package fetch

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/JhuangLab/butils"
	"github.com/JhuangLab/butils/log"
	mpb "github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"
)

var gCurCookies []*http.Cookie
var gCurCookieJar *cookiejar.Jar

func createIOStream(of *os.File, outfn string) *os.File {
	var err error
	if outfn == "" {
		of = os.Stdout
	} else {
		of, err = os.Create(outfn)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		of.Name()
	}
	return of
}

func setQueryFromEnd(from int, size int, total int) (int, int) {
	end := from + size
	if end == -1 || end > total {
		end = total
	}
	if from < 0 {
		from = 0
	} else if from > total {
		from = total
	}
	if end < from {
		end = from
	}
	return from, end
}

func defaultCheckRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 20 {
		return errors.New("stopped after 20 redirects")
	}
	return nil
}

// HTTPDownload can use golang http.Get to query URL with progress bar
func HTTPDownload(url string, destFn string, pg *mpb.Progress, quiet bool, saveLog bool) {
	client := &http.Client{
		CheckRedirect: defaultCheckRedirect,
		Jar:           gCurCookieJar,
	}

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/56.0.2924.87 Safari/537.36")
	if err != nil {
		// handle error
		log.Warn(err)
		return
	}
	gCurCookies = gCurCookieJar.Cookies(req.URL)
	resp, err := client.Do(req)
	if err != nil {
		log.Warn(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if !quiet {
			log.Warnf("Access failed: %s", url)
			fmt.Println("")
		}
		return
	}
	if checkHTTPGetURLRdirect(resp, url, destFn, pg, quiet, saveLog) {
		return
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
		return
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
		io.Copy(dest, reader)
	} else {
		io.Copy(dest, io.Reader(resp.Body))
	}
	defer dest.Close()
}

func checkHTTPGetURLRdirect(resp *http.Response, url string, destFn string, pg *mpb.Progress, quiet bool, saveLog bool) (status bool) {
	if strings.Contains(url, "https://www.sciencedirect.com") {
		v, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			if butils.StrDetect(string(v), "https://pdf.sciencedirectassets.com") {
				url = butils.StrExtract(string(v), `https://pdf.sciencedirectassets.com/.*&type=client`, 1)[0]
				HTTPDownload(url, destFn, pg, quiet, saveLog)
				return true
			}
		}
	}
	return false
}

func init() {
	gCurCookies = nil
	//var err error;
	gCurCookieJar, _ = cookiejar.New(nil)
}
