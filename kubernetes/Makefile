# ccloudexporter deployment
CCLOUD_NS ?= ccloudexporter

all: install k8s_prom_apply
install: k8s_create_ns k8s_secret k8s_apply

k8s_create_ns:
	kubectl create ns ${CCLOUD_NS}

k8s_delete_ns:
	kubectl delete ns ${CCLOUD_NS}

k8s_secret:
	kubectl -n ${CCLOUD_NS} create secret generic ccloud-exporter-vars --from-env-file ../ccloud_exporter.env

k8s_apply:
	kubectl -n ${CCLOUD_NS} apply --record -f ccloudexporter.dp.yaml
	kubectl -n ${CCLOUD_NS} apply --record -f ccloudexporter.svc.yaml

# debug
k8s_get:
	watch kubectl -n ${CCLOUD_NS} get pod,deployment,service,secret

k8s_ccloudexporter_logs:
	kubectl -n ${CCLOUD_NS} logs $(shell kubectl -n ${CCLOUD_NS} get pod | grep -i cloud | cut -f1 -d" ") ccloudexporter -f



# prometheus operator deployment
PROM_NS ?= monitoring
k8s_prom_ui:
	kubectl -n ${PROM_NS} port-forward prometheus-${PROM_NS}-prometheus-oper-prometheus-0 9091:9090

k8s_prom_apply:
	kubectl -n ${PROM_NS} apply --record -f ccloudexporter.crd.yaml

k8s_prom_delete:
	kubectl -n ${PROM_NS} delete -f ccloudexporter.crd.yaml
