package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

var (
	port  = "5678"
	token string
)

type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type Action struct {
	Name  string `json:"name"`
	Text  string `json:"text"`
	Type  string `json:"type"`
	Value string `json:"value"`
	Style string `json:"style"`
}

type Attachment struct {
	Fields         []Field  `json:"fields"`
	Actions        []Action `json:"actions"`
	Fallback       string   `json:"fallback"`
	CallbackID     string   `json:"callback_id"`
	Color          string   `json:"color"`
	AttachmentType string   `json:"attachment_type"`
	Title          string   `json:"title"`
}

type Response struct {
	Attachments []Attachment `json:"attachments"`
	Channel     string       `json:"channel"`
	TS          string       `json:"ts"`
}

type Team struct {
	ID     string `json:"id"`
	Domain string `json:"domain"`
}

type Channel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type User struct {
	Channel
}

type Request struct {
	Payload
}

type UserAction []string

type Payload struct {
	Actions         []Action `json:"actions"`
	Team            Team     `json:"team"`
	Channel         Channel  `json:"channel"`
	User            User     `json:"user"`
	OriginalMessage Response `json:"original_message"`
	Type            string   `json:"type"`
	CallbackID      string   `json:"callback_id"`
	ActionTS        string   `json:"action_ts"`
	MessageTS       string   `json:"message_ts"`
	AttachmentID    string   `json:"attachment_id"`
	Token           string   `json:"token"`
	ResponseURL     string   `json:"response_url"`
	TriggerID       string   `json:"trigger_id"`
	IsAppUnfurl     bool     `json:"is_app_unfurl"`
}

func replyToSlack(r Response, w http.ResponseWriter, endPointName string) {
	reqJSON, _ := json.Marshal(r)

	// fmt.Println(string(reqJSON), "Sending to Slack")

	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://slack.com/api/"+endPointName, bytes.NewBuffer([]byte(reqJSON)))

	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err, "Error")
		http.Error(w, "error", http.StatusBadRequest)
		return
	}

	// body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	// fmt.Printf(string(body))
}

func leaveRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Can use this one with this method", http.StatusBadGateway)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	channel := r.FormValue("channel_id")
	text := r.FormValue("text")
	fmt.Println(text)

	var name, date, reason, leaveType string

	if strings.Contains(text, "wf") {
		contents := strings.Split(text, "\n")
		submitter := contents[0]
		leaveRequestType := contents[1]
		duration := contents[2]
		leaveRequestReason := contents[4]
		name = strings.Split(submitter, ": ")[1]
		leaveType = strings.Split(leaveRequestType, ":")[1]
		date = strings.Split(duration, ": ")[1]
		reason = strings.Split(leaveRequestReason, ": ")[2]
	} else {
		contents := strings.Split(text, " ")
		name = "<" + contents[0] + ">"
		date = contents[1]
		reason = contents[2]
		leaveType = contents[3]
	}

	fieds := []Field{
		{"Submitter", name, true},
		{"Date", date, true},
		{"Reason", reason, true},
		{"Leave type", leaveType, true},
	}

	actions := []Action{
		{
			Name:  "validation",
			Text:  "Approve",
			Type:  "button",
			Value: "ok",
			Style: "primary",
		},
		{
			Name:  "validation",
			Text:  "No",
			Type:  "button",
			Value: "ko",
			Style: "danger",
		},
	}

	response := Response{
		Attachments: []Attachment{
			{
				Fallback:       "You are unable to choose a validation type",
				CallbackID:     "lqid",
				Color:          "#3AA3E3",
				AttachmentType: "default",
				Fields:         fieds,
				Actions:        actions,
			},
		},
		Channel: channel,
	}

	replyToSlack(response, w, "chat.postMessage")
}

func messageActionHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("[ERROR] Failed to read request body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonStr, err := url.QueryUnescape(string(buf)[8:])
	if err != nil {
		log.Printf("[ERROR] Failed to unescape request body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var message Request
	if err := json.Unmarshal([]byte(jsonStr), &message); err != nil {
		log.Printf("[ERROR] Failed to decode json message from slack: %s", jsonStr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userActions := UserAction{}

	users := append(userActions, message.Payload.User.ID)
	log.Println(users, "users")
	userLength := len(users)
	attLength := len(message.Payload.OriginalMessage.Attachments)

	var usersSayYes []string
	newAttachments := message.Payload.OriginalMessage.Attachments
	var listUsersSay []string

	if userLength > 0 {
		for _, u := range users {
			listUsersSay = append(usersSayYes, userToTagString(u))
		}
		newAttachments[0].Actions[0].Text = strings.Join([]string{"Approved ", "(", strconv.Itoa(userLength), ")"}, "")

		var hasOrHave = "has"
		if userLength > 1 {
			hasOrHave = "have"
		}

		if attLength > 1 {
			textOfElement := strings.Split(newAttachments[1].Title, " ")
			usersSubmitted := textOfElement[:len(textOfElement) - 2]
	
			log.Println(usersSubmitted, "usersSubmitted")


			newAttachments[1] = Attachment{
				Title: strings.Join(listUsersSay, "") + hasOrHave + " approved",
				Color: "good",
			}
		} else {
			newAttachments = append(message.Payload.OriginalMessage.Attachments, Attachment{
				Title: strings.Join(listUsersSay, "") + hasOrHave + " approved",
				Color: "good",
			})
		}
	} else {
		newAttachments[0].Actions[0].Text = "Approved"
	}

	newMessage := Response{
		Channel:     message.Payload.Channel.ID,
		Attachments: newAttachments,
		TS:          message.Payload.OriginalMessage.TS,
	}

	reqJSON, _ := json.Marshal(newMessage)
	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://slack.com/api/chat.update", bytes.NewBuffer([]byte(reqJSON)))

	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err, "Error")
		http.Error(w, "error", http.StatusBadRequest)
		return
	}

	ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	// fmt.Printf(string(body))
}

func userToTagString(u string) string {
	return "<@" + u + "> "
}

func getAllUsersVoted(s string) []string {
	return strings.Split(s, "")
}

func handleRequest() {
	http.HandleFunc("/lq", leaveRequestHandler)
	http.HandleFunc("/ma", messageActionHandler)
	log.Printf("Server is starting at %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func initial() {
	if "" == os.Getenv("TOKEN") {
		panic("Token is not found")
	}

	if "" != os.Getenv("PORT") {
		port = os.Getenv("PORT")
	}

	token = os.Getenv("TOKEN")
}

func main() {
	initial()
	handleRequest()
}
