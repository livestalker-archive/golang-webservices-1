package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

type DataSetUsers struct {
	XMLName xml.Name      `xml:"root"`
	Users   []DataSetUser `xml:"row"`
}

type DataSetUser struct {
	Id     int    `xml:"id"`
	FName  string `xml:"first_name"`
	LName  string `xml:"last_name"`
	Age    int    `xml:"age"`
	About  string `xml:"about"`
	Gender string `xml:"gender"`
}

type TestCase struct {
	Request SearchRequest
	ErrInfo string
}

type By func(u1, u2 *DataSetUser) bool

const (
	ACCESS_TOKEN     = "1234567890"
	DATASET_FILENAME = "dataset.xml"
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
	dsUsers := getDataSet(DATASET_FILENAME)
	if dsUsers == nil {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `[]`)
	}
}

func getDataSet(fileName string) *DataSetUsers {
	xmlFile, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil
	}
	defer xmlFile.Close()
	var dsUsers DataSetUsers
	byteValue, _ := ioutil.ReadAll(xmlFile)
	xml.Unmarshal(byteValue, &dsUsers)
	return &dsUsers
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

func TestFindUser(t *testing.T) {
	cases := []TestCase{
		TestCase{
			Request: SearchRequest{},
		},
	}
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()
	searchClient := &SearchClient{
		AccessToken: "1234567890",
		URL:         ts.URL,
	}
	for _, item := range cases {
		_, err := searchClient.FindUsers(item.Request)
		if err != nil {
			t.Errorf("Error: %s", err)
		}
	}
}
