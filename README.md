
# Downscaler

Downscaler implements an [nginx-ingress-controller](https://github.com/kubernetes/ingress-nginx) default backend capable of scaling applications to zero and wake them up when they receive traffic.

It has three main dependencies:

- The nginx-ingress controller
- A redis database
- Prometheus collecting metrics of `nginx-ingress-controller`

The easiest way of using it is by deploying it using our `helm` chart.

# Setup with Helm

The helm chart will install redis and prometheus for you and configure a job to scrap nginx-ingress metrics.

You need to make a couple of changes on `nginx-ingress` though.

- You need to make sure `nginx-ingress` is configured to collect metrics
- You need to configure `nginx-ingress` to send `503` errors to the default backend

Hopefully, if you've deployed `nginx-ingress` using a helm chart, you basically need to change below values:

```bash
TEMP_FILE=$(mktemp)
cat <<EOF > $TEMP_FILE
defaultBackend:
  enabled: false
controller:
  podAnnotations:
    prometheus.io/port: "10254"
    prometheus.io/scrape: "true"
  defaultBackendService: default/downscaler-apps-backend
  image:
    repository: bernardovale/nginx-ingress-controller
    tag: 0.18.0
  stats:
    enabled: true
  metrics:
    enabled: true
  config:
    custom-http-errors: "503"
EOF
helm upgrade --install downscaler-ingress -f $TEMP_FILE stable/nginx-ingress
```

Deploy downscaler configuring the interval (`downscaler.sleeper.interval`) where sleeper will run and put idle apps to sleep and the amount of time we should use to consider an app idle (`downscaler.sleeper.max.idle`).

```
helm upgrade --install downscaler-apps \
deployments/helm/downscaler \
--set downscaler.sleeper.interval=45s \
--set downscaler.sleeper.max.idle=30s
```


# Activating Ingresses

`downscaler` will only put apps to sleep that were labeled `downscaler.active=true`. You need to manually set the
label to all ingresses that you wish to scale to zero.

You can use the above command to set the label to all ingresses of current namespace:

```
kubectl label ingress downscaler.active=true --all
```
