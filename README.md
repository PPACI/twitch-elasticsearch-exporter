# twitch-elasticsearch-exporter

Poll top 100 streams of specified language every 30s, and store them in the desired ElasticSearch index.

# Get Started

## Init an ES index template

```shell
curl -X PUT -H "Content-Type: application/json" -d @elasticsearch/index-template.json http://localhost:9200/_index_template/streams
```

## Build 

```shell
go build .
```

## Start

```shell
./twitch-surveillance --help
```