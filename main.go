package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/putdotio/go-putio"
	"golang.org/x/oauth2"
)

const namespace = "putio"

var (
	// Configuration flags
	oauthToken = flag.String("oauth-token", "",
		"put.io OAUTH Token.")
	listenPort = flag.String("listen-port", "9101",
		"Address to listen on for requests from Prometheus.")
	metricsPath = flag.String("metrics-path", "/metrics",
		"Path to expose metrics on.")
	configPath = flag.String("config-path", "",
		"Path to env file.")

	// Register metrics for collection
	up = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "up"),
		"Was the last putio query successful.",
		nil, nil,
	)
	accountActive = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "account_active"),
		"Is the putio account being queried currently active.",
		nil, nil,
	)
	diskAvail = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "disk_available"),
		"The available disk space in the account (in bytes).",
		nil, nil,
	)
	diskSize = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "disk_size"),
		"The total disk space, available and used, in the account (in bytes).",
		nil, nil,
	)
	diskUsed = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "disk_used"),
		"The used disk space in the account (in bytes).",
		nil, nil,
	)
	simultaneousDownloadLimit = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "simultaneous_download_limit"),
		"The maximum amount of downloads permitted by the account tier.",
		nil, nil,
	)
	expirationDate = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "account_plan_expiration_date"),
		"The unix time when the currently active putio plan expires.", nil, nil,
	)
	transfersList = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "transfers_by_status"),
		"The number of transfers by status type.", []string{"type"}, nil,
	)
)

type Exporter struct {
	putioToken string
}

func NewExporter(putioToken string) *Exporter {
	return &Exporter{
		putioToken: putioToken,
	}
}

func (e *Exporter) getUnixExpirationDate(expirationDate string) int64 {
	layout := "2006-01-02T15:04:05"
	formattedExpDate, err := time.Parse(layout, expirationDate)
	if err != nil {
		fmt.Println(err)
	}
	return formattedExpDate.Unix()
}

func (e *Exporter) collectAccountInfoMetrics(ch chan<- prometheus.Metric, accountInfo putio.AccountInfo) {
	accountActiveInt := 0.0
	if accountInfo.AccountActive {
		accountActiveInt = 1.0
	}
	ch <- prometheus.MustNewConstMetric(
		accountActive, prometheus.GaugeValue, accountActiveInt,
	)
	ch <- prometheus.MustNewConstMetric(
		expirationDate, prometheus.GaugeValue, float64(e.getUnixExpirationDate(accountInfo.PlanExpirationDate)),
	)
	ch <- prometheus.MustNewConstMetric(
		diskAvail, prometheus.GaugeValue, float64(accountInfo.Disk.Avail),
	)
	ch <- prometheus.MustNewConstMetric(
		diskSize, prometheus.GaugeValue, float64(accountInfo.Disk.Size),
	)
	ch <- prometheus.MustNewConstMetric(
		diskUsed, prometheus.GaugeValue, float64(accountInfo.Disk.Used),
	)
	ch <- prometheus.MustNewConstMetric(
		simultaneousDownloadLimit, prometheus.GaugeValue, float64(accountInfo.SimultaneousDownloadLimit),
	)
}

func (e *Exporter) collectTransfersMetrics(ch chan<- prometheus.Metric, transfers []putio.Transfer) {
	transferStatuses := make(map[string]int)
	for i := 0; i < len(transfers); i++ {
		status := transfers[i].Status
		numOfOccurences, statusExists := transferStatuses[status]
		if statusExists {
			transferStatuses[status] = numOfOccurences + 1
		} else {
			transferStatuses[status] = 1
		}
	}

	for status, count := range transferStatuses {
		ch <- prometheus.MustNewConstMetric(
			transfersList, prometheus.GaugeValue, float64(count), status,
		)
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- up
	ch <- accountActive
	ch <- diskAvail
	ch <- diskSize
	ch <- diskUsed
	ch <- simultaneousDownloadLimit
	ch <- transfersList
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: e.putioToken})
	oauthClient := oauth2.NewClient(context.TODO(), tokenSource)
	client := putio.NewClient(oauthClient)
	accountInfo, ai_err := client.Account.Info(context.TODO())
	transfers, t_err := client.Transfers.List(context.TODO())
	if ai_err != nil || t_err != nil {
		ch <- prometheus.MustNewConstMetric(
			up, prometheus.GaugeValue, 0,
		)
		log.Println("Failed to fetch AccountInfo data from putio.")
		log.Println(ai_err)
		return
	} else if t_err != nil {
		ch <- prometheus.MustNewConstMetric(
			up, prometheus.GaugeValue, 0,
		)
		log.Println("Failed to fetch Transfers data from putio.")
		log.Println(t_err)
		return
	} else {
		ch <- prometheus.MustNewConstMetric(
			up, prometheus.GaugeValue, 1,
		)
	}
	e.collectTransfersMetrics(ch, transfers)
	e.collectAccountInfoMetrics(ch, accountInfo)
}

func main() {
	flag.Parse()
	configFile := *configPath
	if configFile != "" {
		log.Printf("Loading %s env file.\n", configFile)
		err := godotenv.Load(configFile)
		if err != nil {
			log.Printf("Error loading %s env file.\n", configFile)
		}
	} else {
		err := godotenv.Load()
		if err != nil {
			log.Println("No .env file found, assume env variables are set.")
		}
		listenPortOverride := os.Getenv("PUTIO_EXPORTER_LISTEN_PORT")
		if listenPortOverride != "" {
			*listenPort = listenPortOverride
		}
		metricsPathOverride := os.Getenv("PUTIO_EXPORTER_METRICS_PATH")
		if metricsPathOverride != "" {
			*metricsPath = metricsPathOverride
		}
	}
	if *oauthToken == "" {
		putioTokenOverride := os.Getenv("PUTIO_OAUTH_TOKEN")
		if putioTokenOverride == "" {
			log.Fatal("Putio OAUTH token was not provided. See the README for required configuration options.")
		}
		*oauthToken = putioTokenOverride
	}

	exporter := NewExporter(*oauthToken)
	prometheus.MustRegister(exporter)

	http.Handle(*metricsPath, promhttp.Handler())
	log.Printf("Listening on :%s%s", *listenPort, *metricsPath)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", *listenPort), nil))
}
