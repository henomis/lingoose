package ram

import "errors"

var (
	ErrObjectNotFound = errors.New("object not found")
)

type Ram struct {
	memory map[string]interface{}
}

func New() *Ram {
	return &Ram{
		memory: map[string]interface{}{},
	}
}

func (m *Ram) Get(key string) interface{} {
	value, ok := m.memory[key]
	if !ok {
		return nil
	}
	return value
}

func (m *Ram) Set(key string, value interface{}) error {
	m.memory[key] = value
	return nil
}

func (m *Ram) All() map[string]interface{} {
	return m.memory
}

func (m *Ram) Delete(key string) error {
	_, ok := m.memory[key]
	if !ok {
		return ErrObjectNotFound
	}

	delete(m.memory, key)
	return nil
}

func (m *Ram) Clear() error {
	m.memory = map[string]interface{}{}
	return nil
}
