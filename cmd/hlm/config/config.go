package config

import (
	"fmt"

	"github.com/number571/go-peer/modules/encoding"
	"github.com/number571/go-peer/modules/filesystem"
)

var (
	_ IConfig  = &SConfig{}
	_ iAddress = &SAddress{}
)

type SConfig struct {
	FAddress    *SAddress `json:"address"`
	FConnection string    `json:"connection"`
}

type SAddress struct {
	FWebLocal string `json:"web_local"`
	FIncoming string `json:"incoming"`
}

func NewConfig(filepath string, cfg *SConfig) (IConfig, error) {
	configFile := filesystem.OpenFile(filepath)

	if configFile.IsExist() {
		return nil, fmt.Errorf("config file '%s' already exist", filepath)
	}

	if err := configFile.Write(encoding.Serialize(cfg)); err != nil {
		return nil, err
	}

	return cfg, nil
}

func LoadConfig(filepath string) (IConfig, error) {
	configFile := filesystem.OpenFile(filepath)

	if !configFile.IsExist() {
		return nil, fmt.Errorf("config file '%s' does not exist", filepath)
	}

	bytes, err := configFile.Read()
	if err != nil {
		return nil, err
	}

	cfg := new(SConfig)
	if err := encoding.Deserialize(bytes, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (cfg *SConfig) Address() iAddress {
	return cfg.FAddress
}

func (cfg *SConfig) Connection() string {
	return cfg.FConnection
}

func (addr *SAddress) WebLocal() string {
	return addr.FWebLocal
}

func (addr *SAddress) Incoming() string {
	return addr.FIncoming
}
