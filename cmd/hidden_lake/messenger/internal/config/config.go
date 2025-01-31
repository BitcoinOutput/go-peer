package config

import (
	"fmt"

	"github.com/number571/go-peer/pkg/encoding"
	"github.com/number571/go-peer/pkg/filesystem"
	"github.com/number571/go-peer/pkg/logger"
)

var (
	_ IConfig     = &SConfig{}
	_ IAddress    = &SAddress{}
	_ IConnection = &SConnection{}
)

type SConfig struct {
	FLogging    []string     `json:"logging,omitempty"`
	FAddress    *SAddress    `json:"address"`
	FConnection *SConnection `json:"connection"`
	FStorageKey string       `json:"storage_key,omitempty"`

	fLogging *sLogging
}

type sLogging []bool

type SAddress struct {
	FInterface string `json:"interface"`
	FIncoming  string `json:"incoming"`
}

type SConnection struct {
	FService string `json:"service"`
	FTraffic string `json:"traffic,omitempty"`
}

func BuildConfig(pFilepath string, pCfg *SConfig) (IConfig, error) {
	configFile := filesystem.OpenFile(pFilepath)

	if configFile.IsExist() {
		return nil, fmt.Errorf("config file '%s' already exist", pFilepath)
	}

	if err := configFile.Write(encoding.Serialize(pCfg)); err != nil {
		return nil, err
	}

	if err := pCfg.loadLogging(); err != nil {
		return nil, err
	}
	return pCfg, nil
}

func LoadConfig(pFilepath string) (IConfig, error) {
	configFile := filesystem.OpenFile(pFilepath)

	if !configFile.IsExist() {
		return nil, fmt.Errorf("config file '%s' does not exist", pFilepath)
	}

	bytes, err := configFile.Read()
	if err != nil {
		return nil, err
	}

	cfg := new(SConfig)
	if err := encoding.Deserialize(bytes, cfg); err != nil {
		return nil, err
	}

	if err := cfg.loadLogging(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (p *SConfig) loadLogging() error {
	// [info, warn, erro]
	logging := sLogging(make([]bool, 3))

	mapping := map[string]int{
		"info": 0,
		"warn": 1,
		"erro": 2,
	}

	for _, v := range p.FLogging {
		logType, ok := mapping[v]
		if !ok {
			return fmt.Errorf("undefined log type '%s'", v)
		}
		logging[logType] = true
	}

	p.fLogging = &logging
	return nil
}

func (p *SConfig) GetAddress() IAddress {
	return p.FAddress
}

func (p *SConfig) GetConnection() IConnection {
	return p.FConnection
}

func (p *SConfig) GetStorageKey() string {
	return p.FStorageKey
}

func (p *SConnection) GetService() string {
	return p.FService
}

func (p *SConnection) GetTraffic() string {
	return p.FTraffic
}

func (p *SAddress) GetInterface() string {
	return p.FInterface
}

func (p *SAddress) GetIncoming() string {
	return p.FIncoming
}

func (p *SConfig) GetLogging() logger.ILogging {
	return p.fLogging
}

func (p *sLogging) HasInfo() bool {
	return (*p)[0]
}

func (p *sLogging) HasWarn() bool {
	return (*p)[1]
}

func (p *sLogging) HasErro() bool {
	return (*p)[2]
}
