package nestor_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/cavaliercoder/grab/grabtest"
	"github.com/jerminb/nestor"
)

func fileCount(path string) (int, error) {
	i := 0
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return 0, err
	}
	for _, file := range files {
		if !file.IsDir() {
			i++
		}
	}
	return i, nil
}

func TestDownload(t *testing.T) {
	d := nestor.NewDownloader()
	now := time.Now()
	nanos := now.UnixNano()
	filename := fmt.Sprintf("/tmp/nestor_tests/%d", nanos)
	defer os.Remove(filename)
	grabtest.WithTestServer(t, func(url string) {
		err := d.Download(filename, url)
		if err != nil {
			t.Fatalf("expected no error. got %v", err)
		}
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			t.Fatalf("expected file in %s. got nil", filename)
		}
	})
}

func TestExecute_Downloader(t *testing.T) {
	d := nestor.NewDownloader()
	now := time.Now()
	nanos := now.UnixNano()
	filename := fmt.Sprintf("/tmp/nestor_tests/%d", nanos)
	defer os.Remove(filename)
	grabtest.WithTestServer(t, func(url string) {
		_, err := d.Execute(filename, url)
		if err != nil {
			t.Fatalf("expected no error. got %v", err)
		}
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			t.Fatalf("expected file in %s. got nil", filename)
		}
	})
}

func TestDownloadBatch(t *testing.T) {
	tests := 32
	d := nestor.NewDownloader()
	now := time.Now()
	nanos := now.UnixNano()
	filename := fmt.Sprintf("/tmp/nestor_tests/batch%d/", nanos)
	os.MkdirAll(filename, os.ModePerm)
	defer os.RemoveAll(filename)
	grabtest.WithTestServer(t, func(url string) {
		urls := make([]string, tests)
		for i := 0; i < len(urls); i++ {
			urls[i] = url + fmt.Sprintf("/request_%d?", i+1)
		}
		err := d.DownloadBatch(filename, nil, urls...)
		if err != nil {
			t.Fatalf("expected no error. got %v", err)
		}
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			t.Fatalf("expected file in %s. got nil", filename)
		}
		i, err := fileCount(filename)
		if err != nil {
			t.Fatalf("expected no error. got %v", err)
		}
		if i != tests {
			t.Fatalf("expected %d files. got %d", tests, i)
		}
	})
}

/*func TestAfterCopyHookPositive(t *testing.T) {
	now := time.Now()
	nanos := now.UnixNano()
	filename := fmt.Sprintf("/tmp/nestor_tests/%d", nanos)
	defer os.Remove(filename)
	d := nestor.NewDownloader()
	grabtest.WithTestServer(t, func(url string) {
		err := d.Download(filename, func(filepath string) error {
			if _, err := os.Stat(filepath); os.IsNotExist(err) {
				return fmt.Errorf("expected file in %s. got nil", filepath)
			}
			return nil
		}, url)
		if err != nil {
			t.Fatalf("expected no error. got %v", err)
		}
	})
}

func TestAfterCopyHookNegative(t *testing.T) {
	now := time.Now()
	nanos := now.UnixNano()
	filename := fmt.Sprintf("/tmp/nestor_tests/%d", nanos)
	defer os.Remove(filename)
	d := nestor.NewDownloader()
	grabtest.WithTestServer(t, func(url string) {
		err := d.Download(filename, func(filepath string) error {
			if _, err := os.Stat(filepath); os.IsNotExist(err) {
				return nil
			}
			return fmt.Errorf("expected file in %s. got nil", filepath)
		}, url)
		if err == nil {
			t.Fatalf("expected error. got nil")
		}
	})
}*/
