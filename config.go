package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/decred/dcrd/dcrutil"
	v1 "github.com/decred/dcrtime/api/v1"
	flags "github.com/jessevdk/go-flags"
)

const (
	defaultConfigFilename = "dcrtweetbot.conf"
)

var (
	defaultHomeDir = dcrutil.AppDataDir("dcrtweetbot", false)

	defaultConfigFile    = filepath.Join(defaultHomeDir, defaultConfigFilename)
	defaultTargetWords   = []string{}
	defaultIPFSHost      = "localhost:5001"
	defaultDcrtimeHost   = v1.DefaultMainnetTimeHost
	defaultDcrtimePort   = v1.DefaultMainnetTimePort
	defaultEnableReplies = false

	usageMessage string
)

// Config describe the config options available
type Config struct {
	TwitAPIConsumerKey    string   `long:"twitterconsumerkey" description:"Consumer key for Twitter API (required)"`
	TwitAPIConsumerSecret string   `long:"twitterconsumersecret" description:"Consumer secret for Twitter API (required)"`
	TwitAPIToken          string   `long:"twitterapitoken" description:"Token for Twitter API (required)"`
	TwitAPITokenSecret    string   `long:"twitterapitokensecret" description:"Token secret for Twitter API (required)"`
	TargetWords           []string `short:"t" long:"targetwords" description:"The target words to track"`
	ConfigFile            string   `short:"c" long:"config" description:"The configuration file to be used"`
	EnableReplies         bool     `short:"r" long:"enablereplies" description:"Send replies via Twitter API"`
	IPFSHost              string   `long:"ipfshost" description:"The IPFS API host"`
	DcrTimeHost           string   `long:"dcrtimehost" description:"The dcrtime API host"`
	DcrTimePort           string   `long:"dcrtimeport" description:"The dcrtime API port"`
}

func parseCommandLineOptions(cfg *Config) {
	parser := flags.NewParser(cfg, flags.HelpFlag)
	_, err := parser.Parse()
	if err != nil {
		fmt.Println(err)
		if e, ok := err.(*flags.Error); ok && e.Type == flags.ErrHelp {
			// help flag was up
			os.Exit(0)
		}
		fmt.Println("Parse config error: ", err)
		os.Exit(1)
	}
}

func parseConfigFileOptions(cfg *Config) {
	parser := flags.NewParser(cfg, flags.HelpFlag)
	err := flags.NewIniParser(parser).ParseFile(cfg.ConfigFile)
	if err != nil {
		if _, ok := err.(*os.PathError); !ok {
			fmt.Fprintf(os.Stderr, "Error parsing config "+
				"file: %v\n", err)
			fmt.Fprintln(os.Stderr, usageMessage)
			os.Exit(1)
		}
	}

}

func setUsageMessage() {
	appName := filepath.Base(os.Args[0])
	appName = strings.TrimSuffix(appName, filepath.Ext(appName))
	usageMessage = fmt.Sprintf("Use %s -h or --help to show usage", appName)
}

// loadConfig initializes and parses the config using a config file and command
// line options. Command line options takes precedence.
func loadConfig() *Config {

	setUsageMessage()

	config := &Config{
		ConfigFile:    defaultConfigFile,
		TargetWords:   defaultTargetWords,
		IPFSHost:      defaultIPFSHost,
		DcrTimeHost:   defaultDcrtimeHost,
		DcrTimePort:   defaultDcrtimePort,
		EnableReplies: defaultEnableReplies,
	}

	// parse command line options to check for a different config file location
	parseCommandLineOptions(config)

	// clear target words because to avoid duplication
	config.TargetWords = []string{}

	// parse the config file
	parseConfigFileOptions(config)

	// parse the command line options again so it can take precedence over
	// the config file options
	parseCommandLineOptions(config)

	// check that all twitter API options were provided
	if config.TwitAPIConsumerKey == "" || config.TwitAPIConsumerSecret == "" || config.TwitAPIToken == "" || config.TwitAPITokenSecret == "" {
		fmt.Fprintln(os.Stderr, "All twitter API config options are required.", usageMessage)
		os.Exit(1)
	}

	return config
}
