package pool

import (
	"math"
	"sync"
	"sync/atomic"
)

type ReferenceCountable interface {
	setInstance(i Reseter)
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
		o.setInstance(o)
		return o
	}
	return p
}

func (p *referenceCountedPool) Get() ReferenceCountable {
	o := p.pool.Get().(ReferenceCountable)
	o.IncrementReferenceCount()
	return o
}

type ReferenceCounter struct {
	count       *uint32    `sql:"-" yaml:"-" json:"-"`
	destination *sync.Pool `sql:"-" yaml:"-" json:"-"`
	instance    Reseter    `sql:"-" yaml:"-" json:"-"`
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
	if count := atomic.AddUint32(r.count, ^uint32(0)); count == 0 {
		r.instance.Reset()
		r.destination.Put(r.instance)
	} else if count == math.MaxUint32 {
		panic("this should not happen")
	}
}

func (r ReferenceCounter) DecrementReferenceCountByN(n uint32) {
	for ; n > 0; n -= 1 {
		r.DecrementReferenceCount()
	}
}

func (r *ReferenceCounter) setInstance(i Reseter) {
	r.instance = i
}
