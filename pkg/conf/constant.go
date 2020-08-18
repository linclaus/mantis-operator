package conf

const (
	PROMETHEUS_NAMESPACE          = "moebius-system"
	ALERTMANAGER_SECRET_NAME      = "alertmanager-r-prometheus-operator-alertmanager"
	ALERTMANAGER_SECRET_DATA_NAME = "alertmanager.yaml"
	WEBHOOK_URL                   = "http://mantis-api.moebius-system:8000/api/v2/webhook"
)
