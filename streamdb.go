package main

import (
	"encoding/json"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/nicklaw5/helix"
	"strings"
	"time"
)

type indexResult struct {
	*esapi.Response
	Body indexBodyResult
}

type indexBodyResult struct {
	Index   string `json:"_index"`
	Id      string `json:"_id"`
	Version int    `json:"_version"`
	Result  string `json:"result"`
}

type streamDB struct {
	*elasticsearch.Client
}

type timeStream struct {
	helix.Stream
	Timestamp time.Time `json:"@timestamp"`
}

// IndexStream save a stream to the specified Elasticsearch Index
// This function will add a @timestamp property to the given stream before indexing it.
// This make it compatible with the elasticsearch data stream model.
func (s streamDB) IndexStream(stream helix.Stream, index string) (*indexResult, error) {
	ts := timeStream{stream, time.Now()}
	body, err := json.Marshal(ts)
	if err != nil {
		return &indexResult{}, err
	}
	response, err := s.Index(index, strings.NewReader(string(body)))
	if err != nil {
		return &indexResult{}, err
	}
	defer response.Body.Close()
	respBody := indexBodyResult{}
	if err := json.NewDecoder(response.Body).Decode(&respBody); err != nil {
		return nil, err
	}
	return &indexResult{
		Response: response,
		Body:     respBody,
	}, nil
}
