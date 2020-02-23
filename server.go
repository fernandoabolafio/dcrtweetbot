package main

import (
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/thedevsaddam/renderer"
)

var addr = flag.String("addr", "localhost:3000", "http service address")

var upgrader = websocket.Upgrader{} // use default options
var templates *template.Template

var homeTemplate = template.Must(template.ParseFiles("./public/home.html"))

var rnd *renderer.Render

type wsConfig struct {
	WsHost        string
	StampedTweets []displayTweet
}

func streamTweets(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	for tr := range resultsChan {
		b, err := json.Marshal(tr)
		if err != nil {
			log.Println(err)
		}
		err = c.WriteMessage(1, b)
		if err != nil {
			log.Println(err)
		}
	}

}

func home(w http.ResponseWriter, r *http.Request) {
	wscfg := wsConfig{
		WsHost:        "ws://" + r.Host + "/stream",
		StampedTweets: timestampedTweets,
	}
	rnd.HTML(w, http.StatusOK, "home", wscfg)
}

func startServer() {
	opts := renderer.Options{
		ParseGlobPattern: "./public/*.html",
	}

	rnd = renderer.New(opts)

	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./public/css"))))
	http.HandleFunc("/stream", streamTweets)
	http.HandleFunc("/", home)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Server is running on port 3000")
}
