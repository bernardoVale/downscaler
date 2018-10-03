package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bernardoVale/downscaler/backend"
	"github.com/sirupsen/logrus"
)

func statusHandler(lister backend.Lister) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		data, err := lister.List("sleeping")

		if err != nil {
			logrus.WithError(err).Errorf("Could not list backend keys")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Could not list backend keys.")
			return
		}
		w.WriteHeader(http.StatusOK)

		for k, v := range data {
			app := strings.Split(k, ":")
			fmt.Fprintf(w, "%s/%s=%s\n", app[1], app[2], v)
		}
	}
}
