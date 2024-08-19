package main

import (
	"flag"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	_ "embed"

	junit "github.com/joshdk/go-junit"

	"github.com/go-sprout/sprout"
	"github.com/go-sprout/sprout/registry/maps"
)

var (
	mode   string
	input  string
	output string
)

//go:embed tpl/report.html
var reportTemplate string

func render() func(io.Writer) error {
	sproutHandler := sprout.New()
	sproutHandler.AddRegistry(maps.NewRegistry())
	t, err := template.New("report.html").Funcs(sproutHandler.Build()).Parse(reportTemplate)
	if err != nil {
		log.Panic(err)
	}
	inputFile, err := os.ReadFile(input)
	if err != nil {
		log.Panic(err)
	}
	suites, err := junit.Ingest(inputFile)
	if err != nil {
		log.Panic(err)
	}
	return func(w io.Writer) error {
		return t.Execute(w, suites)
	}

}

func server(render func(io.Writer) error) {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		render(w)
	})
	log.Println("server starting at :3000")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Panic(err)
	}

}

func inline() {

}

func main() {
	flag.StringVar(&mode, "mode", "server", "operation mode")
	flag.StringVar(&input, "input", "examples/junit-complete.xml", "junit xml output")
	flag.StringVar(&output, "outputDir", ".", "report output directory")
	renterFunc := render()
	switch mode {
	case "server":
		server(renterFunc)
	default:
		inline()
	}
}
