package cmd

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"os"

	"github.com/Miachol/bapi/format"
	"github.com/Miachol/bapi/types"
	"github.com/openbiox/butils/log"
	"github.com/spf13/cobra"
)

var fmtClis = types.FmtClisT{}
var fmtCmd = &cobra.Command{
	Use:   "fmt [input1 input2]",
	Short: "A set of file format (fmt) command of bapi.",
	Long:  `A set of file format (fmt) command of bapi. More see here https://github.com/Miachol/bapi.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmtCmdRunOptions(cmd)
	},
}

func fmtCmdRunOptions(cmd *cobra.Command) {
	cleanArgs := []string{}
	hasStdin := false
	if cleanArgs, hasStdin = checkStdInFlag(cmd); hasStdin {
		reader := bufio.NewReader(os.Stdin)
		result, err := ioutil.ReadAll(reader)
		if err != nil {
			log.Fatal(err)
		} else if len(result) > 0 {
			fmtClis.Stdin = &result
		}
	} else {
		fmtClis.Stdin = nil
	}

	if len(cleanArgs) >= 1 || hasStdin {
		fmtClis.Files = &cleanArgs
		runFlag := false
		if fmtClis.PrettyJSON {
			format.PrettyJSON(&fmtClis, bapiClis.Thread)
			runFlag = true
		} else if fmtClis.JSONToSlice {
			format.JSON2Slice(&fmtClis, bapiClis.Thread)
			runFlag = true
		}
		if !runFlag && hasStdin {
			io.Copy(os.Stdout, bytes.NewBuffer(*fmtClis.Stdin))
		}
		bapiClis.HelpFlags = false
	}
	if bapiClis.HelpFlags {
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
	fmtCmd.Flags().IntVarP(&bapiClis.Thread, "thread", "t", 1, "Thread to process.")
	fmtCmd.Flags().BoolVarP(&fmtClis.JSONToSlice, "json-to-slice", "", false, "Convert key-value JSON  to []key-value and easy to export to readable table.")
	fmtCmd.Flags().BoolVarP(&fmtClis.PrettyJSON, "json-pretty", "", false, "Pretty json files.")
	fmtCmd.Flags().IntVarP(&fmtClis.Indent, "indent", "", 4, "Control the indent of output json files.")
	fmtCmd.Flags().BoolVarP(&fmtClis.SortKeys, "sort-keys", "", false, "Control wheather to sort JSON key.")
	fmtCmd.Example = `  bapi ncbi -q "Galectins control MTOR and AMPK in response to lysosomal damage to induce autophagy OR MTOR-independent autophagy induced by interrupted endoplasmic reticulum-mitochondrial Ca2+ communication: a dead end in cancer cells. OR The PARK10 gene USP24 is a negative regulator of autophagy and ULK1 protein stability OR Coordinate regulation of autophagy and the ubiquitin proteasome system by MTOR." | bapi ncbi --xml2json pubmed - | sed 's;}{;,;g' | bapi fmt --json-to-slice --indent 4 -| json2csv -o final.csv`
	JSON := make(map[int]map[string]interface{})
	fmtClis.JSON = &JSON
	Table := make(map[int][]interface{})
	fmtClis.Table = &Table
}
