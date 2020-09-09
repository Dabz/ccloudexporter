-include ../ccloud_exporter.env
# ccloudexporter deployment
CCLOUD_NS ?= ccloudexporter

install: k8s_deploy_ccloudexporter k8s_prom_apply
remove: k8s_delete_ns k8s_prom_delete

k8s_deploy_ccloudexporter: k8s_create_ns k8s_secret k8s_apply

k8s_create_ns:
	-kubectl create ns ${CCLOUD_NS}

k8s_delete_ns:
	kubectl delete ns ${CCLOUD_NS}

k8s_secret_single:
	kubectl -n ${CCLOUD_NS} create secret generic ccloud-exporter-vars --from-env-file ../ccloud_exporter.env

k8s_secret:
	-$(foreach CCLOUD_CLUSTER, $(CCLOUD_CLUSTERS), kubectl -n ${CCLOUD_NS} create secret generic $(CCLOUD_CLUSTER)-vars --from-literal=CCLOUD_API_KEY=${CCLOUD_API_KEY} --from-literal=CCLOUD_API_SECRET=${CCLOUD_API_SECRET} --from-literal=CCLOUD_CLUSTER=$(CCLOUD_CLUSTER) ; )

k8s_apply_single:
	kubectl -n ${CCLOUD_NS} apply --record -f ccloudexporter.dp.yaml
	kubectl -n ${CCLOUD_NS} apply --record -f ccloudexporter.svc.yaml

k8s_apply: k8s_apply_dp k8s_apply_svc

k8s_apply_dp:
	-$(foreach CCLOUD_CLUSTER, $(CCLOUD_CLUSTERS), sed "s/#{cluster}#/$(CCLOUD_CLUSTER)/g" ccloudexporter.dp.yaml | kubectl -n ${CCLOUD_NS} apply --record -f - ; )

k8s_apply_svc:
	-$(foreach CCLOUD_CLUSTER, $(CCLOUD_CLUSTERS), sed "s/#{cluster}#/$(CCLOUD_CLUSTER)/g" ccloudexporter.svc.yaml | kubectl -n ${CCLOUD_NS} apply --record -f - ; )

# debug
k8s_get:
	watch kubectl -n ${CCLOUD_NS} get pod,deployment,service,secret

k8s_ccloudexporter_logs:
	kubectl -n ${CCLOUD_NS} logs $(shell kubectl -n ${CCLOUD_NS} get pod | grep -i cloud | cut -f1 -d" ") ccloudexporter -f



# prometheus operator deployment
PROM_NS ?= monitoring
k8s_prom_ui:
	kubectl -n ${PROM_NS} port-forward prometheus-${PROM_NS}-prometheus-oper-prometheus-0 9091:9090

k8s_prom_apply_single:
	kubectl -n ${PROM_NS} apply --record -f ccloudexporter.crd.yaml

k8s_prom_apply:
	-$(foreach CCLOUD_CLUSTER, $(CCLOUD_CLUSTERS), sed "s/#{cluster}#/$(CCLOUD_CLUSTER)/g" ccloudexporter.crd.yaml | kubectl -n ${PROM_NS} apply --record -f - ; )

k8s_prom_delete:
	-$(foreach CCLOUD_CLUSTER, $(CCLOUD_CLUSTERS), sed "s/#{cluster}#/$(CCLOUD_CLUSTER)/g" ccloudexporter.crd.yaml | kubectl -n ${PROM_NS} delete -f - ; )
