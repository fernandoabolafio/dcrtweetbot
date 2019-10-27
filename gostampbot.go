package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	ipfs "github.com/ipfs/go-ipfs-api"
)

// IPFS shell
var shell *ipfs.Shell
var client *twitter.Client
var config *Config
var count int

func storeOnIPFS(tweet *twitter.Tweet) (string, error) {
	b, err := json.Marshal(tweet)
	v := string(b)
	r := strings.NewReader(v)
	cid, err := shell.Add(r)
	if err != nil {
		return "", err
	}
	return cid, nil
}

func createTwitterClient() (*twitter.Client, error) {
	cfg := oauth1.NewConfig(config.TwitAPIConsumerKey, config.TwitAPIConsumerSecret)
	token := oauth1.NewToken(config.TwitAPIToken, config.TwitAPITokenSecret)
	httpClient := cfg.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	// verify credentials
	params := twitter.AccountVerifyParams{}
	_, _, err := client.Accounts.VerifyCredentials(&params)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func createIPFSShell() (*ipfs.Shell, error) {
	s := ipfs.NewShell(config.IPFSHost)
	_, _, err := s.Version()
	if err != nil {
		return nil, err
	}
	return s, nil
}

func createTwitterStream() (*twitter.Stream, error) {
	params := &twitter.StreamFilterParams{
		Track:         config.TargetWords,
		StallWarnings: twitter.Bool(true),
	}

	return client.Streams.Filter(params)
}

func listenToTweets(stream *twitter.Stream, handler func(tweet *twitter.Tweet)) {
	demux := twitter.NewSwitchDemux()
	demux.Tweet = handler
	go demux.HandleChan(stream.Messages)
}

func trimWords(words []string) []string {
	processedWords := make([]string, len(words))
	for i, word := range words {
		processedWords[i] = strings.TrimSpace(word)
	}
	return processedWords
}

func main() {

	config = loadConfig()

	twitClient, err := createTwitterClient()
	if err != nil {
		log.Fatalf("Could not create a twitter client: %v", err)
	}
	client = twitClient

	shell, err = createIPFSShell()
	if err != nil {
		log.Fatalf("Could not connect to IPFS daemon: %v", err)
	}

	config.TargetWords = trimWords(config.TargetWords)
	stream, err := createTwitterStream()
	if err != nil {
		fmt.Println("stream error: ", err)
	}

	listenToTweets(stream, func(tweet *twitter.Tweet) {
		// @todo: validate tweets with regex pattern
		// @todo: timestamp tweets
		// @todo: reply to tweet thread and dm author
		count++
		fmt.Println(tweet.Text)
		cid, err := storeOnIPFS(tweet)
		if err != nil {
			log.Println("ipfs failed: ", err)
		} else {
			log.Println("ipfs oK", cid)
		}

		log.Println("\n \n ======", count, " TWEETS ======= \n ")
	})

	log.Println(len(config.TargetWords), " tracked words: ", config.TargetWords)
	log.Println("Start of the day!")

	// listen to stop signal and stop the stream before exiting
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)

	stream.Stop()
}
