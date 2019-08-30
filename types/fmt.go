package types

type FmtClisT struct {
	Stdin       *[]byte
	Files       *[]string
	JSON        *map[int]map[string]interface{}
	Table       *map[int][]interface{}
	PrettyJSON  bool
	JSONToSlice bool
	JSONToCSV   bool
	Indent      int
	SortKeys    bool
}
