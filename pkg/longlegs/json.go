package longlegs

import (
	"encoding/json"
	"fmt"
	"os"
)

// PrintJSON is a DEBUG FN to print obj
func printJSON(x interface{}) {
	d, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		fmt.Printf("printjson failed with error: %v", err)
		return
	}
	os.Stdout.Write(d)
	os.Stdout.Write([]byte("\n"))
}
