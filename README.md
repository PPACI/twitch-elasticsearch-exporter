# twitch-elasticsearch-exporter

Poll top 100 streams of specified language every 30s, and store them in the desired ElasticSearch index.

**Stream with less than 500 followers will be ignored**.

Beware, the code is full of hard coded values and potential bad practices. This is a personal project before anything else 🙃.

## Get Started

### Init an ES index template

```shell
curl -X PUT -H "Content-Type: application/json" -d @elasticsearch/index-template.json http://localhost:9200/_index_template/streams
```

### Build 

```shell
go build .
```

### Start

```shell
./twitch-surveillance --help
```

## Grafana Dashboard

An analytic Grafana Dashboard is available in the `grafana` folder.
