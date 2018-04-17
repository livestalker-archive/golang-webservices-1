package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type DataSetUsers struct {
	XMLName xml.Name      `xml:"root"`
	Users   []DataSetUser `xml:"row"`
}

func (ds *DataSetUsers) Filter(s string) {
	newUsers := make([]DataSetUser, 0)
	for _, item := range ds.Users {
		if strings.Contains(item.FName, s) || strings.Contains(item.LName, s) {
			newUsers = append(newUsers, item)
		}
	}
	ds.Users = newUsers
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

// Realize sort functionality
type By func(u1, u2 *DataSetUser) bool

func (by By) Sort(ds DataSetUsers, order int) {
	us := &userSorter{
		ds:    ds,
		by:    by,
		order: order,
	}
	sort.Sort(us)
}

type userSorter struct {
	ds    DataSetUsers
	by    By
	order int
}

func (us *userSorter) Len() int {
	return len(us.ds.Users)
}

func (us *userSorter) Swap(i, j int) {
	us.ds.Users[i], us.ds.Users[j] = us.ds.Users[j], us.ds.Users[i]
}

func (us *userSorter) Less(i, j int) bool {
	res := us.by(&us.ds.Users[i], &us.ds.Users[j])
	if us.order == -1 {
		res = !res
	}
	return res
}

const (
	ACCESS_TOKEN     = "1234567890"
	DATASET_FILENAME = "dataset.xml"
)

// Main search server
func SearchServer(w http.ResponseWriter, r *http.Request) {
	// Get access token
	accessToken := r.Header.Get("AccessToken")
	if accessToken != ACCESS_TOKEN {
		w.WriteHeader(http.StatusUnauthorized)
	}
	// Get query prameters
	//limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	//offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	query := r.URL.Query().Get("query")
	order_field := r.URL.Query().Get("order_field")
	order_by, _ := strconv.Atoi(r.URL.Query().Get("order_by"))
	if order_by != 0 && order_by != 1 && order_by != -1 {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if order_field == "about" {
		w.WriteHeader(http.StatusBadRequest)
		b, _ := json.Marshal(SearchErrorResponse{Error: "ErrorBadOrderField"})
		w.Write(b)
		return
	}
	if order_field == "wrong" {
		w.WriteHeader(http.StatusBadRequest)
		b, _ := json.Marshal(SearchErrorResponse{Error: "NewError"})
		w.Write(b)
		return
	}
	dsUsers := getDataSet(DATASET_FILENAME)
	if dsUsers == nil {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `[]`)
	}
	if order_by != 0 {
		sortDataSet(*dsUsers, order_field, order_by)
	}
	if query != "" {
		dsUsers.Filter(query)
	}
	// TODO limit+offset
	b, err := json.Marshal(dsUsers.Users)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write(b)
}

// Search server which return wrong json
func SearchServerWrongJson(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Wrong Json"))
}

// Search server which return status: bad request
func SearchServerBadRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("Wrong json"))
}

// Read dataset from xml source
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

// Sort dataset by field and specific order
func sortDataSet(ds DataSetUsers, field string, order int) {
	switch field {
	case "id":
		By(byId).Sort(ds, order)
	case "name":
		By(byName).Sort(ds, order)
	case "age":
		By(byAge).Sort(ds, order)
	case "gender":
		By(byGender).Sort(ds, order)
	default:
		By(byId).Sort(ds, order)
	}
}

// Sort by user id
func byId(u1, u2 *DataSetUser) bool {
	return u1.Id < u2.Id
}

// Sort by full user name
func byName(u1, u2 *DataSetUser) bool {
	return u1.FName+u1.LName < u2.FName+u2.LName
}

// Sort by user age
func byAge(u1, u2 *DataSetUser) bool {
	return u1.Age < u2.Age
}

// Sort by gender
func byGender(u1, u2 *DataSetUser) bool {
	return u1.Gender < u2.Gender
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
	req := SearchRequest{OrderBy: 2}
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
			Request: SearchRequest{
				Limit: 999,
			},
		},
		TestCase{
			Request: SearchRequest{
				OrderBy: -1,
				Query:   "Kane",
			},
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

func TestFindUserWrongJson(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerWrongJson))
	defer ts.Close()
	searchClient := &SearchClient{
		AccessToken: "1234567890",
		URL:         ts.URL,
	}
	_, err := searchClient.FindUsers(SearchRequest{})
	if err == nil {
		t.Error("Pass wrong Json")
	}
}

func TestFindUserTimeout(t *testing.T) {
	searchClient := &SearchClient{
		AccessToken: "1234567890",
		URL:         "",
	}
	_, err := searchClient.FindUsers(SearchRequest{})
	if err == nil {
		t.Error("Pass without server")
	}
}

func TestFindUserBadRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerBadRequest))
	defer ts.Close()
	searchClient := &SearchClient{
		AccessToken: "1234567890",
		URL:         ts.URL,
	}
	_, err := searchClient.FindUsers(SearchRequest{})
	if err == nil {
		t.Error("Pass bad request and wrong json")
	}
}

func TestFindUserBadOrderField(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()
	searchClient := &SearchClient{
		AccessToken: "1234567890",
		URL:         ts.URL,
	}
	_, err := searchClient.FindUsers(SearchRequest{OrderField: "about"})
	if err == nil {
		t.Error("Pass bad order field")
	}
}

func TestFindUserNewError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()
	searchClient := &SearchClient{
		AccessToken: "1234567890",
		URL:         ts.URL,
	}
	_, err := searchClient.FindUsers(SearchRequest{OrderField: "wrong"})
	if err == nil {
		t.Error("Pass bad order field")
	}
}
