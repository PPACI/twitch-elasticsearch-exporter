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
	"time"
)

func main() {
	clientId := flag.String("client-id", "", "Twitch helixClient ID")
	clientSecret := flag.String("client-secret", "", "Twitch helixClient secret")
	language := flag.String("twitch-language", "fr", "Code of the twitch language to poll like fr, en")
	esUrl := flag.String("elasticsearch-url", "http://localhost:9200", "Comma separated list of url of elasticsearch")
	esIndex := flag.String("elasticsearch-index", "streams", "Elasticsearch index to use")
	verbose := flag.Bool("verbose", false, "Enable verbose mode")
	flag.Parse()
	if *clientId == "" || *clientSecret == "" {
		fmt.Println("Twitch Surveillance")
		fmt.Println("Usage:")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *verbose {
		log.SetLevel(log.DebugLevel)
	}
	log.Infoln("Init StreamDB")
	streamDB := newStreamDB(strings.Split(*esUrl, ","))
	log.Infoln("Init Helix Client")
	helixClient := newHelixClient(*clientId, *clientSecret)
	log.Infoln("Start Polling Loop")
	pollStreamLoop(helixClient, streamDB, esIndex, *language)
}

func pollStreamLoop(helixClient *helix.Client, streamDB *streamDB, esIndex *string, language string) {
	// Force execution now
	err := pollStream(helixClient, streamDB, esIndex, language)
	if err != nil {
		log.Errorln(err)
	}
	// Start the real polling loop
	for next := range time.Tick(30 * time.Second) {
		err := pollStream(helixClient, streamDB, esIndex, language)
		if err != nil {
			log.Errorln(err)
		}
		log.Infof("Polling is done. Waiting unil %v", next)
	}
}

func pollStream(helixClient *helix.Client, streamDB *streamDB, esIndex *string, language string) error {
	streams, err := helixClient.GetStreams(&helix.StreamsParams{First: 100, Language: []string{language}})
	if err != nil {
		return err
	}
	if streams.StatusCode != 200 {
		return err
	}
	log.Debugf("Got %v streams\n", len(streams.Data.Streams))
	stored, skipped := 0, 0
	for _, stream := range streams.Data.Streams {
		log.Debugf("%+v\n", stream)
		streamLog := log.WithField("title", stream.Title).WithField("User", stream.UserName)
		if stream.ViewerCount < 500 {
			streamLog.Debugf("%v followers is less than the minimum of 500\n", stream.ViewerCount)
			skipped++
			continue
		}
		follower, err := getFollower(helixClient, stream.UserID)
		if err != nil {
			return err
		}
		IndexStream := Stream{
			Stream:        stream,
			Timestamp:     time.Now(),
			FollowerCount: follower,
		}
		streamLog.Debugf("Follower fetched. Got %v followers.\n", IndexStream.FollowerCount)
		streamLog.Debugln("Indexing to DB")
		indexStream, err := streamDB.IndexStream(IndexStream, *esIndex)
		if err != nil {
			return err
		}
		if indexStream.IsError() {
			log.WithField("statuscode", indexStream.StatusCode).Fatalf("%+v\n", indexStream)
		}
		fields := log.Fields{"Id": indexStream.Body.Id, "Index": indexStream.Body.Index}
		streamLog.WithFields(fields).Debugln("Stored stream in DB.")
		stored++
	}
	log.Infof("Stored %v data points in DB\n", stored)
	log.Infof("Skipped %v data points due to low viewer count\n", skipped)
	return nil
}

// getFollowerForStreams return the number of follower for a specific userID
func getFollower(helixClient *helix.Client, userID string) (int, error) {
	follows, err := helixClient.GetUsersFollows(&helix.UsersFollowsParams{
		First: 0,
		ToID:  userID,
	})
	if err != nil {
		return 0, err
	}
	return follows.Data.Total, nil
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
