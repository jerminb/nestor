package nestor

import (
	"database/sql"
	"regexp"
	"strings"
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
	vaultService         *VaultService
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
		val, err := dbr.vaultService.GetSecretFromPath(dbr.vaultPasswordPath, keyName)
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
	vr, err := NewVaultService(vaultURL, vaultToken)
	if err != nil {
		return nil, err
	}
	return &DBPropertyRetriever{
		readFromVault:        true,
		vaultPasswordKeyName: vaultPasswordKeyName,
		vaultPasswordPath:    vaultPasswordPath,
		vaultService:         vr,
	}, nil
}
