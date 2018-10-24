package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bernardoVale/downscaler/internal/types"
	"github.com/go-redis/redis"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	// FormatHeader name of the header used to extract the format
	FormatHeader = "X-Format"

	// CodeHeader name of the header used as source of the HTTP statu code to return
	CodeHeader = "X-Code"

	// ContentType name of the header that defines the format of the reply
	ContentType = "Content-Type"

	// OriginalURI name of the header with the original URL from NGINX
	OriginalURI = "X-Original-URI"

	// Namespace name of the header that contains information about the Ingress namespace
	Namespace = "X-Namespace"

	// IngressName name of the header that contains the matched Ingress
	IngressName = "X-Ingress-Name"

	// ServiceName name of the header that contains the matched Service in the Ingress
	ServiceName = "X-Service-Name"

	// ServicePort name of the header that contains the matched Service port in the Ingress
	ServicePort = "X-Service-Port"

	// ErrFilesPathVar is the name of the environment variable indicating
	// the location on disk of files served by the handler.
	ErrFilesPathVar = "ERROR_FILES_PATH"

	//HostName name of the header that contains the host directive in the Ingress
	HostName = "X-Hostname"

	// Schema name of the header that contains the request schema
	Schema = "X-Schema"
)

func init() {
	viper.SetEnvPrefix("BACKEND")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetDefault("host", "127.0.0.1:6379")
	viper.SetDefault("password", "")
	viper.AutomaticEnv()
}

func main() {

	backendHost := viper.GetString("host")
	backendPassword := viper.GetString("password")
	await := newAwakingApps()

	logrus.Info("Starting default backend")
	client := redis.NewClient(&redis.Options{
		Addr:     backendHost,
		Password: backendPassword,
		DB:       0, // use default DB
	})
	err := client.Ping().Err()
	if err != nil {
		logrus.WithError(err).Panicf("Could not ping backend")
		panic(err)
	}

	http.HandleFunc("/", errorHandler(client, await))

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.ListenAndServe(fmt.Sprintf(":8080"), nil)
}

func waitForIt(client *redis.Client, namespace, ingress string, await *awakingApps, logger *logrus.Entry) {
	logger.Infof("Waiting for awake state of app %s/%s", namespace, ingress)
	timeout := time.After(15 * time.Minute)
	tick := time.Tick(time.Second * 2)
	app := types.App(fmt.Sprintf("%s/%s", namespace, ingress))
	for {
		select {
		case <-tick:
			val, err := client.Get(app.Key()).Result()
			if err != nil {
				logger.WithError(err).Errorf("Failed to get app status: %v", err)
			}
			if val == "awake" {
				close(await.state[app.String()].redirect)
				return
			}
		case <-timeout:
			logger.Infof("Timeout while waiting for app %s/%s", namespace, ingress)
			close(await.state[app.String()].timeout)
			return
		}
	}
}

func errorHandler(client *redis.Client, await *awakingApps) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// ext := "html"

		ingressName := r.Header.Get(IngressName)
		namespace := r.Header.Get(Namespace)
		schema := r.Header.Get(Schema)
		hostname := r.Header.Get(HostName)
		uri := r.Header.Get(OriginalURI)
		originalURL := fmt.Sprintf("%s://%s%s", schema, hostname, uri)

		app := types.App(fmt.Sprintf("%s/%s", namespace, ingressName))
		u1 := uuid.NewV4()
		logger := logrus.WithFields(logrus.Fields{"id": u1, "app": app})

		if ingressName != "" {
			val, err := client.Get(app.Key()).Result()
			if err != nil {
				if err != redis.Nil {
					panic(err)
				}
				logger.Infof("Key no present on redis, app wasn't put to sleep by downscaler: %s", app.Key())
				fmt.Fprintf(w, "App %v is down - Error 503", app)
				return
			}
			switch val {
			case "sleeping":
				err := client.Publish("wakeup", fmt.Sprintf("%s/%s", namespace, ingressName)).Err()
				if err != nil {
					logger.Errorf("Failed to publish wakeup message: %v", err)
				}
				logger.Info("App is sleeping")
				registered := await.registerApp(app.String())
				if registered {
					logger.Info("Go wait for it")
					go waitForIt(client, namespace, ingressName, await, logger)
				}

				select {
				case <-await.state[app.String()].redirect:
					logger.Info("Got redirect request")
					await.delete(app.String())
					http.Redirect(w, r, originalURL, http.StatusSeeOther)
					return
				case <-await.state[app.String()].timeout:
					logger.Info("Got timeout")
					await.delete(app.String())
				}
			case "waking_up":
				logger.Info("app is waking_up")
				registered := await.registerApp(app.String())
				if registered {
					logger.Info("Registering on waking_up")
					go waitForIt(client, namespace, ingressName, await, logger)
				}
				select {
				case <-await.state[app.String()].redirect:
					logger.Info("Got redirect request")
					await.delete(app.String())
					http.Redirect(w, r, originalURL, http.StatusSeeOther)
					return
				case <-await.state[app.String()].timeout:
					logger.Info("Got timeout")
					await.delete(app.String())
				}
			case "awake":
				logger.Infof("App is awake, redirecting request")
				time.Sleep(500 * time.Millisecond)
				http.Redirect(w, r, originalURL, http.StatusSeeOther)
				return
			}
		}
		// Not collected by custom error
		fmt.Fprintf(w, "Page not found - 404")
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
