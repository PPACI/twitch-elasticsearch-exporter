package main

import (
	"encoding/json"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/nicklaw5/helix"
	"strings"
	"time"
)

const (
	esIndex = "streams"
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

func (s streamDB) IndexStream(stream helix.Stream) (*indexResult, error) {
	ts := timeStream{stream, time.Now()}
	body, err := json.Marshal(ts)
	if err != nil {
		return &indexResult{}, err
	}
	response, err := s.Index(esIndex, strings.NewReader(string(body)))
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
