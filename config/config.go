// Package config provides configuration settings for the application
package config

type Config struct {
	// The port on which the server will listen
	Port int
	Mongo Mongo
}

type Mongo struct {
	// The MongoDB connection string
	ConnectionString string
}

// NewConfig creates a new Config instance with default values
func NewConfig() *Config {
	return &Config{
		Port:     8080,
		Mongo:   Mongo{
			ConnectionString: "mongodb://localhost:27017",
			// mongosh "mongodb://root:1234@localhost:27017"
		},
	}
}