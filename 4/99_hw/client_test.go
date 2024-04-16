package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

type TestCase struct {
	Request   *SearchRequest
	Result    *SearchResponse
	IsError   bool
	ErrorText string
}

// код писать тут
func SearchServerDummy(w http.ResponseWriter, r *http.Request) {
	key := r.FormValue("query")
	switch key {
	case "An":
		basicPersonsDataset = PersonsDataset{}
		basicPersonsDataset.Init()
		handler(w, r)
	case "100500":
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{"status": 400, "err": "bad_balance"}`)
	case "__broken_json":
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{"status": 400`) //broken json
	case "__timeout":
		time.Sleep(2 * time.Second)
	case "__bad_access_token":
		w.WriteHeader(http.StatusUnauthorized)
	case "__bad_request_broken_json":
		w.WriteHeader(http.StatusBadRequest)
	case "__bad_request_bad_field":
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"Status": 400, "Error": "ErrorBadOrderField"}`)
	case "__bad_request_unknown":
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"Status": 400, "Error": "unknown"}`)
	case "__unknown_error":
		w.WriteHeader(http.StatusTeapot)
	case "__internal_error":
		fallthrough
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func TestSearchRequestFindUsers(t *testing.T) {
	cases := []TestCase{
		{
			Request:   &SearchRequest{Limit: -1},
			Result:    nil,
			IsError:   true,
			ErrorText: "limit must be > 0",
		},
		{
			Request:   &SearchRequest{Offset: -1},
			Result:    nil,
			IsError:   true,
			ErrorText: "offset must be > 0",
		},
		{
			Request:   &SearchRequest{Query: "__timeout"},
			Result:    nil,
			IsError:   true,
			ErrorText: "timeout for limit=1&offset=0&order_by=0&order_field=&query=__timeout",
		},
		{
			Request:   &SearchRequest{Query: "__bad_access_token"},
			Result:    nil,
			IsError:   true,
			ErrorText: "Bad AccessToken",
		},
		{
			Request:   &SearchRequest{Query: "__bad_request_broken_json"},
			Result:    nil,
			IsError:   true,
			ErrorText: "cant unpack error json: unexpected end of JSON input",
		},
		{
			Request:   &SearchRequest{Query: "__bad_request_bad_field", OrderField: "bad_field"},
			Result:    nil,
			IsError:   true,
			ErrorText: "OrderFeld bad_field invalid",
		},
		{
			Request:   &SearchRequest{Query: "__bad_request_unknown"},
			Result:    nil,
			IsError:   true,
			ErrorText: "unknown bad request error: unknown",
		},
		{
			Request:   &SearchRequest{Query: "__internal_error"},
			Result:    nil,
			IsError:   true,
			ErrorText: "SearchServer fatal error",
		},
		{
			Request:   &SearchRequest{Query: "__broken_json"},
			Result:    nil,
			IsError:   true,
			ErrorText: "cant unpack result json: unexpected end of JSON input",
		},
		//
		{
			Request: &SearchRequest{Query: "An", Limit: 1, Offset: 0, OrderField: "Id"},
			Result: &SearchResponse{Users: []User{
				{Id: 16, Name: "Annie", Age: 35, About: "Consequat fugiat veniam commodo nisi nostrud culpa pariatur. Aliquip velit adipisicing dolor et nostrud. Eu nostrud officia velit eiusmod ullamco duis eiusmod ad non do quis.\n", Gender: "female"},
			},
				NextPage: true},
			IsError: false,
		},
		//
		{
			Request: &SearchRequest{Query: "An", Limit: 1, Offset: 100, OrderField: "Id"},
			Result: &SearchResponse{Users: []User{},
				NextPage: false},
			IsError: false,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServerDummy))

	for caseNum, item := range cases {
		c := &SearchClient{
			AccessToken: "",
			URL:         ts.URL,
		}

		result, err := c.FindUsers(*item.Request)

		if err != nil && !item.IsError {
			t.Errorf("[%d] unexpected error: %#v", caseNum, err)
		}
		if err == nil && item.IsError {
			t.Errorf("[%d] expected error, got nil", caseNum)
		}
		if err != nil && err.Error() != item.ErrorText {
			t.Errorf("[%d] expected error '%s', got '%s'", caseNum, item.ErrorText, err.Error())
		}
		if !reflect.DeepEqual(item.Result, result) {
			t.Errorf("[%d] wrong result, expected %#v, got %#v", caseNum, item.Result, result)
		}
	}
	ts.Close()
}
