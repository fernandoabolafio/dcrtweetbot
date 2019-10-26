package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/decred/dcrd/dcrutil"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	ipfs "github.com/ipfs/go-ipfs-api"
)

// IPFS shell
var shell *ipfs.Shell
var client *twitter.Client
var config *Config

const (
	defaultConfigFilename = "gostampbot.conf"
)

var (
	defaultHomeDir = dcrutil.AppDataDir("gostampbot", false)

	// DefaultConfigFile points to politeiawww's default config file.
	defaultConfigFile  = filepath.Join(defaultHomeDir, defaultConfigFilename)
	defaultTargetWords = []string{"@gostampbot"}
)

// Config describe the config options available for the program
type Config struct {
	TwitAPIConsumerKey    string
	TwitAPIConsumerSecret string
	TwitAPIToken          string
	TwitAPITokenSecret    string
	TargetWords           []string
	ConfigFile            string
}

func loadConfig() {
	// parse cli arguments to check for a different config file
	cfgFile := flag.String("configfile", defaultConfigFile, "the config file to be used")

	config = &Config{
		ConfigFile:  *cfgFile,
		TargetWords: defaultTargetWords,
	}

	log.Println("Reading config file from: ", config.ConfigFile)
	file, err := os.Open(config.ConfigFile)
	if err != nil {
		log.Println("Could not open config file. Skipping. Error:", err)
	}

	if file != nil {
		// @todo: Parse and get options from config file
	}

	// override cfg options with cli options
	twitCKey := flag.String("consumerkey", "", "Twitter API Consumer Key (required)")
	flag.Parse()

	// @todo: finish checking for config arguments
	// twitCSecret := flag.String("consumersecret", "", "Twitter API Consumer Secret")
	// twitToken := flag.String("token", "", "Twitter API Token");
	// twitTokenSecret := flag.String("tokensecret", "", "Twitter API Token Secret")

	if config.TwitAPIConsumerKey == "" && (*twitCKey == "") {
		log.Printf("Invalid configuration. Check valid options below: \n")
		flag.PrintDefaults()
		os.Exit(1)
	}

}

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

func createTwitterClient() *twitter.Client {
	//@todo: get twitter api client params from config
	config := oauth1.NewConfig("gb6CAlY1LLe3PL8UjFmKeMg0W", "C5Uy9ooeIyupNTFwkro8RIukgWPL4f1cQgONc4Xl0kcpfWBDxD")
	token := oauth1.NewToken("903809097664954369-nRdvR7RUQ3QN0OE0XC5G6QcYK125its", "SfSdwDC20FMHOEcC2UF199roLS8DXjPe26EQxebrnTQ62")
	httpClient := config.Client(oauth1.NoContext, token)

	return twitter.NewClient(httpClient)
}

func createIPFSShell() *ipfs.Shell {
	// @todo: create a config field to hold the ipfs shell host and
	// replace it
	return ipfs.NewShell("localhost:5001")
}

func createTwitterStream() (*twitter.Stream, error) {
	params := &twitter.StreamFilterParams{
		Track:         []string{"yhgu37"},
		StallWarnings: twitter.Bool(true),
	}

	return client.Streams.Filter(params)
}

func listenToTweets(stream *twitter.Stream, handler func(tweet *twitter.Tweet)) {
	demux := twitter.NewSwitchDemux()
	demux.Tweet = handler
	go demux.HandleChan(stream.Messages)
}

func main() {
	// appName := filepath.Base(os.Args[0])

	loadConfig()

	client = createTwitterClient()
	shell = createIPFSShell()

	stream, err := createTwitterStream()
	if err != nil {
		fmt.Println("stream error: ", err)
	}

	listenToTweets(stream, func(tweet *twitter.Tweet) {
		// @todo: validate tweets with regex pattern
		// @todo: timestamp tweets
		// @todo: reply to tweet thread and dm author
		fmt.Println(tweet.Text)
		cid, err := storeOnIPFS(tweet)
		if err != nil {
			log.Println("ipfs failed: ", err)
		} else {
			log.Println("ipfs oK", cid)
		}
	})

	log.Println("Start of the day !")

	// listen to stop signal and stop the stream before exiting
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)

	stream.Stop()
}
