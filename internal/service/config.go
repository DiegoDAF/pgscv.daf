// Package service is a pgSCV service helper
package service

import (
	"fmt"
	"os"
	"strings"

	"github.com/cherts/pgscv/internal/log"
	"github.com/cherts/pgscv/internal/model"
	"github.com/cherts/pgscv/internal/tunnel"
)

// Label struct describe targets labels
type Label struct {
	Name  string `yaml:"name" json:"name"`
	Value string `yaml:"value" json:"value"`
}

// ConnSetting describes connection settings required for connecting to particular service.
// This is primarily used for describing services defined by user in the config file (or env vars).
type ConnSetting struct {
	// ServiceType defines type of service for which these connection settings are used.
	ServiceType string `yaml:"service_type"`
	// Conninfo is the connection string in service-specific format.
	Conninfo string `yaml:"conninfo"`
	// ConninfoFile is a path to a file containing the connection string (alternative to Conninfo).
	ConninfoFile string `yaml:"conninfo_file"`
	// BaseURL is the base URL for connecting to HTTP services.
	BaseURL string `yaml:"baseurl"`
	// TargetLabels array of labels for /targets endpoint
	TargetLabels *[]Label `yaml:"target_labels"`
	// SSHTunnel is the SSH tunnel configuration for this service (optional).
	SSHTunnel *tunnel.SSHTunnelConfig `yaml:"ssh_tunnel"`
}

// ResolveConninfo reads the connection string from ConninfoFile if set, falling back to Conninfo.
func (cs *ConnSetting) ResolveConninfo() error {
	if cs.ConninfoFile == "" {
		return nil
	}
	data, err := os.ReadFile(cs.ConninfoFile)
	if err != nil {
		return fmt.Errorf("failed to read conninfo_file %s: %w", cs.ConninfoFile, err)
	}
	cs.Conninfo = strings.TrimSpace(string(data))
	log.Infof("conninfo loaded from file %s", cs.ConninfoFile)
	return nil
}

// ConnsSettings defines a set of all connection settings of exact services.
type ConnsSettings map[string]ConnSetting

// ParsePostgresDSNEnv is a public wrapper over parseDSNEnv.
func ParsePostgresDSNEnv(key, value string) (string, ConnSetting, error) {
	return parseDSNEnv("POSTGRES_DSN", strings.Replace(key, "DATABASE_DSN", "POSTGRES_DSN", 1), value)
}

// ParsePgbouncerDSNEnv is a public wrapper over parseDSNEnv.
func ParsePgbouncerDSNEnv(key, value string) (string, ConnSetting, error) {
	return parseDSNEnv("PGBOUNCER_DSN", key, value)
}

// parseDSNEnv returns valid ConnSetting accordingly to passed prefix and environment key/value.
func parseDSNEnv(prefix, key, value string) (string, ConnSetting, error) {
	var stype string
	switch prefix {
	case "POSTGRES_DSN":
		stype = model.ServiceTypePostgresql
	case "PGBOUNCER_DSN":
		stype = model.ServiceTypePgbouncer
	default:
		return "", ConnSetting{}, fmt.Errorf("invalid prefix %s", prefix)
	}

	// Prefix must be the part of key.
	if !strings.HasPrefix(key, prefix) {
		return "", ConnSetting{}, fmt.Errorf("invalid key %s", key)
	}

	// Nothing to parse if prefix and key are the same, just use the type as service ID.
	if key == prefix {
		return stype, ConnSetting{ServiceType: stype, Conninfo: value}, nil
	}

	// If prefix and key are not the same, strip prefix from key and use the rest as service ID.
	// Use double Trim to avoid leaking 'prefix' string into ID value (see unit tests for examples).
	id := strings.TrimPrefix(strings.TrimPrefix(key, prefix), "_")

	if id == "" {
		return "", ConnSetting{}, fmt.Errorf("invalid value '%s' is in %s", value, key)
	}

	return id, ConnSetting{ServiceType: stype, Conninfo: value}, nil
}

// ParsePatroniURLEnv is a public wrapper over parseURLEnv.
func ParsePatroniURLEnv(key, value string) (string, ConnSetting, error) {
	return parseURLEnv("PATRONI_URL", key, value)
}

// parseURLEnv returns valid ConnSetting accordingly to passed prefix and environment key/value.
func parseURLEnv(prefix, key, value string) (string, ConnSetting, error) {
	var stype string
	switch prefix {
	case "PATRONI_URL":
		stype = model.ServiceTypePatroni
	default:
		return "", ConnSetting{}, fmt.Errorf("invalid prefix %s", prefix)
	}

	// Prefix must be the part of key.
	if !strings.HasPrefix(key, prefix) {
		return "", ConnSetting{}, fmt.Errorf("invalid key %s", key)
	}

	// Nothing to parse if prefix and key are the same, just use the type as service ID.
	if key == prefix {
		return stype, ConnSetting{ServiceType: stype, BaseURL: value}, nil
	}

	// If prefix and key are not the same, strip prefix from key and use the rest as service ID.
	// Use double Trim to avoid leaking 'prefix' string into ID value (see unit tests for examples).
	id := strings.TrimPrefix(strings.TrimPrefix(key, prefix), "_")

	if id == "" {
		return "", ConnSetting{}, fmt.Errorf("invalid value '%s' is in %s", value, key)
	}

	return id, ConnSetting{ServiceType: stype, BaseURL: value}, nil
}
