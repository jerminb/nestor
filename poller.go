package nestor

import (
	"errors"
	"net"
	"net/http"
	"reflect"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	//DefaultConnectionTimeout default connection timeout
	DefaultConnectionTimeout time.Duration = time.Second * 5
	//DefaultDialTimeout default dial timeout
	DefaultDialTimeout time.Duration = time.Second * 5
	//DefaultTLSHandshakeTimeout default tls handshake timeout
	DefaultTLSHandshakeTimeout time.Duration = 5 * time.Second
)

//ErrorMaxCountExceeded is the error returned when error count for a pollee exceeds its max error count
var ErrorMaxCountExceeded = errors.New("exceeded max error count")

//PollResponse is the message that is returned by polled through pollResponseChannel
type PollResponse struct {
	ResponseStatus string
	Error          error
}

//Pollee defines an url to monitor for a specific response with an error count monitor.
//Connection timeouts are counted as error.
type Pollee struct {
	url                 string
	ErrorCount          int
	MaxErrorCount       int
	SuccessStatus       string
	request             *http.Request
	client              *http.Client
	pollResponseChannel chan<- *PollResponse
}

func (p *Pollee) sendPollResponse(responseStatus string, err error) {
	p.pollResponseChannel <- &PollResponse{
		ResponseStatus: responseStatus,
		Error:          err,
	}
}

// Poll executes an httpVerb request for url
// and returns the HTTP status string or an error asyncronously through
// pollResponseChannel
func (p *Pollee) Poll() {
	log.Infof("ErrorCount %d, MaxErrorCount %d", p.ErrorCount, p.MaxErrorCount)
	if p.ErrorCount >= p.MaxErrorCount {
		p.sendPollResponse("", ErrorMaxCountExceeded)
		return
	}
	p.ErrorCount++
	log.Debugf("Polling %s ....", p.url)
	resp, err := p.client.Do(p.request)
	if err != nil {
		log.Debugf("Failed polling %s with %v", p.url, err)
		p.sendPollResponse("", err)
	}
	respStatus := ""
	if resp != nil {
		defer resp.Body.Close()
		respStatus = resp.Status
		log.Debugf("Response statue %s", respStatus)
	}
	if respStatus == p.SuccessStatus {
		log.Debugf("Expected %s. Got %s", p.SuccessStatus, respStatus)
		p.ErrorCount = 0
	}
	p.sendPollResponse(respStatus, nil)
}

//NewPollee is a constructor for Pollee class
func NewPollee(url string, httpMethod string, maxErrorCount int, successStatus string, responseChannel chan<- *PollResponse) (*Pollee, error) {
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: DefaultDialTimeout,
		}).Dial,
		TLSHandshakeTimeout: DefaultTLSHandshakeTimeout,
	}
	var netClient = &http.Client{
		Timeout:   DefaultConnectionTimeout,
		Transport: netTransport,
	}
	req, err := http.NewRequest(httpMethod, url, nil)
	if err != nil {
		return nil, err
	}

	return &Pollee{
		url:                 url,
		ErrorCount:          0,
		MaxErrorCount:       maxErrorCount,
		SuccessStatus:       successStatus,
		client:              netClient,
		request:             req,
		pollResponseChannel: responseChannel,
	}, nil
}

//Poller monitors a Pollee object on fixed intervals and returns success if
// if success status is achieved or failure if max error count of pollee is expired
type Poller struct {
}

//Monitor is the blocking implementation of polling logic.
// When the function returns either a successful status result is received (returned as true)
// or max error count is reached (returned as false)
// error is for internal error handling
func (p *Poller) Monitor(url string, httpMethod string, maxErrorCount int, successStatus string, pollInterval time.Duration) (bool, error) {
	ticker := time.NewTicker(pollInterval)
	pollResponseChannel := make(chan *PollResponse)
	pe, err := NewPollee(url, httpMethod, maxErrorCount, successStatus, pollResponseChannel)
	if err != nil {
		return false, err
	}
	for {
		select {
		case <-ticker.C:
			go pe.Poll()
		case r := <-pollResponseChannel:
			if r.Error != nil {
				return false, r.Error
			}
			if r.ResponseStatus == pe.SuccessStatus {
				return true, nil
			}
		}
	}
}

//Execute executes Poller's Monitor to implement Executable interface
func (p *Poller) Execute(params ...interface{}) (result []reflect.Value, err error) {
	return execute(p.Monitor, params...)
}

//NewPoller is constructor for Poller class
func NewPoller() *Poller {
	return &Poller{}
}
