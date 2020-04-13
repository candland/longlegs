package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/candland/longlegs/pkg/longlegs"
	"github.com/spf13/cobra"
)

// indexCmd represents the index command
var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println("Indexing")

		urlStr := "https://candland.net"
		indexLimit := 1

		site, err := longlegs.NewSite(urlStr)
		if err != nil {
			panic(err)
		}

		processDebug := func(page longlegs.Page) longlegs.Page {
			printJSON(page)
			return page
		}

		site.Index(indexLimit, longlegs.Pipeline(processDebug))

		printJSON(site)
	},
}

func init() {
	rootCmd.AddCommand(indexCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// indexCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// indexCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

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
