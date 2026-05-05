package mrcr

type Resolver struct {}

func NewResolver() *Resolver {
	return &Resolver{}
}

func (r *Resolver) Resolve() bool {
	return true
}
