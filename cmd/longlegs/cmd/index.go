package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/candland/longlegs/pkg/longlegs"
	"github.com/spf13/cobra"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type MySite struct {
	ProcessedCount int
}

func (site *MySite) Process(spider *longlegs.Spider, page longlegs.Page) {
	site.ProcessedCount++
	log.Printf(">  %s took %d ms", page.Id, page.Ms)
}

func (site *MySite) ProcessError(spider *longlegs.Spider, page longlegs.Page) {
	site.ProcessedCount++
	log.Printf(">  %s took %d ms", page.Id, page.Ms)
}

func (site *MySite) Blocked(spider *longlegs.Spider, url url.URL) {
	log.Printf(">  %s blocked", url.String())
}

// / flags
var depth int
var limit int

var debug bool

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

		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

		log.Info().Msgf("Indexing %s", args[0])

		spider, err := longlegs.NewSpider("longlegs/0", args[0])
		if err != nil {
			panic(err)
		}

		site := &MySite{}
		spider.Crawl(site, depth, limit)

		printJSON(site)
	},
}

func init() {
	indexCmd.Flags().IntVarP(&depth, "depth", "d", 0, "Depth to crawl.")
	indexCmd.Flags().IntVarP(&limit, "limit", "l", 0, "Number of page to limit crawl to.")
	indexCmd.Flags().BoolVarP(&debug, "debug", "v", false, "Display debug messages.")
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
