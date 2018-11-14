package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

var (
	base      = "https://conversations.messagebird.com/v1/"
	channelID = "YOUR CHANNEL ID HERE"

	httpListenAddress = flag.String("http-listen-address", ":5007", "The address to listen for http")
	publicAddress     = flag.String("public-address", "https://ngrok.io/TOTO", "The public address of this bot on the interwebs")
	accessKey         = flag.String("access-key", "invalid", "MessageBird API access key to talk to conversations api.")

	baseURL *url.URL

	manager *todoManager
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()

	baseURL, _ = url.Parse(base)

	manager = newTodoManager()

	http.HandleFunc("/create-hook", createHookHandler)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "ok")
	})

	log.Println("listening for http on ", *httpListenAddress)
	log.Fatal(http.ListenAndServe(*httpListenAddress, nil))
}
