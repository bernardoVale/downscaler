package main

import (
	"context"
	"flag"
	"os"
	"strings"
	"time"

	"net/http"

	"github.com/bernardoVale/downscaler/internal/kube"
	"github.com/bernardoVale/downscaler/internal/metrics"
	"github.com/bernardoVale/downscaler/internal/storage"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	configPaths = []string{
		".",
		"/etc/downscaler",
		"$HOME/downscaler",
	}
	reconciliatorCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name:      "count",
		Subsystem: "reconciliator",
		Namespace: "downscaler",
		Help:      "Counts reconciliator actions",
	})
	reconciliatorErrCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name:      "errors",
		Subsystem: "reconciliator",
		Namespace: "downscaler",
		Help:      "Counts reconciliator errors",
	})
	wakingUpCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name:      "count",
		Subsystem: "wakingup",
		Namespace: "downscaler",
		Help:      "Counts waking up actions",
	})
	wakingUpErr = prometheus.NewCounter(prometheus.CounterOpts{
		Name:      "errors",
		Subsystem: "wakingup",
		Namespace: "downscaler",
		Help:      "Counts waking up actions",
	})
	awakeCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name:      "count",
		Subsystem: "awake",
		Namespace: "downscaler",
		Help:      "Counts awake actions",
	})
	awakeErr = prometheus.NewCounter(prometheus.CounterOpts{
		Name:      "errors",
		Subsystem: "awake",
		Namespace: "downscaler",
		Help:      "Counts awake actions",
	})
	awakeTimeout = prometheus.NewCounter(prometheus.CounterOpts{
		Name:      "timeout",
		Subsystem: "awake",
		Namespace: "downscaler",
		Help:      "Counts awake timeout",
	})
	sleepingCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name:      "count",
		Subsystem: "sleeping",
		Namespace: "downscaler",
		Help:      "Counts sleeper actions",
	})
	sleepingErr = prometheus.NewCounter(prometheus.CounterOpts{
		Name:      "errors",
		Subsystem: "sleeping",
		Namespace: "downscaler",
		Help:      "Counts sleeper actions",
	})

	sleepingGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:      "sleeping",
		Subsystem: "apps",
		Namespace: "downscaler",
		Help:      "Apps sleeping",
	})
)

func init() {
	viper.SetEnvPrefix("DOWNSCALER")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetDefault("backend.host", "127.0.0.1:6379")
	viper.SetDefault("backend.password", "")
	viper.SetDefault("metrics.host", "http://localhost:9090")
	viper.SetDefault("sleeper.interval", "4h")
	viper.SetDefault("sleeper.max.idle", "10h")
	viper.SetDefault("debug", false)
	viper.SetDefault("metrics.expvars.bind", ":9090")

	viper.SetConfigName("downscaler")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()

	for _, p := range configPaths {
		viper.AddConfigPath(p)
	}
	mustWithMsg(viper.ReadInConfig(), "Could not read config")

	prometheus.MustRegister(reconciliatorCounter)
	prometheus.MustRegister(reconciliatorErrCounter)
	prometheus.MustRegister(sleepingGauge)
	prometheus.MustRegister(wakingUpCounter)
	prometheus.MustRegister(wakingUpErr)
	prometheus.MustRegister(awakeCounter)
	prometheus.MustRegister(awakeErr)
	prometheus.MustRegister(awakeTimeout)
	prometheus.MustRegister(sleepingCounter)
	prometheus.MustRegister(sleepingErr)
}

func main() {
	help := flag.Bool("h", false, "Show help")
	flag.Parse()

	if *help {
		os.Stderr.WriteString("All configuration is done using downscaler.yaml configuration files.\n")
		os.Stderr.WriteString("")
		os.Stderr.WriteString("You can save the files at:\n")
		for _, p := range configPaths {
			os.Stderr.WriteString(p)
			os.Stderr.WriteString("\n")
		}
		os.Stderr.WriteString("You can also configure it using env variable.\n")
		os.Stderr.WriteString("All variables starts with DOWNSCALER. Example: DOWNSCALER_BACKEND_HOST")
		flag.Usage()
		os.Exit(1)
	}

	backendHost := viper.GetString("backend.host")
	backendPassword := viper.GetString("backend.password")
	prometheusURL := viper.GetString("metrics.host")
	sleeperInterval := viper.GetString("sleeper.interval")
	sleeperMaxIdle := viper.GetString("sleeper.max.idle")

	interval, err := time.ParseDuration(sleeperInterval)
	mustWithMsg(err, "Could not parse sleeper.internal. Use a valid interval such as 10s, 2h, 1d")
	_, err = time.ParseDuration(sleeperMaxIdle)
	mustWithMsg(err, "Could not parse sleeper.max.idle. Use a valid interval such as 10s, 2h, 1d")

	ctx := context.Background()
	sleeperConfig := sleeperConfig{
		sleepAfter:      sleeperMaxIdle,
		sleeperInterval: interval,
	}

	redis := storage.NewRedisClient(backendHost, backendPassword, "wakeup")
	defer redis.Close()

	//abscure code, if metrics.host is a file use fakeMetrics
	var metricsClient metrics.Client
	if _, err := os.Stat(prometheusURL); os.IsNotExist(err) {
		metricsClient = metrics.NewPrometheusClient(ctx, prometheusURL)
	} else {
		logrus.Info("Creating new fake metrics client")
		metricsClient = metrics.NewFakeMetricsClient(prometheusURL)
	}

	kubeClient, err := kube.NewKubernetesClient()
	mustWithMsg(err, "Failed to create a Kubernetes client")

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		mustWithMsg(http.ListenAndServe(viper.GetString("metrics.expvars.bind"), nil), "could not expose metrics")
	}()

	// Reconciliate first
	reconciliate(ctx, redis, kubeClient)

	// if nginx-ingress metrics aren't being collected by prometheus
	// sleeper will think that all ingresses are inactive and will
	// put all apps to sleep. This method checks if the metric is available
	// before we start sleeper and wakeuper
	mustWithMsg(waitForPrometheusMetric(metricsClient), "Failed to check essential prometheus metric")

	go sleeper(sleeperConfig, redis, metricsClient, kubeClient)
	go wakeuper(ctx, redis, kubeClient)

	<-ctx.Done()
}

func mustWithMsg(err error, message string) {
	if err != nil {
		logrus.WithError(err).Panicf(message)
		panic(err)
	}
}
