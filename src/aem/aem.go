package aem

type Monitor struct {}

func NewMonitor() *Monitor {
	return &Monitor{}
}

func (m *Monitor) Score() int {
	return 100
}
