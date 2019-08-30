package cmd

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"os"

	"github.com/Miachol/bapi/format"
	"github.com/openbiox/butils/log"
	"github.com/spf13/cobra"
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
	sortKey    bool
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
		result, err := ioutil.ReadAll(reader)
		if err != nil {
			log.Fatal(err)
		} else if len(result) > 0 {
			fmtClis.stdin = result
		}
	}

	if len(cleanArgs) >= 1 || hasStdin {
		fmtClis.files = cleanArgs
		runFlag := false
		if fmtClis.prettyJSON {
			format.PrettyJSON(&fmtClis.files, &(fmtClis.stdin), &(fmtClis.json), fmtClis.thread, &fmtClis.indent, fmtClis.sortKey)
			runFlag = true
		} else if fmtClis.json2slice {
			format.JSON2Slice(fmtClis.files, &(fmtClis.stdin), &(fmtClis.json), &(fmtClis.table), fmtClis.thread, &fmtClis.indent, fmtClis.sortKey)
			runFlag = true
		}
		if !runFlag && hasStdin {
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

func init() {
	fmtCmd.Flags().IntVarP(&fmtClis.thread, "thread", "t", 1, "Thread to process.")
	fmtCmd.Flags().BoolVarP(&fmtClis.json2slice, "json-to-slice", "", false, "Convert key-value JSON  to []key-value and easy to export to readable table.")
	fmtClis.json = make(map[int]map[string]interface{})
	fmtClis.table = make(map[int][]interface{})
}
