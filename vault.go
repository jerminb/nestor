package nestor

import (
	"errors"
	"fmt"

	vaultAPI "github.com/hashicorp/vault/api"
)

var (
	ErrRenewerNotRenewable = errors.New("secret is not renewable")
)

//VaultService is an abstraction around vault to expose the necessary functionality
// without having to go through hashicorp apis
type VaultService struct {
	client *vaultAPI.Client
}

//GetSecretFromPath returns secret from a given path with a given keyname
func (vs *VaultService) GetSecretFromPath(path string, keyName string) (interface{}, error) {
	secretValues, err := vs.client.Logical().Read(path)
	if err != nil {
		return nil, err
	}
	if secretValues == nil {
		return nil, fmt.Errorf("value for keyname %s under path %s not found", keyName, path)
	}
	for propName, propValue := range secretValues.Data {
		if propName == keyName {
			return propValue, nil
		}
	}
	return nil, fmt.Errorf("value for keyname %s not found", keyName)
}

//RenewSelfToken renews client token
func (vs *VaultService) RenewSelfToken() error {
	selfSecret, err := vs.client.Auth().Token().LookupSelf()
	if err != nil {
		return err
	}
	if !selfSecret.Renewable {
		return ErrRenewerNotRenewable
	}
	_, err = vs.client.Auth().Token().RenewSelf(0)
	if err != nil {
		return err
	}
	return nil
}

//NewVaultService is constructor for VaultService
func NewVaultService(url string, token string) (*VaultService, error) {
	cfg := vaultAPI.DefaultConfig()
	cfg.Address = url
	c, err := vaultAPI.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	c.SetToken(token)
	return &VaultService{
		client: c,
	}, nil
}
