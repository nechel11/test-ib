package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

type Request struct {
	Method  string
	Queue   string
	Message string
	Timeout string
}

type Response struct {
	Status   int
	Response string
}

type TestCase struct {
	Req  Request
	Resp Response
}

func (r Request) URL() string {
	// Message and Timeout can't both exist
	url := "/" + r.Queue
	if r.Message != "" {
		url += "?v=" + r.Message
	}
	if r.Timeout != "" {
		url += "?timeout=" + r.Timeout
	}
	return url
}

func TestBasic(t *testing.T) {
	cases := []TestCase{
		{
			Req: Request{
				Method:  "PUT",
				Queue:   "color",
				Message: "red",
			},
			Resp: Response{
				Status: http.StatusOK,
			},
		},
		{
			Req: Request{
				Method:  "PUT",
				Queue:   "color",
				Message: "green",
			},
			Resp: Response{
				Status: http.StatusOK,
			},
		},

		{
			Req: Request{
				Method:  "PUT",
				Queue:   "name",
				Message: "alex",
			},
			Resp: Response{
				Status: http.StatusOK,
			},
		},
		{
			Req: Request{
				Method:  "PUT",
				Queue:   "name",
				Message: "anna",
			},
			Resp: Response{
				Status: http.StatusOK,
			},
		},

		{
			Req: Request{
				Method: "GET",
				Queue:  "color",
			},
			Resp: Response{
				Status:   http.StatusOK,
				Response: "red",
			},
		},
		{
			Req: Request{
				Method: "GET",
				Queue:  "color",
			},
			Resp: Response{
				Status:   http.StatusOK,
				Response: "green",
			},
		},
		{
			Req: Request{
				Method: "GET",
				Queue:  "color",
			},
			Resp: Response{
				Status: http.StatusNotFound,
			},
		},
		{
			Req: Request{
				Method: "GET",
				Queue:  "color",
			},
			Resp: Response{
				Status: http.StatusNotFound,
			},
		},

		{
			Req: Request{
				Method: "GET",
				Queue:  "name",
			},
			Resp: Response{
				Status:   http.StatusOK,
				Response: "alex",
			},
		},
		{
			Req: Request{
				Method: "GET",
				Queue:  "name",
			},
			Resp: Response{
				Status:   http.StatusOK,
				Response: "anna",
			},
		},
		{
			Req: Request{
				Method: "GET",
				Queue:  "name",
			},
			Resp: Response{
				Status: http.StatusNotFound,
			},
		},
	}

	handler := getHandler()

	for _, c := range cases {
		url := c.Req.URL()
		req := httptest.NewRequest(c.Req.Method, url, nil)
		w := httptest.NewRecorder()
		handler(w, req)

		if w.Code != c.Resp.Status {
			t.Errorf("expected status: %d, got: %d", c.Resp.Status, w.Code)
		}

		resp := w.Result()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}

		bodyStr := string(body)
		if bodyStr != c.Resp.Response {
			t.Errorf("expected body: %s, got: %s", c.Resp.Response, body)
		}
	}
}

func TestBadRequest(t *testing.T) {
	cases := []TestCase{
		{
			Req: Request{
				Method: "PUT",
			},
			Resp: Response{
				Status: http.StatusBadRequest,
			},
		},
		{
			Req: Request{
				Method:  "PUT",
				Queue:   "queue",
				Message: "message",
			},
			Resp: Response{
				Status: http.StatusOK,
			},
		},
		{
			Req: Request{
				Method:  "GET",
				Timeout: "str",
			},
			Resp: Response{
				Status: http.StatusBadRequest,
			},
		},
		{
			Req: Request{
				Method: "POST",
			},
			Resp: Response{
				Status: http.StatusBadRequest,
			},
		},
	}

	handler := getHandler()

	for _, c := range cases {
		url := c.Req.URL()
		req := httptest.NewRequest(c.Req.Method, url, nil)
		w := httptest.NewRecorder()
		handler(w, req)

		if w.Code != c.Resp.Status {
			t.Errorf("expected status: %d, got: %d", c.Resp.Status, w.Code)
		}

		resp := w.Result()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}

		bodyStr := string(body)
		if bodyStr != c.Resp.Response {
			t.Errorf("expected body: %s, got: %s", c.Resp.Response, body)
		}
	}
}

func TestTimeout(t *testing.T) {
	cases := []TestCase{
		{
			Req: Request{
				Method:  "GET",
				Timeout: "1",
			},
			Resp: Response{
				Status: http.StatusNotFound,
			},
		},
	}

	handler := getHandler()

	for _, c := range cases {
		url := c.Req.URL()
		req := httptest.NewRequest(c.Req.Method, url, nil)
		w := httptest.NewRecorder()
		handler(w, req)

		if w.Code != c.Resp.Status {
			t.Errorf("expected status: %d, got: %d", c.Resp.Status, w.Code)
		}

		resp := w.Result()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}

		bodyStr := string(body)
		if bodyStr != c.Resp.Response {
			t.Errorf("expected body: %s, got: %s", c.Resp.Response, body)
		}
	}
}

func runTestCase(c TestCase, handler http.HandlerFunc, t *testing.T) {
	url := c.Req.URL()
	req := httptest.NewRequest(c.Req.Method, url, nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != c.Resp.Status {
		t.Errorf("expected status: %d, got: %d", c.Resp.Status, w.Code)
	}

	resp := w.Result()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}

	bodyStr := string(body)
	if bodyStr != c.Resp.Response {
		t.Errorf("expected body: %s, got: %s", c.Resp.Response, body)
	}
}

func TestOrder(t *testing.T) {
	const N = 20

	handler := getHandler()

	gets := make([]TestCase, N)
	for i := range gets {
		gets[i] = TestCase{
			Req: Request{
				Method:  "GET",
				Queue:   "queue",
				Timeout: "3",
			},
			Resp: Response{
				Status:   http.StatusOK,
				Response: strconv.Itoa(i),
			},
		}

		go func(ind int) {
			runTestCase(gets[ind], handler, t)
		}(i)
		time.Sleep(50 * time.Millisecond)
	}

	puts := make([]TestCase, N)
	for i := range puts {
		puts[i] = TestCase{
			Req: Request{
				Method:  "PUT",
				Queue:   "queue",
				Message: strconv.Itoa(i),
			},
			Resp: Response{
				Status: http.StatusOK,
			},
		}
		runTestCase(puts[i], handler, t)
	}
}

func TestConcurrent(t *testing.T) {
	const N = 1000

	handler := getHandler()

	for i := 0; i < N; i++ {
		go func() {
			put := TestCase{
				Req: Request{
					Method:  "PUT",
					Queue:   "queue",
					Message: "message",
				},
				Resp: Response{
					Status: http.StatusOK,
				},
			}
			runTestCase(put, handler, t)

			time.Sleep(50 * time.Millisecond)

			get := TestCase{
				Req: Request{
					Method: "GET",
					Queue:  "queue",
				},
				Resp: Response{
					Status: http.StatusOK,
				},
			}

			url := get.Req.URL()
			req := httptest.NewRequest(get.Req.Method, url, nil)
			w := httptest.NewRecorder()
			handler(w, req)

			if w.Code != get.Resp.Status {
				t.Errorf("expected status: %d, got: %d", get.Resp.Status, w.Code)
			}
		}()
	}
}
