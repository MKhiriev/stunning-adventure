package utils

import (
	"sync"
)

type Resettable interface{ Reset() }

type Pool[T Resettable] struct {
	*sync.Pool
}

func New[T Resettable](newFunc func() T) *Pool[T] {
	return &Pool[T]{
		Pool: &sync.Pool{
			New: func() any {
				return newFunc()
			},
		},
	}
}

func (p *Pool[T]) Get() T {
	return p.Pool.Get().(T)
}

func (p *Pool[T]) Put(obj T) {
	obj.Reset()
	p.Pool.Put(obj)
}
