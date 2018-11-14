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
	base      = "https://conversations.messagebird.com/v1/"
	channelID = "dd619fdc24ee404f82d8986d0fe1e05e"

	httpListenAddress = flag.String("http-listen-address", ":5007", "The address to listen for http")
	publicAddress     = flag.String("public-address", "https://ngrok.io/TOTO", "The public address of this bot on the interwebs")
	accessKey         = flag.String("access-key", "invalid", "MessageBird API access key to talk to conversations api.")

	baseURL *url.URL
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()

	baseURL, _ = url.Parse(base)

	http.HandleFunc("/create-hook", createHookHandler)

	log.Println("listening for http on ", *httpListenAddress)
	log.Fatal(http.ListenAndServe(*httpListenAddress, nil))
}

type whPayload struct {
	Conversation conversation `json:"conversation"`
	Message      message      `json:"message"`
	Type         string       `json:"type"`
}

type message struct {
	ID        string  `json:"id"`
	Direction string  `json:"direction"`
	Type      string  `json:"type"`
	Content   content `json:"content"`
}

type content struct {
	Text string `json:"text"`
}

type conversation struct {
	ID string `json:"id"`
}

func createHookHandler(w http.ResponseWriter, r *http.Request) {
	whp := &whPayload{}
	err := json.NewDecoder(r.Body).Decode(whp)
	if err != nil {
		log.Println("Err: got weird body on the webhook")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal Server Error")
		return
	}

	if whp.Message.Direction != "received" {
		// you will get *all* messages on the webhook. Even the ones this bot sends to the channel. We don't want to answer those.
		fmt.Fprintf(w, "ok")
		return
	}

	// just echo
	err = respond(whp.Conversation.ID, whp.Message.Content.Text)
	if err != nil {
		log.Println("Err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal Server Error")

		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "ok")
}

func respond(conversationID, responseBody string) error {
	// POST https://conversations.messagebird.com/v1/conversations/{id}/messages

	u := *baseURL
	u.Path = path.Join(baseURL.Path, "conversations", conversationID, "messages")

	msg := message{
		Content: content{
			Text: responseBody,
		},
		Type: "text",
	}

	var b bytes.Buffer
	err := json.NewEncoder(&b).Encode(&msg)
	if err != nil {
		return fmt.Errorf("Error encoding buffer: %v", err)
	}

	req, err := http.NewRequest("POST", u.String(), &b)
	req.Header.Set("Authorization", "AccessKey "+*accessKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Bad response code from api when trying to create message: %s. Body: %s", resp.Status, string(body))
	}

	log.Println("All good. Response body: ", string(body))
	return nil
}
