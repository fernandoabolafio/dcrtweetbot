package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	piutil "github.com/decred/politeia/util"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	ipfs "github.com/ipfs/go-ipfs-api"
)

const numberOfTweetsVisible = 100

var shell *ipfs.Shell
var client *twitter.Client
var config *Config
var dcrtimeHost string
var timestampedTweets = make([]tweetResult, 100)
var count int

type displayTweet struct {
	ID       int64
	Text     string
	UserName string
}

type tweetResult struct {
	Cid     string
	Digest  string
	isReply bool
	Tweet   displayTweet
	Thread  []displayTweet
}

var resultsChan chan tweetResult

func storeOnIPFS(tweetThread []*twitter.Tweet) (string, error) {
	b, err := json.Marshal(tweetThread)
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

func getTweetThread(ID int64, thread []*twitter.Tweet) ([]*twitter.Tweet, error) {
	trimUser := true
	includeRetweets := false
	includeEntities := false
	params := &twitter.StatusShowParams{
		TrimUser:         &trimUser,
		IncludeMyRetweet: &includeRetweets,
		IncludeEntities:  &includeEntities,
	}
	tweet, _, err := client.Statuses.Show(ID, params)
	if err != nil {
		return nil, err
	}
	thread = append(thread, tweet)
	if tweet.InReplyToStatusID == 0 {
		return thread, nil
	}
	return getTweetThread(tweet.InReplyToStatusID, thread)
}

func getDcrtimeHost() string {
	return "https://" + piutil.NormalizeAddress(config.DcrTimeHost, config.DcrTimePort)
}

// handleTweet will receive a Tweet as the single input, fetch all the tweet
// thread (e.g the parent tweets in the thread, if any). Then it will make
// the necessary transformations to timestamp it in the Decred network and
// store on IPFS. The result details are sent in a different channel which
// will handle the results procedure.
func handleTweet(tweet *twitter.Tweet) {
	// @todo: validate tweets with regex patterns
	tweetThread := []*twitter.Tweet{}
	var displayThread []displayTweet

	// get all the parent tweets in the thread recusively
	tweetThread, err := getTweetThread(tweet.ID, tweetThread)

	if err != nil {
		log.Println("Cannot process twitter thread!:", err)
	} else {
		log.Println("\n Thread size: ", len(tweetThread))
	}

	// create the digest from the tweet thread
	b, err := json.Marshal(tweetThread)
	digest := piutil.Digest(b)
	var digests []*[sha256.Size]byte
	var d [sha256.Size]byte
	copy(d[:], digest[:sha256.Size])
	digests = append(digests, &d)

	// timestamp the digests using dcrtime
	err = piutil.Timestamp("test", getDcrtimeHost(), digests)
	if err != nil {
		log.Println("Could not timestamp", err)
	} else {
		log.Println("Timestamp OK")
	}

	// store the thread using IPFS
	cid, err := storeOnIPFS(tweetThread)
	if err != nil {
		log.Println("ipfs failed: ", err)
	} else {
		log.Println("ipfs OK", cid)
	}
	isReply := len(tweetThread) > 0
	if isReply {
		displayThread := make([]displayTweet, len(tweetThread))
		for i, reply := range tweetThread {
			displayThread[i] = displayTweet{
				ID:       reply.ID,
				Text:     reply.Text,
				UserName: reply.User.ScreenName,
			}
		}
	}
	// combine the results and send it through the results channel
	tr := tweetResult{
		Cid:    cid,
		Digest: hex.EncodeToString(digest[:]),
		Tweet: displayTweet{
			ID:       tweet.ID,
			Text:     tweet.Text,
			UserName: tweet.User.ScreenName,
		},
		isReply: isReply,
		Thread:  displayThread,
	}
	cacheTweetResult(count, tr)
	count++
	resultsChan <- tr

	log.Println("\n \n ======", count, " TWEETS ======= \n ")
}

func cacheTweetResult(count int, tr tweetResult) {
	if count < numberOfTweetsVisible {
		timestampedTweets[count] = tr
	} else {
		timestampedTweets = timestampedTweets[1:]
		timestampedTweets = append(timestampedTweets, tr)
	}
}

func handleTweetResult(tweetRes tweetResult) {
	if !config.EnableReplies {
		return
	}
	// reply to tweet thread with the timestmap and ipfs results
	opt := &twitter.StatusUpdateParams{
		InReplyToStatusID: tweetRes.Tweet.ID,
	}
	t, _, err := client.Statuses.Update("Thread stored! Cid: "+tweetRes.Cid+" and digest: "+tweetRes.Digest, opt)
	if err != nil {
		fmt.Println("Could not reply to tweet: ", err)
	} else {
		fmt.Println("Tweet successful sent, ID: ", t.ID)
	}
}

func listenToTweetResults() {
	for tr := range resultsChan {
		handleTweetResult(tr)
	}
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

	resultsChan = make(chan tweetResult)

	listenToTweets(stream, handleTweet)

	go listenToTweetResults()

	log.Println(len(config.TargetWords), " tracked words: ", config.TargetWords)
	log.Println("Start of the day!")

	startServer()

	// listen to stop signal and stop the stream before exiting
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)

	stream.Stop()
}
