package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"
)

//custom seach CP - https://cse.google.com

var (
	verificationToken = os.Getenv("VERIFICATION_TOKEN") // used to verify request came from Slack
	authToken         = os.Getenv("AUTH_TOKEN")         // used for slack API calls
	slackHookURL      = os.Getenv("SLACK_HOOK_URL")     // url for (optional) Slack hook
	googleAPIKey      = os.Getenv("GOOGLE_API_KEY")     // API key for google custom search
	port              = ":" + os.Getenv("PORT")
)

// Result represents the GoogleAPIs image search result
type Result struct {
	Items []Item `json:"items"`
}

// Item represents a GoogleAPIs image item
type Item struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	DisplayLink string `json:"displayLink"`
}

// SlackRequest resprents the form data received from a slack slash command integration
type SlackRequest struct {
	Token       string `json:"token,omitempty"`
	TeamID      string `json:"team_id,omitempty"`
	TeamDomain  string `json:"team_domain,omitempty"`
	ChannelID   string `json:"channel_id,omitempty"`
	ChannelName string `json:"channel_name,omitempty"`
	UserID      string `json:"user_id,omitempty"`
	UserName    string `json:"user_name,omitempty"`
	Command     string `json:"command,omitempty"`
	Text        string `json:"text,omitempty"`
	ResponseURL string `json:"response_url,omitempty"`
	TriggerID   string `json:"trigger_id,omitempty"`
}

// Data represents outbound Slack post data
type Data struct {
	Channel     string       `json:"channel"`
	Scope       string       `json:"scope"`
	Attachments []Attachment `json:"attachments"`
}

// Attachment represents the needed fields for a Data attachment
type Attachment struct {
	Pretext  string `json:"pretext"`
	ImageURL string `json:"image_url"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/test", handler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		log.Print("health called")
		_, err := w.Write([]byte("OK"))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	})

	if port == ":" {
		port = ":9090"
	}
	log.Fatal(http.ListenAndServe(port, mux))
}

//Use for testing Slack Messaging
func handler(w http.ResponseWriter, r *http.Request) {
	s, err := parseForm(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if s.Token != verificationToken {
		http.Error(w, "No/Incorrect Token", http.StatusUnauthorized)
		return
	}

	link, err := googleMeme(s.Text)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = slackSendMessage(link, s.ChannelID, s.Text)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func parseForm(r *http.Request) (*SlackRequest, error) {
	err := r.ParseForm()
	if err != nil {
		return nil, err
	}
	return &SlackRequest{
		Token:       r.Form.Get("token"),
		TeamID:      r.Form.Get("team_id"),
		TeamDomain:  r.Form.Get("team_domain"),
		ChannelID:   r.Form.Get("channel_id"),
		ChannelName: r.Form.Get("channel_name"),
		Command:     r.Form.Get("command"),
		ResponseURL: r.Form.Get("response_url"),
		Text:        r.Form.Get("text"),
		TriggerID:   r.Form.Get("trigger_id"),
		UserID:      r.Form.Get("user_id"),
		UserName:    r.Form.Get("user_name"),
	}, nil
}

func googleMeme(text string) (string, error) {
	//googleapis query
	query := url.PathEscape(text + " meme")
	num := 10
	cx := "010251510427321670814:7o209j8g99y"
	url := fmt.Sprintf("https://www.googleapis.com/customsearch/v1?key=%s&cx=%s&num=%d&q=%s&searchType=image&source=lnms", googleAPIKey, cx, num, query)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	//decode googleapis result
	var result Result
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}
	ran := rand.New(rand.NewSource(time.Now().UnixNano()))
	selected := result.Items[ran.Intn(len(result.Items))]
	return selected.Link, nil
}

// send message to Slack via webhook
func slackPostMessage(link, channelID, text string) error {
	d := Data{
		Channel: channelID,
		Attachments: []Attachment{{
			Pretext:  text,
			ImageURL: link,
		}},
	}

	j, err := json.Marshal(d)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", slackHookURL, bytes.NewReader(j))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	cli := &http.Client{}
	_, err = cli.Do(req)
	return err
}

// send message to Slack via chat.sendMessage API endpoint
func slackSendMessage(link, channelID, text string) error {
	slackPostURL := "https://slack.com/api/chat.postMessage"
	d := Data{
		Channel: channelID,
		Scope:   "chat:write:user",
		Attachments: []Attachment{{
			Pretext:  text,
			ImageURL: link,
		}},
	}

	j, err := json.Marshal(d)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", slackPostURL, bytes.NewReader(j))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authToken))
	cli := &http.Client{}
	_, err = cli.Do(req)
	return err
}
