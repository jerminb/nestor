package nestor

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/hashicorp/vault/api"
)

// VaultService is the main interface into the vault API
type VaultService struct {
	vaultURL string
	// the vault client
	client *api.Client
	// skip tls verify
	skipTLSVerify bool
}

// buildHTTPTransport constructs a http transport for the http client
func buildHTTPTransport(skipTLSVerify bool, vaultCaFile string) (*http.Transport, error) {
	// step: create the vault sidekick
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 10 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipTLSVerify,
		},
	}

	if vaultCaFile != "" {
		caCert, err := ioutil.ReadFile(vaultCaFile)
		if err != nil {
			return nil, fmt.Errorf("unable to read in the ca: %s, reason: %s", vaultCaFile, err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		transport.TLSClientConfig.RootCAs = caCertPool
	}

	return transport, nil
}

// newVaultClient creates and authenticates a vault client
func newVaultClient(url string, token string, skipTLSVerify bool, vaultCaFile string) (*api.Client, error) {
	var err error

	config := api.DefaultConfig()
	config.Address = url

	config.HttpClient.Transport, err = buildHTTPTransport(skipTLSVerify, vaultCaFile)
	if err != nil {
		return nil, err
	}

	// step: create the actual client
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	// step: set the token for the client
	client.SetToken(token)
	return client, nil
}

//NewVaultService is constructor for VaultService struct
func NewVaultService(url string, token string, skipTLSVerify bool, vaultCaFile string) (*VaultService, error) {
	var err error

	// step: create the config for client
	service := new(VaultService)
	service.vaultURL = url
	service.client, err = newVaultClient(url, token, skipTLSVerify, vaultCaFile)

	if err != nil {
		return nil, err
	}

	return service, nil
}
