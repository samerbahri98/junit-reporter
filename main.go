package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"

	junit "github.com/joshdk/go-junit"

	"github.com/go-sprout/sprout"
	"github.com/go-sprout/sprout/registry/maps"
)

var (
	mode   string
	input  string
	output string
)

func server() {
	sproutHandler := sprout.New()
	sproutHandler.AddRegistry(maps.NewRegistry())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t, err := template.New("report.html").Funcs(sproutHandler.Build()).ParseFiles("tpl/report.html")
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

		t.Execute(w, suites)
	})
	log.Println("server starting at :3333")
	if err := http.ListenAndServe(":3333", nil); err != nil {
		log.Panic(err)
	}

}

func inline() {

}

func main() {
	flag.StringVar(&mode, "mode", "server", "operation mode")
	flag.StringVar(&input, "input", "examples/junit-complete.xml", "junit xml output")
	flag.StringVar(&output, "outputDir", ".", "report output directory")

	switch mode {
	case "server":
		server()
	default:
		inline()
	}
}
