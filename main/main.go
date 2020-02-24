package main

import (
	"fmt"

	/*"github.com/jerminb/nestor"
	log "github.com/sirupsen/logrus"
	gock "gopkg.in/h2non/gock.v1"*/

	"time"

	"github.com/jasonlvhit/gocron"
)

func task() {
	fmt.Println("Task running")
}

func bgcron() {
	<-gocron.Start()
}

func main() {
	/*log.SetLevel(log.DebugLevel)
	defer gock.Off()
	gock.New("http://server.com").
		Get("/bar").
		Reply(200)
	responseChan := make(chan *nestor.PollResponse)
	p, err := nestor.NewPollee("http://server.com/bar", "GET", 2, "200 OK", responseChan)
	log.Info("Pollee setup")
	if err != nil {
		log.Errorf("expected nil. got %v", err)
	}
	log.Info("Running pollee")
	go p.Poll()
	r := <-responseChan
	if r.Error != nil {
		log.Errorf("expected nil. got %v", r.Error)
	}*/
	gocron.Every(1).Second().Do(task)
	go bgcron()
	fmt.Println("first run after start")
	time.Sleep(2 * time.Second)
	fmt.Println("second run after start")

}
