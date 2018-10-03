package backend

import (
	"time"

	"github.com/rusenask/k8s-kv/kv"
)

type KubernetesClient struct {
	base *kv.KV
}

func (cli KubernetesClient) Retrieve(key string) (string, error) {
	data, err := cli.base.Get(key)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (cli KubernetesClient) Post(key string, value string, ttl time.Duration) error {
	return cli.base.Put(key, []byte(value))
}

func (cli KubernetesClient) List(pattern string) (map[string]string, error) {
	kv := make(map[string]string)
	data, err := cli.base.List(pattern)
	if err != nil {
		return make(map[string]string), err
	}
	for k, v := range data {
		kv[k] = string(v)
	}
	return kv, nil
}

func NewKubernetesClient(backend *kv.KV) KubernetesClient {
	return KubernetesClient{base: backend}
}
