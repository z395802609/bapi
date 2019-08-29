package cmd

import (
	"bufio"
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
	stdin      []byte
	files      []string
	json       map[int]map[string]interface{}
	table      map[int][]interface{}
	prettyJSON bool
	json2slice bool
	json2csv   bool
	thread     int
	indent     string
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

	cleanArgs := []string{}
	hasStdin := false
	if cleanArgs, hasStdin = checkStdInFlag(cmd); hasStdin {
		reader := bufio.NewReader(os.Stdin)
		result, err := reader.ReadString('\n')
		if err == nil {
			log.Fatal(err)
		} else if result != "" {
			fmtClis.stdin = []byte(result)
		}
	}

	if len(cleanArgs) >= 1 || hasStdin {
		fmtClis.files = cleanArgs
		runFlag := false
		if fmtClis.prettyJSON {
			prettyJSON()
			runFlag = true
		}
		if fmtClis.json2slice {
			JSON2Slice()
			runFlag = true
		}
		if !runFlag {
			io.Copy(os.Stdout, bytes.NewBuffer(fmtClis.stdin))
		}
		bapiClis.helpFlags = false
	}
	if bapiClis.helpFlags {
		cmd.Help()
	}
}

func checkStdInFlag(cmd *cobra.Command) (args []string, hasStdin bool) {
	for _, v := range cmd.Flags().Args() {
		if v != "-" {
			args = append(args, v)
		} else {
			hasStdin = true
		}
	}
	return args, hasStdin
}

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func prettyJSON() {
	sem := make(chan bool, bapiClis.thread)
	var m map[string]interface{}
	m = make(map[string]interface{})
	var d []byte
	for k, fn := range fmtClis.files {
		sem <- true
		go func(fn string, k int) {
			defer func() {
				<-sem
			}()
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
	if len(fmtClis.stdin) > 0 {
		var m2 map[string]interface{}
		m2 = make(map[string]interface{})
		json.Unmarshal(fmtClis.stdin, &m2)
		d = pretty.Pretty(fmtClis.stdin)
		fmtClis.json[-1] = m2
		io.Copy(os.Stdout, bytes.NewBuffer(d))
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
				d, _ = json.MarshalIndent(final, "", fmtClis.indent)
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
	fmtCmd.Flags().StringVarP(&fmtClis.indent, "indent", "", "    ", "Control the indent of output json files.")
	fmtCmd.Flags().BoolVarP(&fmtClis.prettyJSON, "json-pretty", "", false, "Pretty json files.")
	fmtCmd.Flags().BoolVarP(&fmtClis.json2slice, "json-to-slice", "", false, "Convert key-value JSON  to []key-value and easy to export to readable table.")
	fmtClis.json = make(map[int]map[string]interface{})
	fmtClis.table = make(map[int][]interface{})
}
