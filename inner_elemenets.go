package xtream

import "encoding/xml"

type InnerElements interface {
	Elements() []Element
	AddElement(Element)
}

type Registrable interface {
	SetFactory(Factory)
}

type Element interface{}

type InnerXML struct {
	XMLName xml.Name
	XML     string `xml:",innerxml"`
}

type elements struct {
	outer    *xml.Name
	reg      Factory
	elements []Element
	rawXML   []*InnerXML
}

func NewElements(outer *xml.Name) *elements {
	return &elements{
		outer:    outer,
		elements: make([]Element, 0),
		rawXML:   make([]*InnerXML, 0),
	}
}

func (es *elements) SetFactory(reg Factory) {
	es.reg = reg
}

func (es *elements) AddElement(e Element) {
	es.elements = append(es.elements, e)
}

func (es *elements) Elements() []Element {
	return es.elements
}

func (es *elements) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	elementObject, err := es.decodeElement(d, &start)
	if err != nil {
		return err
	}

	if innerXML, ok := elementObject.(*InnerXML); ok {
		es.rawXML = append(es.rawXML, innerXML)
	} else {
		es.AddElement(elementObject)
	}

	return nil
}

func (es *elements) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if err := e.Encode(es.elements); err != nil {
		return err
	}

	return e.Encode(es.rawXML)
}

func (es *elements) decodeElement(d *xml.Decoder, start *xml.StartElement) (interface{}, error) {
	if es.reg == nil {
		es.reg = NodeFactory
	}

	if es.outer == nil {
		es.outer = &xml.Name{}
	}

	element := es.reg.Get(es.outer, &start.Name)
	if element == nil {
		element = &InnerXML{}
	}

	if err := d.DecodeElement(element, start); err != nil {
		return nil, err
	}
	return element, nil
}
