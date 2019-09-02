package nestor_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jerminb/nestor"
	gock "gopkg.in/h2non/gock.v1"
)

func TestPollee(t *testing.T) {
	defer gock.Off()
	gock.New("http://foo.com").
		Get("/bar").
		Reply(200).
		JSON(map[string]string{"foo": "bar"})
	responseChan := make(chan *nestor.PollResponse)
	p, err := nestor.NewPollee("http://foo.com/", "GET", 1, "200 OK", responseChan)
	if err != nil {
		t.Fatalf("expected nil. got %v", err)
	}
	go p.Poll()
	r := <-responseChan
	if r.Error != nil {
		t.Fatalf("expected nil. got %v", r.Error)
	}
	if r.ResponseStatus != "200 OK" {
		t.Fatalf("expected status 200 OK got %s", r.ResponseStatus)
	}
}

func TestFailedPollee(t *testing.T) {
	defer gock.Off()
	gock.New("http://foo.com").
		Get("/bar").
		Reply(200).
		JSON(map[string]string{"foo": "bar"})
	responseChan := make(chan *nestor.PollResponse)
	p, err := nestor.NewPollee("http://bar.br/", "GET", 1, "200 OK", responseChan)
	if err != nil {
		t.Fatalf("expected nil. got %v", err)
	}
	go p.Poll()
	r := <-responseChan
	if r.Error == nil {
		t.Fatalf("expected error. got nil")
	}
}

func TestClientTimeout(t *testing.T) {
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d := map[string]interface{}{
			"bar": "bar",
			"foo": "foo",
		}

		time.Sleep(6000 * time.Millisecond)
		b, err := json.Marshal(d)
		if err != nil {
			t.Error(err)
		}
		io.WriteString(w, string(b))
		w.WriteHeader(http.StatusOK)
	})

	//backend := httptest.NewServer(http.TimeoutHandler(handlerFunc, 20*time.Millisecond, "server timeout"))
	backend := httptest.NewServer(http.HandlerFunc(handlerFunc))
	url := backend.URL
	responseChan := make(chan *nestor.PollResponse)
	p, err := nestor.NewPollee(url, "GET", 1, "200 OK", responseChan)
	if err != nil {
		t.Fatalf("expected nil. got %v", err)
	}
	go p.Poll()
	r := <-responseChan
	if r.Error == nil {
		t.Fatalf("expected error. got nil")
	}
}
func TestLessThanMaxErrorCount(t *testing.T) {
	counter := 0
	maxErrorCount := 2
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d := map[string]interface{}{
			"bar": "bar",
			"foo": "foo",
		}

		if counter < (maxErrorCount - 1) {
			counter++
			w.WriteHeader(http.StatusNotFound)
			return
		}

		b, err := json.Marshal(d)
		if err != nil {
			t.Error(err)
		}
		io.WriteString(w, string(b))
		w.WriteHeader(http.StatusOK)
	})

	//backend := httptest.NewServer(http.TimeoutHandler(handlerFunc, 20*time.Millisecond, "server timeout"))
	backend := httptest.NewServer(http.HandlerFunc(handlerFunc))
	url := backend.URL
	responseChan := make(chan *nestor.PollResponse)
	p, err := nestor.NewPollee(url, "GET", maxErrorCount, "200 OK", responseChan)
	if err != nil {
		t.Fatalf("expected nil. got %v", err)
	}
	go p.Poll()
	r := <-responseChan
	go p.Poll()
	r = <-responseChan
	if r.Error != nil {
		t.Fatalf("expected nil. got %v", err)
	}
	if r.ResponseStatus != "200 OK" {
		t.Fatalf("expected status 200 OK got %s", r.ResponseStatus)
	}
}

func TestMoreThanMaxErrorCount(t *testing.T) {
	counter := 0
	maxErrorCount := 1
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d := map[string]interface{}{
			"bar": "bar",
			"foo": "foo",
		}

		if counter < (maxErrorCount) {
			counter++
			w.WriteHeader(http.StatusNotFound)
			return
		}

		b, err := json.Marshal(d)
		if err != nil {
			t.Error(err)
		}
		io.WriteString(w, string(b))
		w.WriteHeader(http.StatusOK)
	})

	//backend := httptest.NewServer(http.TimeoutHandler(handlerFunc, 20*time.Millisecond, "server timeout"))
	backend := httptest.NewServer(http.HandlerFunc(handlerFunc))
	url := backend.URL
	responseChan := make(chan *nestor.PollResponse)
	p, err := nestor.NewPollee(url, "GET", maxErrorCount, "200 OK", responseChan)
	if err != nil {
		t.Fatalf("expected nil. got %v", err)
	}
	go p.Poll()
	r := <-responseChan
	go p.Poll()
	r = <-responseChan
	if r.Error == nil {
		t.Fatalf("expected error. got nil")
	}
	if r.Error != nestor.ErrorMaxCountExceeded {
		t.Fatalf("expected ErrorMaxCountExceeded. got %v", r.Error)
	}
}

func TestPoller(t *testing.T) {
	counter := 0
	maxErrorCount := 3
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d := map[string]interface{}{
			"bar": "bar",
			"foo": "foo",
		}

		if counter < (maxErrorCount - 1) {
			counter++
			w.WriteHeader(http.StatusNotFound)
			return
		}

		b, err := json.Marshal(d)
		if err != nil {
			t.Error(err)
		}
		io.WriteString(w, string(b))
		w.WriteHeader(http.StatusOK)
	})

	//backend := httptest.NewServer(http.TimeoutHandler(handlerFunc, 20*time.Millisecond, "server timeout"))
	backend := httptest.NewServer(http.HandlerFunc(handlerFunc))
	url := backend.URL
	poller := nestor.NewPoller()
	start := time.Now()
	res, err := poller.Monitor(url, "GET", maxErrorCount, "200 OK", time.Second*1)
	end := time.Now()
	elapsed := end.Sub(start)
	if err != nil {
		t.Fatalf("expected no error. got %v", err)
	}
	if !res {
		t.Fatalf("expected success=true. got false")
	}
	if elapsed < time.Second*2 {
		t.Fatalf("expected 10 seconds. got %v", elapsed)
	}
}

func TestPollerExecutablePositive(t *testing.T) {
	counter := 0
	maxErrorCount := 3
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d := map[string]interface{}{
			"bar": "bar",
			"foo": "foo",
		}

		if counter < (maxErrorCount - 1) {
			counter++
			w.WriteHeader(http.StatusNotFound)
			return
		}

		b, err := json.Marshal(d)
		if err != nil {
			t.Error(err)
		}
		io.WriteString(w, string(b))
		w.WriteHeader(http.StatusOK)
	})
	backend := httptest.NewServer(http.HandlerFunc(handlerFunc))
	url := backend.URL
	poller := nestor.NewPoller()
	res, err := poller.Execute(url, "GET", maxErrorCount, "200 OK", time.Second*1)
	if err != nil {
		t.Fatalf("expected no error. got %v", err)
	}
	if res == nil {
		t.Fatalf("expected result. got nil")
	}
}

func TestPollerExecutableNegative(t *testing.T) {
	counter := 0
	maxErrorCount := 3
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d := map[string]interface{}{
			"bar": "bar",
			"foo": "foo",
		}

		if counter < (maxErrorCount - 1) {
			counter++
			w.WriteHeader(http.StatusNotFound)
			return
		}

		b, err := json.Marshal(d)
		if err != nil {
			t.Error(err)
		}
		io.WriteString(w, string(b))
		w.WriteHeader(http.StatusOK)
	})
	backend := httptest.NewServer(http.HandlerFunc(handlerFunc))
	url := backend.URL
	poller := nestor.NewPoller()
	_, err := poller.Execute(url, "GET", maxErrorCount, "200 OK")
	if err == nil {
		t.Fatalf("expected error. got nil")
	}
}
