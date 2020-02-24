package nestor

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	vaultAPI "github.com/hashicorp/vault/api"
)

const (
	defaultPasswordKey string = "value"
)

//DBFactory is a class to construct DB instances from connection strings.
type DBFactory struct {
	readFromVault        bool
	vaultPasswordPath    string
	vaultPasswordKeyName string
}

//GetSQLDB is a singleton function to return a sql database based on package configs
func GetSQLDB(url string) (*sql.DB, error) {
	return nil, nil
}

//DBPropertyRetriever is a struct to loop through a databaser connection string to retrieve properites
type DBPropertyRetriever struct {
	readFromVault        bool
	vaultPasswordPath    string
	vaultPasswordKeyName string
	vaultRetriever       *vaultRetriever
}

func cleanup(input string, characters string) string {
	for _, s := range characters {
		input = strings.Trim(input, string(s))
	}
	return input
}

//GetPassword returns a password either from connection string (if found)
// of from a vault path based on predefined configurations
func (dbr *DBPropertyRetriever) GetPassword(connectionString string) string {
	if dbr.readFromVault {
		keyName := defaultPasswordKey
		if dbr.vaultPasswordKeyName != "" {
			keyName = dbr.vaultPasswordKeyName
		}
		val, err := dbr.vaultRetriever.getSecretFromPath(dbr.vaultPasswordPath, keyName)
		if err != nil {
			return ""
		}
		return val.(string)
	}
	regex := regexp.MustCompile(`\:(.*?)\@`) //password
	//TODO: implement vault retrieval
	return cleanup(regex.FindString(connectionString), ":@/?")
}

//GetUsername returns a username from connection string
func (dbr *DBPropertyRetriever) GetUsername(connectionString string) string {
	regex := regexp.MustCompile(`^(.*?)[\:|\@]`) //username
	return cleanup(regex.FindString(connectionString), ":@/?")
}

//GetURL returns a url+port(if found) from connection string.
func (dbr *DBPropertyRetriever) GetURL(connectionString string) string {
	regex := regexp.MustCompile(`\@(.*?)\/`) //url+port
	return cleanup(regex.FindString(connectionString), ":@/?")
}

//GetDBName returns dbname from connection string
func (dbr *DBPropertyRetriever) GetDBName(connectionString string) string {
	regex := regexp.MustCompile(`\/(.*?)(\?|$)`) //dbname
	return cleanup(regex.FindString(connectionString), ":@/?")
}

//GetParameters returns db parameters from connection string
func (dbr *DBPropertyRetriever) GetParameters(connectionString string) string {
	regex := regexp.MustCompile(`\?(.*?)$`) //parameters
	return cleanup(regex.FindString(connectionString), ":@/?")
}

//NewDBPropertyRetriever is the contructor for DBPropertyRetriever class
func NewDBPropertyRetriever() *DBPropertyRetriever {
	return &DBPropertyRetriever{}
}

//NewDBPropertyRetrieverWithVault is the contructor for DBPropertyRetriever class with vault retriever
func NewDBPropertyRetrieverWithVault(vaultURL string, vaultToken string, vaultPasswordPath string, vaultPasswordKeyName string) (*DBPropertyRetriever, error) {
	vr, err := newVaultRetriever(vaultURL, vaultToken)
	if err != nil {
		return nil, err
	}
	return &DBPropertyRetriever{
		readFromVault:        true,
		vaultPasswordKeyName: vaultPasswordKeyName,
		vaultPasswordPath:    vaultPasswordPath,
		vaultRetriever:       vr,
	}, nil
}

type vaultRetriever struct {
	client *vaultAPI.Client
}

func (vr *vaultRetriever) getSecretFromPath(path string, keyName string) (interface{}, error) {
	secretValues, err := vr.client.Logical().Read(path)
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

func newVaultRetriever(url string, token string) (*vaultRetriever, error) {
	cfg := vaultAPI.DefaultConfig()
	cfg.Address = url
	c, err := vaultAPI.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	c.SetToken(token)
	return &vaultRetriever{
		client: c,
	}, nil
}
