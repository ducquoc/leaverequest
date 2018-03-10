package main

import (
	"strings"
	"io/ioutil"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"encoding/json"
)

var (
	port = "5678"
	token string
)

type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool `json:"short"`
}

type Action struct {
	Name string `json:"name"`
	Text string `json:"text"`
	Type string `json:"type"`
	Value string `json:"value"`
	Style string `json:"style"`
}

type Attachment struct {
	Fallback string `json:"fallback"`
	CallbackID string `json:"callback_id"`
	Color string `json:"color"`
	AttachmentType string `json:"attachment_type"`
	Fields []Field `json:"fields"`
	Actions []Action `json:"actions"`
}

type Response struct {
	Attachments []Attachment `json:"attachments"`
	Channel string `json:"channel"`
}

type Request struct {
	Token string
	TeamID string
	TeamDomain string
	ChannelID string
	ChannelName string
	UserID string
	UserName string
	Command string
	Text string
	ResponseURL string
	TriggerID string
}

func ReplySlack(r Response, w http.ResponseWriter) {
	reqJSON, _ := json.Marshal(r)
	
	fmt.Println(string(reqJSON), "Sending to Slack")


	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", bytes.NewBuffer([]byte(reqJSON)))

	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer " + token)
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err, "Error")
		http.Error(w, "error", http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Printf(string(body))
}

func LeaveRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Can use this one with this method", http.StatusBadGateway)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	channel := r.FormValue("channel_id")
	text := r.FormValue("text")
	fmt.Println(text);

	contents := strings.Split(text, " ")
	name, date, reason := contents[0], contents[1], contents[2]
	fmt.Println(contents);

	fieds := []Field{
		{"Submitter", "<" + name + ">", true},
		{"Date", date, true},
		{"Reason", reason, true},
	}

	actions := []Action{
		{
			Name: "validation",
			Text: "OK for me",
			Type: "button",
			Value: "ok",
			Style: "primary",
		},
		{
			Name: "validation",
			Text: "KO for me",
			Type: "button",
			Value: "ko",
			Style: "danger",
		},
	}

	response := Response{
		Attachments: []Attachment{
			{
				Fallback: "You are unable to choose a validation type",
				CallbackID: "wopr_game",
				Color: "#3AA3E3",
				AttachmentType: "default",
				Fields: fieds,
				Actions: actions,
			},
		},
		Channel: channel,
	}

	ReplySlack(response, w)
}

func HandleRequest() {
	http.HandleFunc("/lq", LeaveRequestHandler)
	log.Printf("Server is starting at %s", port)
	log.Fatal(http.ListenAndServe(":" + port, nil))
}

func Init() {
	if "" == os.Getenv("TOKEN") {
		panic("Token is not found")
	}

	if "" != os.Getenv("PORT") {
		port = os.Getenv("PORT")
	}

	token = os.Getenv("TOKEN")
}

func main() {
	Init()
	HandleRequest()
}
