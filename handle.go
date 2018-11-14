package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"strings"
)

/*
{
   "conversation":{
      "id":"55c66895c22a40e39a8e6bd321ec192e",
      "contactId":"db4dd5087fb343738e968a323f640576",
      "status":"active",
      "createdDatetime":"2018-08-17T10:14:14Z",
      "updatedDatetime":"2018-08-17T14:30:31.915292912Z",
      "lastReceivedDatetime":"2018-08-17T14:30:31.898389294Z"
   },
   "message":{
      "id":"ddb150149e2c4036a48f581544e22cfe",
      "conversationId":"55c66895c22a40e39a8e6bd321ec192e",
      "channelId":"23a780701b8849f7b974d8620a89a279",
      "status":"received",
      "type":"text",
      "direction":"received",
      "content":{
         "text":"Gqghhd"
      },
      "createdDatetime":"2018-08-17T14:30:31.898389294Z",
      "updatedDatetime":"2018-08-17T14:30:31.915292912Z"
   },
   "type":"message.created"
}
*/

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

	err = handleMessage(whp)
	if err != nil {
		log.Println("Err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal Server Error")

		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "ok")
}

func handleMessage(whp *whPayload) error {

	if whp.Message.Direction != "received" {
		// we sent this message. No need to respond.
		return nil
	}

	list := manager.fetch(whp.Conversation.ID)

	text := whp.Message.Content.Text
	text = regexp.MustCompile(" +").ReplaceAllString(text, " ")
	parts := strings.Split(text, " ")
	command := strings.ToLower(parts[0])

	responseBody := "I don't understand. Type 'help' to get help."

	switch command {
	case "help":
		responseBody = `Hi, I'm a TODO list bot.

Commands:

*add <item>*
adds an item to the todo list.

*list*
shows the list

*done <item number>*
marks item as done

*bye*
closes the list

	                   `
	case "add":
		if len(parts) < 2 {
			return respond(whp.Conversation.ID, "err... the 'add' command needs a second param: the todo item you want to save. Something like 'add buy milk'.")
		}
		item := strings.Join(parts[1:], " ")
		list.add(item)
		responseBody = "added."
	case "done":
		if len(parts) < 2 {
			return respond(whp.Conversation.ID, "err... the 'done' command needs the index of the item you want to mark as done. Something like 'done 2'.")
		}
		itemNo, err := strconv.Atoi(parts[1])
		if err != nil {
			return respond(whp.Conversation.ID, "err... that doesn't look like a number. ")
		}
		err = list.done(itemNo)
		if err == errItemNotFound {
			responseBody = "item not found."
		} else {
			// it should be fine
			responseBody = "marked as done."
		}
	case "list":
		if len(list.read()) > 0 {
			responseBody = ""
			for i, item := range list.read() {
				responseBody += fmt.Sprintf("%d: %s \n", i, item)
			}
		} else {
			responseBody = "nothing."
		}
	case "bye":
		defer archiveConversation(whp.Conversation.ID)
		manager.close(whp.Conversation.ID)
		responseBody = "bye!"
	}

	return respond(whp.Conversation.ID, responseBody)
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

	return nil
}

func archiveConversation(conversationID string) error {
	// PATCH https://conversations.messagebird.com/v1/conversations/{id}

	u := *baseURL
	u.Path = path.Join(baseURL.Path, "conversations", conversationID)

	payload := struct {
		Status string `json:"status"`
	}{
		Status: "archive",
	}

	var b bytes.Buffer
	err := json.NewEncoder(&b).Encode(&payload)
	if err != nil {
		return fmt.Errorf("Error encoding buffer: %v", err)
	}

	req, err := http.NewRequest("PATCH", u.String(), &b)
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
		return fmt.Errorf("Bad response code from api when trying to archive conversation: %s. Body: %s", resp.Status, string(body))
	}

	return nil
}
