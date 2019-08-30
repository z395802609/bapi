package format

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"

	jsoniter "github.com/json-iterator/go"
	"github.com/openbiox/butils"
	"github.com/openbiox/butils/log"
	"github.com/tidwall/pretty"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func PrettyJSON(files *[]string, stdin *[]byte, backendJson *map[int]map[string]interface{}, thread int, indent *string, sortKeys bool) {
	sem := make(chan bool, thread)
	var m map[string]interface{}
	m = make(map[string]interface{})
	var d []byte
	for k, fn := range *files {
		sem <- true
		go func(fn string, k int) {
			defer func() {
				<-sem
			}()
			outfn := butils.StrReplaceAll(fn, "json$", "pretty.json")
			(*files)[k] = outfn
			d, err := ioutil.ReadFile(fn)
			if err != nil {
				log.Fatal(err)
			}
			opt := pretty.Options{
				Indent:   *indent,
				SortKeys: sortKeys,
			}
			d = pretty.PrettyOptions(d, &opt)
			json.Unmarshal(d, &m)
			(*backendJson)[k] = m
			f, err := os.OpenFile(outfn, os.O_RDWR|os.O_CREATE, 0664)
			io.Copy(f, bytes.NewBuffer(d))
		}(fn, k)
	}
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
	if len(*stdin) > 0 {
		var m2 map[string]interface{}
		m2 = make(map[string]interface{})
		json.Unmarshal(*stdin, &m2)
		d = pretty.Pretty(*stdin)
		(*backendJson)[-1] = m2
		io.Copy(os.Stdout, bytes.NewBuffer(d))
	}
}

func JSON2Slice(files []string, stdin *[]byte, backendJson *map[int]map[string]interface{}, backendTable *map[int][]interface{}, thread int, indent *string, sortKeys bool) {
	sem := make(chan bool, thread)
	var final []interface{}
	var j string
	for k, fn := range files {
		sem <- true
		go func(fn string, k int) {
			defer func() {
				<-sem
			}()
			var m map[string]interface{}
			m = make(map[string]interface{})
			outfn := butils.StrReplaceAll(fn, "json$", "slice.json")
			files[k] = outfn
			var d []byte
			var err error
			if len((*backendJson)[k]) > 0 {
				m = (*backendJson)[k]
			} else {
				d, err = ioutil.ReadFile(fn)
				if err != nil {
					log.Fatal(err)
				}
			}
			json.Unmarshal(d, &m)

			for j = range m {
				final = append(final, m[j])
			}
			(*backendTable)[k] = final
			if j != "" {
				d, _ = json.Marshal(final)
				opt := pretty.Options{
					Indent:   *indent,
					SortKeys: sortKeys,
				}
				d = pretty.PrettyOptions(d, &opt)
				f, err := os.OpenFile(outfn, os.O_RDWR|os.O_CREATE, 0664)
				if err != nil {
					log.Fatal(err)
				}
				defer f.Close()
				io.Copy(f, bytes.NewBuffer(d))
			}
		}(fn, k)
	}
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
	if len(*stdin) > 0 {
		var m2 map[string]interface{}
		m2 = make(map[string]interface{})
		json.Unmarshal(*stdin, &m2)
		var final []interface{}
		var j string
		for j = range m2 {
			final = append(final, m2[j])
		}
		(*backendTable)[-1] = final
		if j != "" {
			d, _ := json.MarshalIndent(final, "", *indent)
			d = pretty.Pretty(d)
			io.Copy(os.Stdout, bytes.NewBuffer(d))
		}
	}
}
