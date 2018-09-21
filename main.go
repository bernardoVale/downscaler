package main

import (
	"context"

	"github.com/bernardoVale/downscaler/backend"
	"github.com/sirupsen/logrus"
)

func sleepCandidates(active map[string]int, all []string) []string {
	for i, ingress := range all {
		if _, ok := active[ingress]; ok {
			// Remove active ingresses from sleep candidates
			all[i] = all[len(all)-1]
			all[len(all)-1] = ""
			all = all[:len(all)-1]
		}
	}
	logrus.Info("Total candidates ", len(all))
	return all
}

func checkPrometheusMetrics(ctx context.Context, collector IngressCollector) map[string]int {
	//"rate(nginx_ingress_controller_requests{status=\"200\"}[12h])"
	results, err := collector.getIngresses(ctx, "sum(rate(nginx_ingress_controller_requests{status=\"200\"}[12h])) by (ingress,exported_namespace)")
	if err != nil {
		logrus.Errorf("Could not check prometheus metrics:%v", err)
	}
	return results
}

func retrieveKubernetesIngresses(ctx context.Context) []string {
	return []string{
		"ac-identity/acidentity-staging", "academy/academy-production", "acdc/acdcholiday-staging", "acdc/acdcrequest-staging", "acdc/acdctimesheet-staging", "acdc/acdctravel-staging", "acdc/acdcvacation-staging", "acdc-legacy/acdclegacy-staging", "acdc2/acdc2-staging", "acinsight/acinsightui-staging", "acob/aconboarder-staging", "acpm/acpm-staging", "alphaquester/alphaquester-staging", "default/admission", "default/todo-staging", "eba/eba-production", "eba/eba-staging", "jenkins/jenkins-staging", "kube-system/k8s-dashbord", "miles/miles-staging", "miles/milesui-staging", "mule/mule-production", "mule/mule-staging", "parking/parking-production", "parking/parking-staging", "qa-test/qatest", "sso/sso-staging", "superstars/superstars-staging", "superstars/superstarsfront-staging", "website/acms-staging", "website/website-staging",
	}
}

func main() {

	ctx := context.Background()
	logrus.Info("Estabilishing connection with backend")
	redis := backend.NewRedisClient("127.0.0.1:6379", "npCYPR7uAt")

	prometheus := NewPrometheusClient()
	clientSet := mustAuthenticate()
	deploymentsClient := clientSet.AppsV1()
	//
	// prometheus.getIngresses(ctx, )
	checkPrometheusMetrics(ctx, prometheus)
	logrus.Infoln("Starting sleeper process")
	go sleeper(ctx, redis, prometheus, deploymentsClient)

	ctx.Done()
	<-ctx.Done()
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
