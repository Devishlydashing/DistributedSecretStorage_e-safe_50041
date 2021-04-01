package config

import (
	"hash/fnv"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"gopkg.in/yaml.v2"
)

type Config struct {
	ConfigServer    `yaml:"Server"`
	ConfigNode      `yaml:"Node"`
	ConfigTimeout   `yaml:"Timeout"`
	ConfigLocksmith `yaml:"Locksmith"`
}

type ConfigServer struct {
	Host string `yaml:"Host"`
	Port int    `yaml:"Port"`
}

type ConfigLocksmith struct {
	Port int `yaml:"Port"`
}

type ConfigNode struct {
	Number            int `yaml:"Number"`
	HeartbeatInterval int `yaml:"HeartbeatInterval"`
	VirtualNodesCount int `yaml:"VirtualNodesCount"`
}

type ConfigTimeout struct {
	HeartBeatTimeout    int `yaml:"HeartbeatTimeout"`
	NodeCreationTimeout int `yaml:"NodeCreationTimeout"`
}

// LoadConfig loads the config from YAML file and return the config object
func LoadConfig(path string) (*Config, error) {
	cfg := &Config{}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)

	if err := d.Decode(&cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

var lock = &sync.Mutex{}

var globalConfig *Config

func GetHash(s string) (uint32, error) {
	h := fnv.New32a()
	_, err := h.Write([]byte(s))
	if err != nil {
		return 0, err
	}
	return h.Sum32(), nil
}

// GetConfig is a singleton method that gets the loaded configuration object
func GetConfig() (*Config, error) {
	if globalConfig == nil {
		lock.Lock()
		defer lock.Unlock()
		if globalConfig == nil {
			_, file, _, _ := runtime.Caller(0)
			paths := strings.Split(file, "/")
			paths = paths[:len(paths)-2]
			rootPath := "/" + filepath.Join(paths...)
			configPath := filepath.Join(rootPath, "config.yaml")
			config, err := LoadConfig(configPath)
			if err != nil {
				return nil, err
			}
			globalConfig = config
		}
	}
	return globalConfig, nil
}
