package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type TestCase struct {
	Request SearchRequest
	ErrInfo string
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hello"))
}

// Check imput guards
func TestFindUserInput(t *testing.T) {
	cases := []TestCase{
		TestCase{
			Request: SearchRequest{
				Limit: -1,
			},
			ErrInfo: "Limit not checked",
		},
		TestCase{
			Request: SearchRequest{
				Offset: -1,
			},
			ErrInfo: "Offset not checked",
		},
	}
	searchClient := &SearchClient{}
	for _, item := range cases {
		_, err := searchClient.FindUsers(item.Request)
		if err == nil {
			t.Error(item.ErrInfo)
		}
	}
}

func TestFindUser(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()
	searchClient := &SearchClient{
		AccessToken: "1234567890",
		URL:         ts.URL,
	}
}
