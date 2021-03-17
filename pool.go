package pool

import (
	"sync"
	"sync/atomic"
)

type ReferenceCountable interface {
	SetInstance(i Reseter)
	IncrementReferenceCount()
	IncrementReferenceCountByN(n uint32)
	DecrementReferenceCount()
	DecrementReferenceCountByN(n uint32)
	Reseter
}

type referenceCountedPool struct {
	pool *sync.Pool
}

func NewReferenceCountedPool(factory func(referenceCounter ReferenceCounter) ReferenceCountable) *referenceCountedPool {
	p := new(referenceCountedPool)
	p.pool = new(sync.Pool)
	p.pool.New = func() interface{} {
		o := factory(ReferenceCounter{
			count:       new(uint32),
			destination: p.pool,
		})
		return o
	}
	return p
}

func (p *referenceCountedPool) Get() ReferenceCountable {
	o := p.pool.Get().(ReferenceCountable)
	o.SetInstance(o)
	o.IncrementReferenceCount()
	return o
}

type ReferenceCounter struct {
	count       *uint32    `sql:"-" yaml:"-"`
	destination *sync.Pool `sql:"-" yaml:"-"`
	instance    Reseter    `sql:"-" yaml:"-"`
}

func (r ReferenceCounter) IncrementReferenceCount() {
	atomic.AddUint32(r.count, 1)
}

func (r ReferenceCounter) IncrementReferenceCountByN(n uint32) {
	atomic.AddUint32(r.count, n)
}

type Reseter interface {
	Reset()
}

func (r ReferenceCounter) DecrementReferenceCount() {
	if atomic.LoadUint32(r.count) == 0 {
		panic("this should not happen")
	}
	if atomic.AddUint32(r.count, ^uint32(0)) == 0 {
		r.instance.Reset()
		r.destination.Put(r.instance)
		r.instance = nil
	}
}

func (r ReferenceCounter) DecrementReferenceCountByN(n uint32) {
	for ; n > 0; n -= 1 {
		r.DecrementReferenceCount()
	}
}

func (r *ReferenceCounter) SetInstance(i Reseter) {
	r.instance = i
}
