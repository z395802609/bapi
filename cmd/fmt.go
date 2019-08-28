package cmd

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"

	jsoniter "github.com/json-iterator/go"
	"github.com/openbiox/butils"
	"github.com/openbiox/butils/log"
	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"
)

type fmtClisT struct {
	files      []string
	json       map[int]map[string]interface{}
	table      map[int][]interface{}
	fmtJSON    bool
	json2slice bool
	json2csv   bool
	thread     int
}

var fmtClis = fmtClisT{}
var fmtCmd = &cobra.Command{
	Use:   "fmt [input1 input2]",
	Short: "A set of file format (fmt) command of bapi.",
	Long:  `A set of file format (fmt) command of bapi. More see here https://github.com/Miachol/bapi.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmtCmdRunOptions(cmd)
	},
}

func fmtCmdRunOptions(cmd *cobra.Command) {
	if bapiClis.quiet {
		log.SetOutput(ioutil.Discard)
	} else {
		log.SetOutput(os.Stderr)
	}

	if len(cmd.Flags().Args()) >= 1 {
		fmtClis.files = cmd.Flags().Args()
		if fmtClis.fmtJSON {
			fmtJson()
		}
		if fmtClis.json2slice {
			JSON2Slice()
		}
		bapiClis.helpFlags = false
	}
	if bapiClis.helpFlags {
		cmd.Help()
	}
}

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func fmtJson() {
	sem := make(chan bool, bapiClis.thread)
	for k, fn := range fmtClis.files {
		sem <- true
		go func(fn string, k int) {
			defer func() {
				<-sem
			}()
			var m map[string]interface{}
			m = make(map[string]interface{})
			outfn := butils.StrReplaceAll(fn, "json$", "pretty.json")
			fmtClis.files[k] = outfn
			d, err := ioutil.ReadFile(fn)
			if err != nil {
				log.Fatal(err)
			}
			json.Unmarshal(d, &m)
			fmtClis.json[k] = m
			d = pretty.Pretty(d)
			f, err := os.OpenFile(outfn, os.O_RDWR|os.O_CREATE, 0664)
			io.Copy(f, bytes.NewBuffer(d))
		}(fn, k)
	}
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
}

func JSON2Slice() {
	sem := make(chan bool, bapiClis.thread)
	for k, fn := range fmtClis.files {
		sem <- true
		go func(fn string, k int) {
			defer func() {
				<-sem
			}()
			var m map[string]interface{}
			m = make(map[string]interface{})
			outfn := butils.StrReplaceAll(fn, "json$", "slice.json")
			fmtClis.files[k] = outfn
			var d []byte
			var err error
			if len(fmtClis.json[k]) > 0 {
				m = fmtClis.json[k]
			} else {
				d, err = ioutil.ReadFile(fn)
				if err != nil {
					log.Fatal(err)
				}
			}
			json.Unmarshal(d, &m)
			var final []interface{}
			var j string
			for j = range m {
				final = append(final, m[j])
			}
			fmtClis.table[k] = final
			if j != "" {
				d, _ = json.MarshalIndent(final, "", "      ")
				d = pretty.Pretty(d)
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
}

func init() {
	fmtCmd.Flags().BoolVarP(&fmtClis.fmtJSON, "json", "", false, "fmt input json files.")
	fmtCmd.Flags().BoolVarP(&fmtClis.json2slice, "json-to-slice", "", false, "Convert key-value JSON  to []key-value and easy to export to readable table.")
	fmtClis.json = make(map[int]map[string]interface{})
	fmtClis.table = make(map[int][]interface{})
}
