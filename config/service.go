package config

import (
	"encoding/json"
	"fmt"
)

// Service struct is used for configuring how the service will run
// by reading the values from the environment or using the default values.
type Service struct {
	Port   string `envconfig:"serviceport,default=8017" json:"Port"`
	DBConnection1	string `envconfig:"dbconnection1`
	DBConnection2	string `envconfig:"dbconnection2`
}

// String returns a printable representation of the config as JSON.
// Use the struct field tag `json:"-"` to hide fields that should not be revealed such as credentials and secrets.
func (s Service) String() string {
	json, err := json.Marshal(s)
	if err != nil {
		return fmt.Sprintf("Error marshalling service configuration JSON: %v", err)
	}
	return fmt.Sprintf("Service Configuration: %s", json)
}
