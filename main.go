package main

import (
	"encoding/json"
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

type ObjectContext struct {
	InputS3Url  string `json:"inputS3Url"`
	OutputRoute string `json:"outputRoute"`
	OutputToken string `json:"outputToken"`
}

type Event struct {
	ObjectContext ObjectContext `json:"getObjectContext"`
}

var (
	mode   string
	input  string
	output string
)

//go:embed tpl/report.html
var reportTemplate string

func server() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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
		t.Execute(w, suites)
	})
	log.Println("server starting at :3000")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Panic(err)
	}

}

func objectLambda() {
	http.HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
		var event Event
		err := json.NewDecoder(r.Body).Decode(&event)
		if err != nil {
			log.Panic(err)
		}
		req, err := http.Get(event.ObjectContext.InputS3Url)
		if err != nil {
			log.Panic(err)
		}
		defer req.Body.Close()
		sproutHandler := sprout.New()
		sproutHandler.AddRegistry(maps.NewRegistry())
		t, err := template.New("report.html").Funcs(sproutHandler.Build()).Parse(reportTemplate)
		if err != nil {
			log.Panic(err)
		}

		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Panic(err)
		}
		suites, err := junit.Ingest(body)
		if err != nil {
			log.Panic(err)
		}
		t.Execute(w, suites)
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
	switch mode {
	case "server":
		server()
	case "objectLambda":
		objectLambda()
	default:
		inline()
	}
}
