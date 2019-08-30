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
			fmtClis.Stdin = result
		}
	}

	if len(cleanArgs) >= 1 || hasStdin {
		fmtClis.Files = cleanArgs
		runFlag := false
		if fmtClis.PrettyJSON {
			format.PrettyJSON(&fmtClis, bapiClis.Thread)
			runFlag = true
		} else if fmtClis.JSONToSlice {
			format.JSON2Slice(&fmtClis, bapiClis.Thread)
			runFlag = true
		}
		if !runFlag && hasStdin {
			io.Copy(os.Stdout, bytes.NewBuffer(fmtClis.Stdin))
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
	fmtClis.JSON = make(map[int]map[string]interface{})
	fmtClis.Table = make(map[int][]interface{})
}
