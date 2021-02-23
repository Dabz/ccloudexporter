This directory contains provisioning artifacts and templates for adding a pre-built Grafana dashboards and datasource (Prometheus) to Grafana.
These are automatically [provisioned](https://grafana.com/docs/grafana/latest/administration/provisioning/) in the Grafana instance started if using Docker Compose.

- `ccloud-exporter.json`: JSON definition of Grafana dashboard, suitable for importing into the Grafana UI.
- `provisioning/dashboards/ccloud-exporter.json`: Same dashboard template, but with the script `create-provisioning.sh` run over it, to replace the variable for the Prometheus datasource name with a real datasource name.  Grafana does not yet support [provisioning from templates that contain variables](https://github.com/grafana/grafana/issues/10786).
- `provisioning/datasources/prometheus_ccloudexporter.yaml`: Provisions a standard Prometheus datasource for the provisioned dashboard to connect to.

![Grafana](/grafana/grafana.png)
