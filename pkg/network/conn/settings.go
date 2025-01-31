package conn

import "time"

var (
	_ ISettings = &sSettings{}
)

const (
	cNetworkKey  = "network-key"
	cMessageSize = (1 << 20)
	cMaxVoidSize = 1
	cTimeWait    = time.Minute
)

type SSettings sSettings
type sSettings struct {
	FNetworkKey  string
	FMessageSize uint64
	FMaxVoidSize uint64
	FTimeWait    time.Duration
}

func NewSettings(pSett *SSettings) ISettings {
	return (&sSettings{
		FNetworkKey:  pSett.FNetworkKey,
		FMessageSize: pSett.FMessageSize,
		FMaxVoidSize: pSett.FMaxVoidSize,
		FTimeWait:    pSett.FTimeWait,
	}).useDefaultValues()
}

func (p *sSettings) useDefaultValues() ISettings {
	if p.FNetworkKey == "" {
		p.FNetworkKey = cNetworkKey
	}
	if p.FMessageSize == 0 {
		p.FMessageSize = cMessageSize
	}
	if p.FMaxVoidSize == 0 {
		p.FMaxVoidSize = cMaxVoidSize
	}
	if p.FTimeWait == 0 {
		p.FTimeWait = cTimeWait
	}
	return p
}

func (p *sSettings) GetNetworkKey() string {
	return p.FNetworkKey
}

func (p *sSettings) GetMessageSize() uint64 {
	return p.FMessageSize
}

func (p *sSettings) GetMaxVoidSize() uint64 {
	return p.FMaxVoidSize
}

func (p *sSettings) GetTimeWait() time.Duration {
	return p.FTimeWait
}
