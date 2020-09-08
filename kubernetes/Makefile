NAMESPACE ?= cloudexporter

install: k8s_secret k8s_apply

k8s_create_ns:
	kubectl create ns ${NAMESPACE}

k8s_delete_ns:
	kubectl delete ns ${NAMESPACE}

k8s_secret:
	kubectl -n ${NAMESPACE} create secret generic ccloud-exporter-vars --from-env-file ../ccloud_exporter.env

k8s_apply:
	kubectl -n ${NAMESPACE} apply --record -R -f .

# debug
k8s_get:
	watch kubectl -n ${NAMESPACE} get pod,deployment,service,secret

k8s_cloudexporter_logs:
	kubectl -n ${NAMESPACE} logs $(shell kubectl -n cloudexporter get pod | grep -i cloud | cut -f1 -d" ") cloudexporter -f


# prometheus operator deployment
PROM_NS ?= monitoring
k8s_prom_ui:
	kubectl -n ${PROM_NS} port-forward prometheus-${PROM_NS}-prometheus-oper-prometheus-0 9091:9090
