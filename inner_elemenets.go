package xtream

import "encoding/xml"

type InnerElements interface {
	Elements() []Element
	AddElement(Element)
	Registrable
}

type Registrable interface {
	SetFactory(Factory)
	SetXMLName(*xml.Name)
}

type Element interface{}

type InnerXML struct {
	XMLName xml.Name
	XML     string `xml:",innerxml"`
}

// Implementation of InnerElements interface, which (un)marshals it's elements recursively
type elements struct {
	// Name of the containing XML element, set in the constructor
	outer *xml.Name

	// A factory which can be used to unmarshal inner elements of this element
	reg Factory

	// Inner elements of this element
	elements []Element

	// Raw XML (including XMLName) of inner elements that may not be unmarshalled
	rawXML []*InnerXML
}

func NewElements() *elements {
	return &elements{
		elements: make([]Element, 0),
		rawXML:   make([]*InnerXML, 0),
	}
}

func (es *elements) SetXMLName(outer *xml.Name) {
	es.outer = outer
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

// Marshal all the elements and raw XML
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
