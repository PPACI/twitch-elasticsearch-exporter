package main

import (
	"context"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/nicklaw5/helix"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/oauth2/twitch"
	"net/http"
	"time"
)

type config struct {
	HelixClientId     string `env:"HELIX_CLIENT_ID"`
	HelixClientSecret string `env:"HELIX_CLIENT_SECRET"`
	TwitchLanguage    string `env:"TWITCH_LANGUAGE" envDefault:"fr"`
	EsAddonUri        string `env:"ES_ADDON_URI"`
	EsAddonUser       string `env:"ES_ADDON_USER"`
	EsAddonPassword   string `env:"ES_ADDON_PASSWORD"`
	EsIndexPrefix     string `env:"ES_INDEX_PREFIX" envDefault:"streams"`
	LogVerbose        bool   `env:"LOG_VERBOSE" envDefault:"false"`
	HttpPort          int    `env:"PORT" envDefault:"8080"`
}

var c config

func init() {
	opts := env.Options{RequiredIfNoDef: true}
	if err := env.Parse(&c, opts); err != nil {
		log.Fatal(err)
	}
}

func main() {
	if c.LogVerbose {
		log.SetLevel(log.DebugLevel)
	}
	log.Infoln("Init StreamDB")
	streamDB := newStreamDB(c)
	log.Infoln("Init Helix Client")
	helixClient := newHelixClient(c)
	log.Infoln("Start Polling Loop")
	go pollStreamLoop(helixClient, streamDB, c)

	//Handling healthcheck
	http.HandleFunc("/health", func(writer http.ResponseWriter, request *http.Request) {
		_, err := writer.Write([]byte("OK"))
		if err != nil {
			log.Fatal(err)
		}
	})
	log.Infof("Server listening on localhost:%d", c.HttpPort)
	err := http.ListenAndServe(fmt.Sprintf(":%d", c.HttpPort), nil)
	if err != nil {
		log.Fatal(err)
	}
}

func pollStreamLoop(helixClient *helix.Client, streamDB *streamDB, c config) {
	// Force execution now
	err := pollStream(helixClient, streamDB, c.EsIndexPrefix, c.TwitchLanguage)
	if err != nil {
		log.Errorln(err)
	}
	// Start the real polling loop
	for next := range time.Tick(30 * time.Second) {
		err := pollStream(helixClient, streamDB, c.EsIndexPrefix, c.TwitchLanguage)
		if err != nil {
			log.Errorln(err)
		}
		log.Infof("Polling is done. Waiting unil %v", next)
	}
}

func pollStream(helixClient *helix.Client, streamDB *streamDB, esIndex string, language string) error {
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
		indexStream, err := streamDB.IndexStream(IndexStream, esIndex)
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

func newHelixClient(c config) *helix.Client {
	oauth2Config := &clientcredentials.Config{
		ClientID:     c.HelixClientId,
		ClientSecret: c.HelixClientSecret,
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

func newStreamDB(c config) *streamDB {
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{c.EsAddonUri},
		Username:  c.EsAddonUser,
		Password:  c.EsAddonPassword,
	})
	if err != nil {
		log.Fatalln(err)
	}
	streamClient := &streamDB{client}
	return streamClient
}
