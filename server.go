package main

import (
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"net/http"

	"github.com/thedevsaddam/renderer"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:3000", "http service address")

var upgrader = websocket.Upgrader{} // use default options
var templates *template.Template

var homeTemplate = template.Must(template.ParseFiles("./public/home.html"))

var rnd *renderer.Render

type wsConfig struct {
	WsHost string
}

func echo(w http.ResponseWriter, r *http.Request) {
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

	// for {
	// 	mt, message, err := c.ReadMessage()
	// 	if err != nil {
	// 		log.Println("read:", err)
	// 		break
	// 	}
	// 	log.Printf("recv: %s", message)
	// 	err = c.WriteMessage(mt, message)
	// 	if err != nil {
	// 		log.Println("write:", err)
	// 		break
	// 	}
	// }
}

func home(w http.ResponseWriter, r *http.Request) {
	wscfg := wsConfig{
		WsHost: "ws://" + r.Host + "/echo",
	}
	rnd.HTML(w, http.StatusOK, "home", wscfg)
	// homeTemplate.Execute(w, "ws://"+r.Host+"/echo")
}

func startServer() {

	opts := renderer.Options{
		ParseGlobPattern: "./public/*.html",
	}

	rnd = renderer.New(opts)

	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./public/css"))))
	http.HandleFunc("/echo", echo)
	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

// func main() {
// 	flag.Parse()
// 	log.SetFlags(0)
// 	http.HandleFunc("/echo", echo)
// 	http.HandleFunc("/", home)
// 	log.Fatal(http.ListenAndServe(*addr, nil))
// }
