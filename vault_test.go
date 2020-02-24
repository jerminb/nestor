package nestor_test

import (
	"net"
	"testing"

	"github.com/jerminb/nestor"
	"github.com/jerminb/nestor/testserver"
)

func TestVaultConstructorPositive(t *testing.T) {
	testserver.WithTestVaultServer(t, func(url string, listner net.Listener, token string) {
		_, err := nestor.NewVaultService(url, token, false, "")
		if err != nil {
			t.Fatalf("expected nil. got %v", err)
		}
	})
}
