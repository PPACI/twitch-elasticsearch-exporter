package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/nicklaw5/helix"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/oauth2/twitch"
	"os"
	"strings"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func main() {
	clientId := flag.String("client-id", "", "Twitch helixClient ID")
	clientSecret := flag.String("client-secret", "", "Twitch helixClient secret")
	esUrl := flag.String("elasticsearch-url", "http://localhost:9200", "Comma separated list of url of elasticsearch")
	flag.Parse()
	if *clientId == "" || *clientSecret == "" {
		fmt.Println("Twitch Surveillance")
		fmt.Println("Usage:")
		flag.PrintDefaults()
		os.Exit(1)
	}
	streamDB := newStreamDB(strings.Split(*esUrl, ","))
	helixClient := newHelixClient(*clientId, *clientSecret)
	streams, err := helixClient.GetStreams(&helix.StreamsParams{First: 100, Language: []string{"fr"}})
	if err != nil {
		log.Fatalln(err)
	}
	if streams.StatusCode != 200 {
		log.Fatalln(streams.ErrorMessage)
	}
	log.Debugf("Got %v streams\n", len(streams.Data.Streams))
	stored := 0
	for _, stream := range streams.Data.Streams {
		log.Debugf("%+v\n", stream)
		streamLog := log.WithField("title", stream.Title).WithField("User", stream.UserName)
		streamLog.Debugln("Storing stream in DB.")
		indexStream, err := streamDB.IndexStream(stream)
		if err != nil {
			log.Fatalln(err)
		}
		if indexStream.IsError() {
			log.WithField("statuscode", indexStream.StatusCode).Fatalf("%+v\n", indexStream)
		}
		streamLog.WithField("Id", indexStream.Body.Id).Debugln("Stored stream in DB.")
		stored++
	}
	log.Infof("Stored %v data points in DB\n", stored)
}

func newHelixClient(clientId string, clientSecret string) *helix.Client {
	oauth2Config := &clientcredentials.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		TokenURL:     twitch.Endpoint.TokenURL,
	}
	oauthClient := oauth2Config.Client(context.Background())
	client, err := helix.NewClient(&helix.Options{
		HTTPClient: oauthClient,
		ClientID:   "ts357p9xfog4rouvtia1pm8m6wev43",
	})
	if err != nil {
		log.Fatalln(err)
	}
	return client
}

func newStreamDB(esUrl []string) *streamDB {
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: esUrl,
	})
	if err != nil {
		log.Fatalln(err)
	}
	streamClient := &streamDB{client}
	return streamClient
}
