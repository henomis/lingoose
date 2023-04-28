// Package ram provides a memory storage that stores data in RAM.
package ram

import (
	"errors"

	"github.com/henomis/lingoose/types"
)

var (
	ErrObjectNotFound = errors.New("object not found")
)

type ram struct {
	memory types.M
}

func New() *ram {
	return &ram{
		memory: types.M{},
	}
}

func (r *ram) Get(key string) interface{} {
	value, ok := r.memory[key]
	if !ok {
		return nil
	}
	return value
}

func (r *ram) Set(key string, value interface{}) error {
	r.memory[key] = value
	return nil
}

func (r *ram) All() types.M {
	return r.memory
}

func (r *ram) Delete(key string) error {
	_, ok := r.memory[key]
	if !ok {
		return ErrObjectNotFound
	}

	delete(r.memory, key)
	return nil
}

func (r *ram) Clear() error {
	r.memory = types.M{}
	return nil
}
