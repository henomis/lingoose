// Package ram provides a memory storage that stores data in RAM.
package ram

import (
	"errors"

	"github.com/rsest/lingoose/types"
)

var (
	ErrObjectNotFound = errors.New("object not found")
)

//nolint:revive,stylecheck
type Ram struct {
	memory types.M
}

func New() *Ram {
	return &Ram{
		memory: types.M{},
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

func (r *Ram) All() types.M {
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
	r.memory = types.M{}
	return nil
}
