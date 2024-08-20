package main

import (
	"context"
	"flag"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	_ "embed"

	junit "github.com/joshdk/go-junit"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

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
	sproutHandler := sprout.New()
	sproutHandler.AddRegistry(maps.NewRegistry())

	t, err := template.New("report.html").Funcs(sproutHandler.Build()).Parse(reportTemplate)
	if err != nil {
		log.Panic(err)
	}
	http.HandleFunc("/proxy", func(w http.ResponseWriter, r *http.Request) {
		if err != nil {
			log.Panic(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		bucketName := r.URL.Query().Get("bucketName")
		objectName := r.URL.Query().Get("objectName")
		if bucketName == "" || objectName == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		s3Endpoint := os.Getenv("AWS_ENDPOINT")
		minioClient, err := minio.New(s3Endpoint, &minio.Options{
			Creds: credentials.NewEnvAWS(),
		})
		if err != nil {
			log.Panic(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		object, err := minioClient.GetObject(context.Background(), bucketName, objectName, minio.GetObjectOptions{})
		if err != nil {
			res := minio.ToErrorResponse(err)
			w.Write([]byte(res.Error()))
			return
		}
		defer object.Close()

		data, err := io.ReadAll(object)
		if err != nil {
			log.Panic(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		suites, err := junit.Ingest(data)
		if err != nil {
			log.Panic(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		t.Execute(w, suites)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err != nil {
			log.Panic(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		inputFile, err := os.ReadFile(input)
		if err != nil {
			log.Panic(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		suites, err := junit.Ingest(inputFile)
		if err != nil {
			log.Panic(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		err = t.Execute(w, suites)
		if err != nil {
			log.Panic(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
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
	default:
		inline()
	}
}
