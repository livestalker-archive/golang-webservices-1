package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

type TestCase struct {
	Request SearchRequest
	ErrInfo string
}

const (
	ACCESS_TOKEN = "1234567890"
)

func SearchServer(w http.ResponseWriter, r *http.Request) {
	// Get access token
	accessToken := r.Header.Get("AccessToken")
	if accessToken != ACCESS_TOKEN {
		w.WriteHeader(http.StatusUnauthorized)
	}
	// Get query prameters
	limit := r.URL.Query().Get("limit")
	offset := r.URL.Query().Get("offset")
	query := r.URL.Query().Get("query")
	order_field := r.URL.Query().Get("order_field")
	order_by := r.URL.Query().Get("order_by")
	log.Printf("limit: %s, offset: %s, query: %s, order_field: %s, order_by: %s", limit, offset, query, order_field, order_by)
	if v, _ := strconv.Atoi(limit); v >= 999999 {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// Check imput guards
func TestFindUserInput(t *testing.T) {
	cases := []TestCase{
		TestCase{
			Request: SearchRequest{
				Limit: -1,
			},
		},
		TestCase{
			Request: SearchRequest{
				Offset: -1,
			},
		},
	}
	searchClient := &SearchClient{}
	for _, item := range cases {
		_, err := searchClient.FindUsers(item.Request)
		if err == nil {
			t.Errorf("Input parameters not checked: %#v", item.Request)
		}
	}
}

func TestFindUserAuth(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()
	req := SearchRequest{}
	searchClient := &SearchClient{
		AccessToken: "wrong access token",
		URL:         ts.URL,
	}
	_, err := searchClient.FindUsers(req)
	if err == nil {
		t.Error("AccessToken not checked")
	}
}

func TestFindUserInternalError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()
	req := SearchRequest{Limit: 999999}
	searchClient := &SearchClient{
		AccessToken: ACCESS_TOKEN,
		URL:         ts.URL,
	}
	_, err := searchClient.FindUsers(req)
	if err == nil {
		t.Error("Not generate InternalError")
	}
}

//func TestFindUser(t *testing.T) {
//	cases := []TestCase{
//		TestCase{
//			Request: SearchRequest{},
//		},
//	}
//	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
//	defer ts.Close()
//	searchClient := &SearchClient{
//		AccessToken: "1234567890",
//		URL:         ts.URL,
//	}
//	for _, item := range cases {
//		_, err := searchClient.FindUsers(item.Request)
//		if err != nil {
//			t.Errorf("Error: %s", err)
//		}
//	}
//}
