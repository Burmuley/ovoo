package milter

type Milter struct {
}

func (m *Milter) WithMilterService() *Milter {
	return m
}

func (m *Milter) Start() {
	panic("implement me!")
}
