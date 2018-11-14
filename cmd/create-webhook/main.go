package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
)

var (
	channelID = "dd619fdc24ee404f82d8986d0fe1e05e"

	publicAddress = flag.String("public-address", "https://ngrok.io/TOTO", "The public address of this bot on the interwebs")
	accessKey     = flag.String("access-key", "invalid", "MessageBird API access key to talk to conversations api.")
)

type webhook struct {
	Events    []string `json:"events"`
	ChannelID string   `json:"channelId"`
	URL       string   `json:"url"`
}

func main() {
	flag.Parse()

	whurl, _ := url.Parse(*publicAddress)
	whurl.Path = path.Join(whurl.Path, "/create-hook")

	wh := webhook{
		Events:    []string{"message.created"},
		ChannelID: channelID,
		URL:       whurl.String(),
	}

	var b bytes.Buffer
	err := json.NewEncoder(&b).Encode(&wh)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", "https://conversations.messagebird.com/v1/webhooks", &b)
	req.Header.Set("Authorization", "AccessKey "+*accessKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= http.StatusBadRequest {
		panic(fmt.Errorf("Bad response code from api when trying to create webhook: %s. Body: %s", resp.Status, string(body)))
	}

	log.Println("response body: ", string(body))
}
