package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/nicklaw5/helix"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/oauth2/twitch"
	"os"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func main() {
	clientId := flag.String("client-id", "", "Twitch client ID")
	clientSecret := flag.String("client-secret", "", "Twitch client secret")
	flag.Parse()
	if *clientId == "" || *clientSecret == "" {
		fmt.Println("Twitch Surveillance")
		fmt.Println("Usage:")
		flag.PrintDefaults()
		os.Exit(1)
	}
	oauth2Config := &clientcredentials.Config{
		ClientID:     *clientId,
		ClientSecret: *clientSecret,
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
	streams, err := client.GetStreams(&helix.StreamsParams{First: 100, Language: []string{"fr"}})
	if err != nil {
		log.Fatalln(err)
	}
	if streams.StatusCode != 200 {
		log.Fatalln(streams.ErrorMessage)
	}
	log.Debugf("Got %v streams\n", len(streams.Data.Streams))
	for _, stream := range streams.Data.Streams {
		log.Infof("%+v\n", stream)
	}
}
