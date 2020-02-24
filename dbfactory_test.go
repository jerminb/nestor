package nestor_test

import (
	"net"
	"testing"

	"github.com/jerminb/nestor"
	"github.com/jerminb/nestor/testserver"
)

func TestPropertyRetriever(t *testing.T) {
	tests := []struct {
		Source   string
		Username string
		Password string
		URL      string
		DBName   string
		Params   string
	}{
		{"username:password@blah.blah.com:5432/dbname", "username", "password", "blah.blah.com:5432", "dbname", ""},
		{"username@blah.blah.com:5432/dbname", "username", "", "blah.blah.com:5432", "dbname", ""},
		{"username@blah.blah.com/dbname", "username", "", "blah.blah.com", "dbname", ""},
		{"username:password@blah.blah.com:5432/dbname?param1=value&param2=value", "username", "password", "blah.blah.com:5432", "dbname", "param1=value&param2=value"},
	}

	for _, tc := range tests {
		retriever := nestor.NewDBPropertyRetriever()
		if tc.Username != retriever.GetUsername(tc.Source) {
			t.Errorf("expected %s. got %s", tc.Username, retriever.GetUsername(tc.Source))
		}
		if tc.Password != retriever.GetPassword(tc.Source) {
			t.Errorf("expected %s. got %s", tc.Password, retriever.GetPassword(tc.Source))
		}
		if tc.URL != retriever.GetURL(tc.Source) {
			t.Errorf("expected %s. got %s", tc.URL, retriever.GetURL(tc.Source))
		}
		if tc.DBName != retriever.GetDBName(tc.Source) {
			t.Errorf("expected %s. got %s", tc.DBName, retriever.GetDBName(tc.Source))
		}
		if tc.Params != retriever.GetParameters(tc.Source) {
			t.Errorf("expected %s. got %s", tc.Params, retriever.GetParameters(tc.Source))
		}
	}
}

func TestPasswordRetrieverFromVault(t *testing.T) {
	testserver.WithTestVaultServer(t, func(url string, listner net.Listener, token string) {
		retriever, err := nestor.NewDBPropertyRetrieverWithVault(url, token, "secret/client-uuid/sgid/sid/bps-db/password", "value")
		if err != nil {
			t.Fatalf("expected nil. got %v", err)
		}
		pass := retriever.GetPassword("username:password@blah.blah.com:5432/dbname")
		if pass == "" {
			t.Error("expected password. got empty")
		}
		//t.Error("dummy error")
	})
}

func TestPasswordRetrieverFromVault_WrongPath(t *testing.T) {
	testserver.WithTestVaultServer(t, func(url string, listner net.Listener, token string) {
		retriever, err := nestor.NewDBPropertyRetrieverWithVault(url, token, "secret/client-uuid/sgid/", "value")
		if err != nil {
			t.Fatalf("expected nil. got %v", err)
		}
		pass := retriever.GetPassword("username:password@blah.blah.com:5432/dbname")
		if pass != "" {
			t.Errorf("expected empty. got %s", pass)
		}
	})
}

func TestPasswordRetrieverFromVault_WrongKeyname(t *testing.T) {
	testserver.WithTestVaultServer(t, func(url string, listner net.Listener, token string) {
		retriever, err := nestor.NewDBPropertyRetrieverWithVault(url, token, "secret/client-uuid/sgid/sid/bps-db/password", "wrongvalue")
		if err != nil {
			t.Fatalf("expected nil. got %v", err)
		}
		pass := retriever.GetPassword("username:password@blah.blah.com:5432/dbname")
		if pass != "" {
			t.Errorf("expected empty. got %s", pass)
		}
	})
}
