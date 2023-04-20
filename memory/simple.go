package memory

type SimpleMemory struct {
	memory map[string]interface{}
}

func NewSimpleMemory() *SimpleMemory {
	return &SimpleMemory{
		memory: map[string]interface{}{},
	}
}

func (m *SimpleMemory) Get(key string) interface{} {
	value, ok := m.memory[key]
	if !ok {
		return nil
	}
	return value
}

func (m *SimpleMemory) Set(key string, value interface{}) error {
	m.memory[key] = value
	return nil
}

func (m *SimpleMemory) All() map[string]interface{} {
	return m.memory
}
