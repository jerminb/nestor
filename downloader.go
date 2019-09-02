package nestor

import (
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/cavaliercoder/grab"
	log "github.com/sirupsen/logrus"
)

// Hook is a user provided callback function. If a hook returns an error, the
// associated request is canceled and the same error is returned on the Response
// object.
// Hook functions are called synchronously and should never block unnecessarily.
type Hook func(string) error

//Downloader is implemented to manage file download of different sized.
//The goal is to make sure that connectivity, resume and authentication are all
// encapsulated in a single implementation
type Downloader struct {
	client          *grab.Client
	UpdateTicker    int
	BatchWorkerSize int
}

//Download uses a threaded download approach to improve speed and exception handling.
func (d *Downloader) Download(filepath string, url string) error {
	req, err := grab.NewRequest(filepath, url)
	if err != nil {
		return err
	}
	log.Debugf("Downloading %v ...", req.URL())
	resp := d.client.Do(req)
	log.Debugf("%v", resp.HTTPResponse.Status)
	t := time.NewTicker(time.Duration(d.UpdateTicker) * time.Millisecond)
	defer t.Stop()

Loop:
	for {
		select {
		case <-t.C:
			log.Debugf("  transferred %v / %v bytes (%.2f%%)",
				resp.BytesComplete(),
				resp.Size(),
				100*resp.Progress())

		case <-resp.Done:
			// download is complete
			break Loop
		}
	}

	// check for errors
	if err := resp.Err(); err != nil {
		return err
	}

	log.Debugf("Download saved to ./%v", resp.Filename)

	return nil
}

//DownloadBatch sends multiple HTTP requests and downloads the content of the
// requested URLs to the given destination directory
func (d *Downloader) DownloadBatch(filepath string, hook Hook, urls ...string) error {
	fi, err := os.Stat(filepath)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("destination is not a directory")
	}
	reqs := make([]*grab.Request, len(urls))
	for i := 0; i < len(urls); i++ {
		req, err := grab.NewRequest(filepath, urls[i])
		if err != nil {
			return err
		}
		if hook != nil {
			req.AfterCopy = getGrabHookFromHook(hook)
		}
		reqs[i] = req
	}
	responses := d.client.DoBatch(d.BatchWorkerSize, reqs...)

Loop:
	for i := 0; i < len(reqs); {
		select {
		case resp := <-responses:
			log.Debugf("Checking response for filename /%s", resp.Filename)
			if resp == nil {
				log.Debugf("Response is nil breaking the download loop")
				break Loop
			}
			err = d.isComplete(resp)
			if err != nil {
				return fmt.Errorf("%s: %v", resp.Filename, err)
			}
			if err := resp.Err(); err != nil {
				return fmt.Errorf("%s: %v", resp.Filename, err)
			}
			i++
		}
	}
	return nil
}

func (d *Downloader) isComplete(resp *grab.Response) error {
	<-resp.Done
	if !resp.IsComplete() {
		return fmt.Errorf("Response.IsComplete returned false")
	}

	if resp.Start.IsZero() {
		return fmt.Errorf("Response.Start is zero")
	}

	if resp.End.IsZero() {
		return fmt.Errorf("Response.End is zero")
	}

	if eta := resp.ETA(); eta != resp.End {
		return fmt.Errorf("Response.ETA is not equal to Response.End: %v", eta)
	}

	// the following fields should only be set if no error occurred
	if resp.Err() == nil {
		if resp.Filename == "" {
			return fmt.Errorf("Response.Filename is empty")
		}

		if resp.Size() == 0 {
			return fmt.Errorf("Response.Size is zero")
		}
	}
	return nil
}

func getGrabHookFromHook(hook Hook) grab.Hook {
	if hook != nil {
		return func(resp *grab.Response) error {
			return hook(resp.Filename)
		}
	}
	return nil
}

//Execute executes Downloader's Download to implement Executable interface
func (d *Downloader) Execute(params ...interface{}) (result []reflect.Value, err error) {
	return execute(d.Download, params...)
}

//NewDownloader is the constructor for Downloader struct
func NewDownloader() *Downloader {
	return &Downloader{
		client:          grab.NewClient(),
		UpdateTicker:    500,
		BatchWorkerSize: 5,
	}
}
