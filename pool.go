package pool

import (
	"math"
	"sync"
	"sync/atomic"
)

type (
	instance interface {
		Reset()
	}
	ReferenceCountable interface {
		setInstance(i instance)
		IncrementReferenceCount()
		IncrementReferenceCountByN(n uint32)
		DecrementReferenceCount()
		DecrementReferenceCountByN(n uint32)
		instance
	}
	referenceCountedPool struct {
		pool *sync.Pool
	}
	ReferenceCounter struct {
		count       *uint32    `sql:"-" yaml:"-" json:"-"`
		destination *sync.Pool `sql:"-" yaml:"-" json:"-"`
		instance    instance   `sql:"-" yaml:"-" json:"-"`
	}
)

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

func (r ReferenceCounter) IncrementReferenceCount() {
	atomic.AddUint32(r.count, 1)
}

func (r ReferenceCounter) IncrementReferenceCountByN(n uint32) {
	atomic.AddUint32(r.count, n)
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

func (r *ReferenceCounter) setInstance(i instance) {
	r.instance = i
}
