// database.go
package database

import (
	"context"
	"io/ioutil"
	"log"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Database struct {
		URI  string `yaml:"uri"`
		Name string `yaml:"name"`
	} `yaml:"database"`
}

type DB struct {
	Client *mongo.Client
	DB     *mongo.Database
}

func NewDB(logger *log.Logger, configPath string) (*DB, error) {
	// Load the configuration from the YAML file
	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read database config file")
	}

	var config Config
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal database config data")
	}

	// Set client options
	clientOptions := options.Client().ApplyURI(config.Database.URI)

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to MongoDB")
	}

	// Check the connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to ping MongoDB")
	}

	db := client.Database(config.Database.Name)

	return &DB{Client: client, DB: db}, nil
}
