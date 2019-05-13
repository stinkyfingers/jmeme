package main

import (
	"net/http"
	"strings"
	"testing"
)

func TestParseForm(t *testing.T) {
	body := "team_id=123&team_domain=test&channel_id=456"
	req, err := http.NewRequest("POST", "/", strings.NewReader(string(body)))
	if err != nil {
		t.Error(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	s, err := parseForm(req)
	if err != nil {
		t.Error(err)
	}
	if s.TeamID != "123" {
		t.Errorf("expected team ID 123, got %s", s.TeamID)
	}
}
