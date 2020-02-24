package nestor_test

import (
	"testing"
	"time"

	"github.com/jerminb/nestor"
	"github.com/jerminb/nestor/testserver"
)

func TestPollee(t *testing.T) {
	testserver.WithTestServer(t, func(url string) {
		responseChan := make(chan *nestor.PollResponse)
		p, err := nestor.NewPollee(url, "GET", 1, "200 OK", responseChan)
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
	})
}

func TestFailedPollee(t *testing.T) {
	testserver.WithTestServer(t, func(url string) {
		responseChan := make(chan *nestor.PollResponse)
		p, err := nestor.NewPollee(url, "GET", 1, "200 OK", responseChan)
		if err != nil {
			t.Fatalf("expected nil. got %v", err)
		}
		go p.Poll()
		r := <-responseChan
		if r.ResponseStatus == "200 OK" {
			t.Fatalf("expected error. got %s", r.ResponseStatus)
		}
	},
		testserver.MaxErrorCount(1))
}

func TestClientTimeout(t *testing.T) {
	testserver.WithTestServer(t, func(url string) {
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
	},
		testserver.MaxErrorCount(1),
		testserver.TimeToFirstByte(6000*time.Millisecond))

}
func TestLessThanMaxErrorCount(t *testing.T) {
	maxErrorCount := 2
	testserver.WithTestServer(t, func(url string) {
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
	},
		testserver.MaxErrorCount(maxErrorCount-1))
}

func TestMoreThanMaxErrorCount(t *testing.T) {
	maxErrorCount := 1
	testserver.WithTestServer(t, func(url string) {
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
	},
		testserver.MaxErrorCount(maxErrorCount))
}

func TestPoller(t *testing.T) {
	maxErrorCount := 3
	testserver.WithTestServer(t, func(url string) {
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
			t.Fatalf("expected 2 seconds. got %v", elapsed)
		}
	},
		testserver.MaxErrorCount(maxErrorCount-1))
}

func TestPollerExecutablePositive(t *testing.T) {
	maxErrorCount := 3
	testserver.WithTestServer(t, func(url string) {
		poller := nestor.NewPoller()
		res, err := poller.Execute(url, "GET", maxErrorCount, "200 OK", time.Second*1)
		if err != nil {
			t.Fatalf("expected no error. got %v", err)
		}
		if res == nil {
			t.Fatalf("expected result. got nil")
		}
	},
		testserver.MaxErrorCount(maxErrorCount-1))
}

func TestPollerExecutableNegative(t *testing.T) {
	maxErrorCount := 3
	testserver.WithTestServer(t, func(url string) {
		poller := nestor.NewPoller()
		_, err := poller.Execute(url, "GET", maxErrorCount, "200 OK")
		if err == nil {
			t.Fatalf("expected error. got nil")
		}
	},
		testserver.MaxErrorCount(maxErrorCount))
}
