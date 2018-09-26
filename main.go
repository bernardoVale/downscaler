package main

import (
	"context"
	"sync"

	"github.com/bernardoVale/downscaler/backend"
	"github.com/sirupsen/logrus"
)

func main() {

	ctx := context.Background()
	logrus.Info("Estabilishing connection with backend")
	redis := backend.NewRedisClient("127.0.0.1:6379", "rKOsaUDIRK", "wakeup")
	redisAwake := backend.NewRedisClient("127.0.0.1:6379", "rKOsaUDIRK", "awake")
	defer redis.Close()

	prometheus := NewPrometheusClient()
	clientSet := mustAuthenticate()

	// stop := newStopChan()

	// clientSet.AppsV1beta2().Deployments("jenkins").Watch(metav1.ListOptions{})
	// metav1.SingleObject()

	// watchlist := cache.NewListWatchFromClient(clientSet.AppsV1beta2().RESTClient(), "deployments", "jenkins", fields.Everything())
	// _, controller := cache.NewInformer(watchlist, &v1beta2.Deployment{}, time.Second*1, cache.ResourceEventHandlerFuncs{
	// 	UpdateFunc: func(o, n interface{}) {
	// 		// logrus.Infof("Pod updated")
	// 		deployment := n.(*v1beta2.Deployment)

	// 		logrus.Info("Available: %d", deployment.Status.AvailableReplicas)
	// 		logrus.Info("Ready: %d", deployment.Status.ReadyReplicas)

	// 		if deployment.Status.AvailableReplicas == deployment.Status.ReadyReplicas {
	// 			close(stop.c)
	// 			return
	// 		}
	// 		// if newPod.Status.Phase == .PodRunning {
	// 		// 	logrus.Infof("Pod is running")
	// 		// 	close(stop.c)
	// 		// 	return
	// 		// }
	// 		//do something with the updated pod
	// 	},
	// })

	// controller.Run(stop.c)

	// logrus.Info("We're done")

	kubeClient := KubernetesClient{clientSet}

	go sleeper(ctx, redis, prometheus, kubeClient)
	go wakeup(ctx, redis, kubeClient)
	go awaker(ctx, redisAwake, kubeClient)

	<-ctx.Done()
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

type stopChan struct {
	c chan struct{}
	sync.Once
}

func newStopChan() *stopChan {
	return &stopChan{c: make(chan struct{})}
}

func (s *stopChan) closeOnce() {
	s.Do(func() {
		close(s.c)
	})
}
