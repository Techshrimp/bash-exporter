package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gree-gorey/bash-exporter/pkg/run"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	verbMetrics *prometheus.GaugeVec
)

func main() {
	addr := flag.String("web.listen-address", ":9300", "Address on which to expose metrics")
	interval := flag.Int("interval", 300, "Interval for metrics collection in seconds")
	path := flag.String("path", "/usr/local/bash-exporter/run.sh", "path to script")
	prefix := flag.String("prefix", "bash", "Prefix for metrics")
	debug := flag.Bool("debug", false, "Debug log level")
	flag.Parse()
	pathArray := strings.Split(*path, "/")
	name := strings.Split(pathArray[len(pathArray)-1], ".")[0]
	verbMetrics = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: fmt.Sprintf("%s_%s", *prefix, name),
			Help: "tests",
		},
		[]string{"verb"},
	)
	prometheus.MustRegister(verbMetrics)
	http.Handle("/metrics", prometheus.Handler())
	go Run(int(*interval), *path, *debug)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func Run(interval int, path string, debug bool) {
	for {
		var wg sync.WaitGroup
		o := run.Output{}
		wg.Add(1)
		p := run.Params{UseWg: true, Wg: &wg, Path: &path}
		go o.RunJob(&p)
		wg.Wait()
		if debug == true {
			ser, err := json.Marshal(o)
			if err != nil {
				log.Println(err)
			}
			log.Println(string(ser))
		}
		verbMetrics.Reset()
		for metric, value := range o.Result {
			verbMetrics.With(prometheus.Labels{"verb": metric}).Set(float64(value))
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}
}
