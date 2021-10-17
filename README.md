# twitch-elasticsearch-exporter

Poll top 100 streams of specified language every 30s, and store them in the desired ElasticSearch index.

**Stream with less than 500 followers will be ignored**.

Beware, the code is full of hard coded values and potential bad practices. This is a personal project before anything else 🙃.

## Get Started

## Configuration

|Name|Default|Description|
|----|-------|-----------|
|HELIX_CLIENT_ID||Twitch helixClient ID|
|HELIX_CLIENT_SECRET||Twitch helixClient secret|
|TWITCH_LANGUAGE|fr|Code of the twitch language to monitor|
|ES_ADDON_URI||URL of Elasticsearch server|
|ES_ADDON_USER||User to connect to Elasticsearch|
|ES_ADDON_PASSWORD||Password to connect to Elasticsearch|
|ES_INDEX_PREFIX|streams|Elasticsearch index prefix|
|LOG_VERBOSE|false|Enable debug log|

### Elasticsearch index template

This app use [ElasticSearch Index Template](https://www.elastic.co/guide/en/elasticsearch/reference/current/index-templates.html)
to create an index per month.

An index template is provided and can be deployed as:

```shell
curl -X PUT -H "Content-Type: application/json" -d @elasticsearch/index-template.json http://localhost:9200/_index_template/streams
```


## Grafana Dashboard

An analytic Grafana Dashboard using Elasticsearch is available in the `grafana` folder.
