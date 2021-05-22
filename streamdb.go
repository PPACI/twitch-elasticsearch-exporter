package main

import (
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/nicklaw5/helix"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

var indexSuffix string

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

type Stream struct {
	helix.Stream
	FollowerCount int       `json:"follower_count"`
	Timestamp     time.Time `json:"@timestamp"`
}

func init() {
	setIndexSuffix()
	go func() {
		for range time.Tick(5 * time.Minute) {
			setIndexSuffix()
		}
	}()
}

func setIndexSuffix() {
	now := time.Now()
	indexSuffix = fmt.Sprintf("%v-%d", now.Year(), now.Month())
}

// IndexStream save a stream to the specified Elasticsearch Index
// This function will add a @timestamp property to the given stream before indexing it.
// This make it compatible with the elasticsearch data stream model.
// Year and Month will be appended to index name such as streams-2021-03
func (s streamDB) IndexStream(stream Stream, index string) (*indexResult, error) {
	index = index + "-" + indexSuffix
	log.Debugln("Indexing stream to", index)
	body, err := json.Marshal(stream)
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
