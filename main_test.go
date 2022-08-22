package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_main(t *testing.T) {
	testCases := []struct {
		name           string
		httpMethod     string
		key				string
		value				string
		Timeout 			string
		WantHTTPStatus int
		WantMessage		string
	}{
		{
			name:          "invalid method",
			httpMethod:     "ABC",
			key:				"color",
			value:			"blue",
			Timeout			string
			WantHTTPStatus int
			WantMessage		string
		},
		{
			name:             "valid method POST empty JSON",
			httpMethod:       http.MethodPost,
			reqBody:          strings.NewReader("{}"),
			WantHTTPStatus:   http.StatusOK,
			WantJsonResponse: `{"url":""}`+"\n",
		},
		{
			name:             "valid method POST with valid JSON",
			httpMethod:       http.MethodPost,
			reqBody:          strings.NewReader(`{"url":"ozon.ru"}`),
			WantHTTPStatus:   http.StatusOK,
			WantJsonResponse: `{"url":` + `"` + utils.Hash_func("ozon.ru") + `"}` +"\n",
		},
		{
			name:             "valid method POST with invalid JSON",
			httpMethod:       http.MethodPost,
			reqBody:          strings.NewReader(`{"url":ozon.ru}`),
			WantHTTPStatus:   http.StatusBadRequest,
			WantJsonResponse: `400`+ ` json decoding error. request should be {"url" : "value"}` +"\n",
		},
		{
			name:             "valid method GET empty JSON",
			httpMethod:       http.MethodGet,
			reqBody:          strings.NewReader("{}"),
			WantHTTPStatus:   http.StatusOK,
			WantJsonResponse: `{"url":""}` + "\n",
		},
		{
			name:             "valid method GET invalid JSON",
			httpMethod:       http.MethodGet,
			reqBody:          strings.NewReader(`{"url":XXX}`),
			WantHTTPStatus:   http.StatusBadRequest,
			WantJsonResponse: `400`+ ` json decoding error. request should be {"url" : "value"}` +"\n",
		},


		

	}
	// Act
	handler := http.HandlerFunc(PG_handler)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tc.httpMethod, "/", tc.reqBody)
			handler.ServeHTTP(rec, req)
			assert.Equal(t, tc.WantJsonResponse, rec.Body.String())
			assert.Equal(t, tc.WantHTTPStatus, rec.Code)

		})

	}
}
