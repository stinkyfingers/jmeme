package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

//custom seach CP - https://cse.google.com
//https://jmeme.herokuapp.com/
//token eI77YEBucKLqum63p21ADlfH

const (
	API         = "https://curtmfg.slack.com/services/hooks/slackbot?token=JVJ1Y9etcJyECkltRBWDZYpW&channel=%23"
	CHANNEL     = "testgroup"
	WEBHOOK_URL = "https://hooks.slack.com/services/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"
	TOKEN       = "eI77YEBucKLqum63p21ADlfH"
)

type Result struct {
	Items []Item `json:"items"`
}

type Item struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	DisplayLink string `json:"displayLink"`
}

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
}

type SlackResponse struct {
	Text      string `json:"text,omitempty"`
	Channel   string `json:"channel,omitempty"`
	UserName  string `json:"username,omitempty"`
	IconUrl   string `json:"icon_url,omitempty"`
	IconEmoji string `json:"icon_emoji,omitempty"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", googleHandler)
	mux.HandleFunc("/slack", slackTest)

	port := ":" + os.Getenv("PORT")
	if port == ":" {
		port = ":9090"
	}
	err := http.ListenAndServe(port, mux)
	if err != nil {
		log.Print(err)
	}
}

func googleHandler(w http.ResponseWriter, r *http.Request) {
	//decode req body
	var s SlackRequest
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.Token = r.FormValue("token")
	s.TeamID = r.FormValue("team_id")
	s.TeamDomain = r.FormValue("team_domain")
	s.ChannelID = r.FormValue("channel_id")
	s.ChannelName = r.FormValue("channel_name")
	s.UserID = r.FormValue("user_id")
	s.UserName = r.FormValue("user_name")
	s.Command = r.FormValue("command")
	s.Text = r.FormValue("text")

	if s.Token != TOKEN {
		http.Error(w, "No/Incorrect Token", http.StatusUnauthorized)
		return
	}
	q := s.Text + " meme"

	//googleapis query
	q = strings.Replace(q, " ", "+", -1)
	num := "10"
	key := "AIzaSyCyO3v3xEKKu4SV44S-czADtjSwzp39oXM"
	cx := "010251510427321670814:7o209j8g99y"
	url := "https://www.googleapis.com/customsearch/v1?key=" + key + "&cx=" + cx + "&num=" + num + "&q=" + q + "&searchType=image&source=lnms"

	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//decode googleapis result
	var result Result
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ran := rand.New(rand.NewSource(time.Now().UnixNano()))
	selected := result.Items[ran.Intn(len(result.Items))]

	//post to slack
	err = PostToSlack(selected.Link, s.ChannelName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	return
}

func PostToSlack(body, channel string) error {
	cli := &http.Client{}

	payload := SlackResponse{
		Text:     body,
		UserName: "jmeme",
		Channel:  "#" + channel,
	}
	reader, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	payloadStr := string(reader[:])

	req, err := http.NewRequest("POST", WEBHOOK_URL, strings.NewReader(payloadStr))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	_, err = cli.Do(req)

	return nil
}

func slackTest(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	text := r.FormValue("text")
	err = PostToSlack(text, "testgroup")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}
