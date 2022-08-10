package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

type FormatOutputFunc func(string, io.Writer, int) error

// FormattedPrint prints output according to the chosen format using the passed functions
func FormattedPrint(raw string, outputFormat string, out io.Writer, page int, printFunctions map[string]FormatOutputFunc) error {
	if v, ok := printFunctions[outputFormat]; ok {
		return v(raw, out, page)
	}
	return fmt.Errorf("unsupported output format: %s", outputFormat)
}

// PrintJson prints a raw json to an out writer in pretty format (indented)
func PrintJson(raw string, out io.Writer, page int) error {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, []byte(raw), "", "  ")
	if err != nil {
		return err
	}

	fmt.Fprint(out, prettyJSON.String())
	return nil
}
