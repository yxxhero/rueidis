package conn

import "sync"

func newPool(cap int, makeFn func() Wire) *pool {
	if cap <= 0 {
		cap = DefaultPoolSize
	}

	return &pool{
		size: 0,
		make: makeFn,
		list: make([]Wire, 0, cap),
		cond: sync.NewCond(&sync.Mutex{}),
	}
}

type pool struct {
	list []Wire
	cond *sync.Cond
	make func() Wire
	size int
}

func (p *pool) Acquire() (v Wire) {
	p.cond.L.Lock()
	for len(p.list) == 0 && p.size == cap(p.list) {
		p.cond.Wait()
	}
	if len(p.list) == 0 {
		v = p.make()
		p.size++
	} else {
		i := len(p.list) - 1
		v = p.list[i]
		p.list = p.list[:i]
	}
	p.cond.L.Unlock()
	return
}

func (p *pool) Store(v Wire) {
	p.cond.L.Lock()
	if v.Error() == nil {
		p.list = append(p.list, v)
	} else {
		p.size--
	}
	p.cond.L.Unlock()
	p.cond.Signal()
}
