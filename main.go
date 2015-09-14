package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
)

//custom seach CP - https://cse.google.com
//https://jmeme.herokuapp.com/
//token eI77YEBucKLqum63p21ADlfH

const (
	API     = "https://curtmfg.slack.com/services/hooks/slackbot?token=JVJ1Y9etcJyECkltRBWDZYpW&channel=%23"
	CHANNEL = "testgroup"
	TOKEN   = "eI77YEBucKLqum63p21ADlfH"
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
	Token       string `json:"token"`
	TeamID      string `json:"team_id"`
	TeamDomain  string `json:"team_domain"`
	ChannelID   string `json:"channel_id"`
	ChannelName string `json:"channel_name"`
	UserID      string `json:"user_id"`
	UserName    string `json:"user_name"`
	Command     string `json:"command"`
	Text        string `json:"text"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", googleHandler)

	port := ":" + os.Getenv("PORT")
	if port == ":" {
		port = ":8080"
	}
	err := http.ListenAndServe(port, mux)
	if err != nil {
		panic(err)
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
	log.Print(s)
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

	//post to slack
	err = PostToSlack(result.Items[1].Link)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//unecessary writing of json - for debugging
	js, err := json.Marshal(result)
	w.Write([]byte(js))

}

func PostToSlack(body string) error {
	cli := &http.Client{}
	reader := strings.NewReader(body)
	req, err := http.NewRequest("POST", API+CHANNEL, reader)
	if err != nil {
		return err
	}
	_, err = cli.Do(req)
	if err != nil {
		return err
	}
	return nil
}
