package xtream_test

import (
	"encoding/xml"
	"testing"

	"github.com/azhavnerchik/xtream"
)

type stream struct {
	XMLName              xml.Name `xml:"stream"`
	xtream.InnerElements `xml:",any"`
}

type Message struct {
	XMLName              xml.Name `xml:"message"`
	Id                   string   `xml:"id,attr"`
	Type                 string   `xml:"type,attr"`
	xtream.InnerElements `xml:",any"`
}

type Body struct {
	XMLName xml.Name `xml:"body"`
	Lang    string   `xml:"language,attr"`
	Text    string   `xml:",innerxml"`
}

func TestBasic(t *testing.T) {
	msgName, strmName := xml.Name{Local: "message"}, xml.Name{Local: "stream"}

	xtream.NodeRegistry.Add(func() interface{} {
		return &Message{InnerElements: xtream.NewElemenets(&msgName)}
	}, strmName, msgName)
	xtream.NodeRegistry.Add(func() interface{} {
		return &Body{}
	}, msgName, xml.Name{Local: "body"})

	raw_xml := `
<stream>
	<message id="10" type="plain" xmlns="asdasd">
		<body language="en">some strage <br/>xml</body>
	</message>
</stream>
`
	s := &stream{InnerElements: xtream.NewElemenets(&strmName)}
	err := xml.Unmarshal([]byte(raw_xml), s)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%#v", s)
	t.Logf("%#v\n", s.Elements())
	for _, e := range s.Elements() {
		t.Logf("%#v\n", e)
		for _, eb := range e.(xtream.InnerElements).Elements() {
			t.Logf("%#v\n", eb)
		}
	}
}
