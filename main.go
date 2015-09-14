package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

//custom seach CP - https://cse.google.com

const (
	API     = "https://curtmfg.slack.com/services/hooks/slackbot?token=JVJ1Y9etcJyECkltRBWDZYpW&channel=%23"
	CHANNEL = "testgroup"
)

type Result struct {
	Items []Item `json:"items"`
}

type Item struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	DisplayLink string `json:"displayLink"`
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
	q := r.URL.Query().Get("query") + " meme"
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

	var result Result
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = PostToSlack(result.Items[1].Link)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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
