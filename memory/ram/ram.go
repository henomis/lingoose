// Package ram provides a memory storage that stores data in RAM.
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

func (r *Ram) Get(key string) interface{} {
	value, ok := r.memory[key]
	if !ok {
		return nil
	}
	return value
}

func (r *Ram) Set(key string, value interface{}) error {
	r.memory[key] = value
	return nil
}

func (r *Ram) All() map[string]interface{} {
	return r.memory
}

func (r *Ram) Delete(key string) error {
	_, ok := r.memory[key]
	if !ok {
		return ErrObjectNotFound
	}

	delete(r.memory, key)
	return nil
}

func (r *Ram) Clear() error {
	r.memory = map[string]interface{}{}
	return nil
}
