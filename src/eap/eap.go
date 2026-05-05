package eap

type Provider struct {}

func NewProvider() *Provider {
	return &Provider{}
}

func (p *Provider) Bind() bool {
	return true
}
