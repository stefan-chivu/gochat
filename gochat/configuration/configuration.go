// Copyright 2020 Netflix Inc
// Author: Colin McIntosh (colin@netflix.com)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package configuration contains the ServerConfig type that is used by the gateway package
// and all of it's sub-packages.
package configuration

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/rs/zerolog"
	"github.com/stefan-chivu/gochat/gochat/room"
)

// ServerConfig contains all of the configurables and tunables for various components of the gateway.
// Many of these options may be set via command-line flags. See main.go for details on flags that
// are available.
type ServerConfig struct {
	mu sync.RWMutex
	// ClientTLSConfig are the gochat client TLS credentials. Setting this will enable client TLS.
	// TODO Add options to set client certificates by path (i.e. like the server TLS creds).
	ClientTLSConfig *tls.Config `ignored:"true"`
	// Log is the logger used by the gateway code and gateway packages.
	Log zerolog.Logger
	// LogCaller will add the file path and line number to all log messages.
	LogCaller bool `json:"log_caller"`
	// Rooms represent the rooms currently available on the server
	Rooms []*room.Room
	// ServerAddress is the address where other cluster members can reach the gochat server.
	// The first assigned IP address is used if the parameter is not provided.
	ServerAddress string `json:"server_address"`
	// ServerPort is the TCP port where other cluster members can reach the gochat server.
	// ServerListenPort is used if the parameter is not provided.
	ServerPort int `json:"server_port"`
	// ServerListenAddress is the interface IP address the gochat server will listen on.
	ServerListenAddress string `json:"server_listen_address"`
	// ServerListenPort is the TCP port the gochat server will listen on.
	ServerListenPort int `json:"server_listen_port"`
	// ServerTLSCert is the path to the file containing the PEM-encoded x509 gochat server TLS certificate.
	// See the gateway package for instructions for generating a self-signed certificate.
	ServerTLSCert string `json:"server_tls_cert"`
	// ServerTLSCert is the path to the file containing the PEM-encoded x509 gochat server TLS key.
	// See the gateway package for instructions for generating a self-signed certificate key.
	ServerTLSKey string `json:"server_tls_key"`
}

func NewDefaultServerConfig() *ServerConfig {
	config := &ServerConfig{
		Log: zerolog.New(os.Stderr).With().Timestamp().Logger().Level(zerolog.InfoLevel),
	}
	return config
}

func NewServerConfigFromFile(filePath string) (*ServerConfig, error) {
	var config ServerConfig
	err := PopulateServerConfigFromFile(&config, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create new config from file: %v", err)
	}
	return &config, nil
}

func PopulateServerConfigFromFile(config *ServerConfig, filePath string) error {
	path := filepath.Clean(filePath)
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file at '%s': %v", path, err)
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(config); err != nil {
		return fmt.Errorf("failed to parse config file: %v", err)
	}

	return nil
}
