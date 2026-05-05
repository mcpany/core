package gc_immune

type Anchor struct {}

func NewAnchor() *Anchor {
	return &Anchor{}
}

func (a *Anchor) Pin() bool {
	return true
}
