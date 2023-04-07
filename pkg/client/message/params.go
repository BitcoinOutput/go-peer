package message

var (
	_ IParams = &sParams{}
)

type sParams struct {
	fMessageSize uint64
	fWorkSize    uint64
}

func NewParams(pMsgSize, pWorkSize uint64) IParams {
	return &sParams{
		fMessageSize: pMsgSize,
		fWorkSize:    pWorkSize,
	}
}

func (p *sParams) GetMessageSize() uint64 {
	return p.fMessageSize
}

func (p *sParams) GetWorkSize() uint64 {
	return p.fWorkSize
}