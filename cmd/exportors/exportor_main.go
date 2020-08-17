package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/linclaus/mantis-opeartor/pkg/db"
	"github.com/linclaus/mantis-opeartor/pkg/server"
)

type Args struct {
	MetricsAddr      string
	Debug            bool
	ElasticsearchUrl string
	DryRun           bool
}

func main() {
	args := Args{
		ElasticsearchUrl: os.Getenv("ELASTICSEARCH-URL"),
		MetricsAddr:      os.Getenv("METRICS-ADDR"),
	}
	// flag.StringVar(&args.MetricsAddr, "listen-address", ":8080", "The address to listen on for HTTP requests.")
	flag.BoolVar(&args.Debug, "debug", true, "debug or not.")
	flag.BoolVar(&args.DryRun, "dryrun", false, "uses a null db driver that writes received webhooks to stdout")

	flag.Parse()

	var driver db.Storer
	if args.DryRun {
		log.Println("dry-run")
		driver = db.NullDB{}
	} else {
		elasticUrls := strings.Split(args.ElasticsearchUrl, ",")
		driver, _ = db.ConnectES(elasticUrls)
	}
	driver.GetVersion()

	s := server.New(args.Debug, driver)
	s.Start(args.MetricsAddr)
}
