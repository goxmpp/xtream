package xtream

import (
	"encoding/xml"
	"sync"
)

var NodeRegistry = NewRegistry()

type Registry interface {
	Add(cons Constructor, outer, inner xml.Name)
	Get(outer, inner *xml.Name) interface{}
}
type Constructor func() interface{}

type outerNodesRegistry struct {
	reg map[xml.Name]*innerNodesRegistry
	mx  sync.RWMutex
}

type innerNodesRegistry struct {
	reg map[xml.Name]Constructor
	mx  sync.RWMutex
}

func NewRegistry() *outerNodesRegistry {
	return &outerNodesRegistry{reg: make(map[xml.Name]*innerNodesRegistry)}
}

func newINRegistry() *innerNodesRegistry {
	return &innerNodesRegistry{reg: make(map[xml.Name]Constructor)}
}

func (r *outerNodesRegistry) Add(cons Constructor, outer, inner xml.Name) {
	r.mx.Lock() // Think about CAS here
	reg := r.lookup(&outer)
	if reg == nil {
		reg = newINRegistry()
		r.reg[outer] = reg
	}
	r.mx.Unlock()

	reg.add(cons, &inner)
}

func (r *outerNodesRegistry) Get(outer, inner *xml.Name) interface{} {
	r.mx.RLock()
	reg := r.lookup(outer)
	r.mx.RUnlock()

	if reg == nil {
		return nil
	}

	reg.mx.RLock()
	cons, ok := reg.reg[*inner]
	if !ok {
		inner_anyns := xml.Name{Local: inner.Local}
		cons, ok = reg.reg[inner_anyns]
	}
	reg.mx.RUnlock()
	if ok {
		obj := cons()
		if innerEl, ok := obj.(Registrable); ok {
			innerEl.SetRegistry(r)
		}
		return obj
	}

	return nil
}

func (r *outerNodesRegistry) lookup(node *xml.Name) *innerNodesRegistry {
	nsr, ok := r.reg[*node]
	if !ok {
		node_anyns := xml.Name{Local: node.Local}
		nsr = r.reg[node_anyns]
	}
	return nsr
}

func (nsr *innerNodesRegistry) add(cons Constructor, node *xml.Name) {
	nsr.mx.Lock()
	if _, ok := nsr.reg[*node]; !ok {
		nsr.reg[*node] = cons
	} else {
		panic("Node already registered " + node.Local)
	}
	nsr.mx.Unlock()
}
