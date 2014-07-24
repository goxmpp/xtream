package xtream

import (
	"encoding/xml"
	"log"
	"reflect"
	"strings"
	"sync"
)

var NodeFactory = NewFactory()

// Factory pattern interface with a two-level match (outer, inner) of constructors
type Factory interface {
	// Register cons as a constructor for an element named inner, inside an element named outer
	Add(cons Constructor)
	AddNamed(cons Constructor, outer, inner xml.Name)

	// Construct an element named inner for an outer element named outer
	Get(outer, inner *xml.Name) Element
}

type Constructor func(*xml.Name) Element

type outerNodesFactory struct {
	reg map[xml.Name]*innerNodesFactory
	mx  sync.RWMutex
}

type innerNodesFactory struct {
	reg map[xml.Name]Constructor
	mx  sync.RWMutex
}

func NewFactory() *outerNodesFactory {
	return &outerNodesFactory{reg: make(map[xml.Name]*innerNodesFactory)}
}

func newINFactory() *innerNodesFactory {
	return &innerNodesFactory{reg: make(map[xml.Name]Constructor)}
}

func (r *outerNodesFactory) Add(cons Constructor) {
	outer, inner := getNames(cons(nil))
	r.AddNamed(cons, *outer, *inner)
}

func (r *outerNodesFactory) AddNamed(cons Constructor, outer, inner xml.Name) {
	r.mx.Lock() // Think about CAS here
	reg := r.lookup(&outer)
	if reg == nil {
		log.Printf("outerNodesFactory#Add: innerNodesFactory search miss for %#v (created)\n", outer)
		reg = newINFactory()
		r.reg[outer] = reg
	}
	r.mx.Unlock()

	log.Printf("outerNodesFactory#Add: %#v > %#v\n", outer, inner)

	reg.add(cons, &inner)
}

func (r *outerNodesFactory) Get(outer, inner *xml.Name) Element {
	r.mx.RLock()
	reg := r.lookup(outer)
	r.mx.RUnlock()

	if reg == nil {
		log.Printf("outerNodesFactory#Get: XML name search miss for %#v\n", outer)
		return nil
	}

	reg.mx.RLock()
	cons, ok := reg.reg[*inner]
	if !ok {
		log.Printf("outerNodesFactory#Get: qualified XML name search miss for %#v > %#v\n", outer, inner)
		inner_anyns := xml.Name{Local: inner.Local}
		cons, ok = reg.reg[inner_anyns]
	}
	reg.mx.RUnlock()
	if ok {
		obj := cons(inner)
		if innerEl, ok := obj.(Registrable); ok {
			innerEl.SetFactory(r)
		}
		return obj
	} else {
		log.Printf("outerNodesFactory#Get: XML name search miss for %#v > %#v\n", outer, inner)
	}

	return nil
}

func (r *outerNodesFactory) lookup(node *xml.Name) *innerNodesFactory {
	nsr, ok := r.reg[*node]
	if !ok {
		log.Printf("outerNodesFactory#lookup: qualified XML name search miss for %#v\n", node)
		node_anyns := xml.Name{Local: node.Local}
		nsr = r.reg[node_anyns]
	}
	return nsr
}

func (nsr *innerNodesFactory) add(cons Constructor, node *xml.Name) {
	nsr.mx.Lock()
	if _, ok := nsr.reg[*node]; !ok {
		nsr.reg[*node] = cons
	} else {
		panic("Node already registered " + node.Local)
	}
	nsr.mx.Unlock()
}

func getNames(o interface{}) (*xml.Name, *xml.Name) {
	var deref func(tof reflect.Type) reflect.Type
	deref = func(tof reflect.Type) reflect.Type {
		if tof.Kind() == reflect.Ptr {
			return deref(tof.Elem())
		}
		return tof
	}

	if field, ok := deref(reflect.TypeOf(o)).FieldByName("XMLName"); ok {
		var outer, inner xml.Name
		for tag, name := range map[string]*xml.Name{"xml": &inner, "parent": &outer} {
			xmltag := strings.Fields(field.Tag.Get(tag))

			switch len(xmltag) {
			case 0:
				panic("Tag " + tag + " should be defined for XMLName")
			case 1:
				name.Local = xmltag[0]
			case 2:
				name.Space = xmltag[0]
				name.Local = xmltag[1]
			}
		}

		return &outer, &inner
	}

	panic("XMLName field is missing")
}
