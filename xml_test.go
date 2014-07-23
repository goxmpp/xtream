package xtream_test

import (
	"encoding/xml"
	"testing"

	"github.com/goxmpp/xtream"
)

type stream struct {
	XMLName              xml.Name `xml:"stream"`
	xtream.InnerElements `xml:",any"`
}

type Message struct {
	XMLName              xml.Name `xml:"test://message message"`
	Id                   int      `xml:"id,attr"`
	Type                 string   `xml:"type,attr"`
	xtream.InnerElements `xml:",any"`
}

type Body struct {
	XMLName xml.Name `xml:"body" parent:"message"`
	Lang    string   `xml:"language,attr"`
	Text    string   `xml:",innerxml"`
}

func TestBasic(t *testing.T) {
	msgName, strmName := xml.Name{Local: "message"}, xml.Name{Local: "stream"}

	xtream.NodeFactory.AddNamed(func() xtream.Element {
		return &Message{InnerElements: xtream.NewElements(&msgName)}
	}, strmName, msgName)
	xtream.NodeFactory.Add(func() xtream.Element {
		return &Body{}
	})

	raw_xml := `<stream>
	<message xmlns="test://message" id="10" type="plain">
		<body language="en">some strage <br/>xml</body>
	</message>
</stream>`

	s := &stream{InnerElements: xtream.NewElements(&strmName)}
	err := xml.Unmarshal([]byte(raw_xml), s)
	if err != nil {
		t.Fatal(err)
	}

	assert(t, s.XMLName.Local == "stream", "Wrong top level node unmarshaled")

	els := s.Elements()
	assert(t, len(els) == 1, "Wrong number of elements unmarshaled at first level")

	msg, ok := els[0].(*Message)
	assert(t, ok, "First inner node should be a 'message'")
	assert(t, msg.XMLName.Local == "message", "Tag name 'message' unmarshaled correctly")
	assert(t, msg.XMLName.Space == "test://message", "Tag name was unmarshaled correctly")
	assert(t, msg.Id == 10, "Message Id should be 10")
	assert(t, msg.Type == "plain", "Message type should be 'plain'")

	msg_els := msg.Elements()

	assert(t, len(msg_els) == 1, "Message should contain only one inner element")

	if body, ok := msg_els[0].(*Body); ok {
		assert(t, body.XMLName.Local == "body", "Tag name 'body' unmarshaled correctly")
		assert(t, body.XMLName.Space == "test://message", "Tag name was unmarshaled correctly")
		assert(t, body.Lang == "en", "Language tag should be unmarshaled correctly")
		assert(t, body.Text == "some strage <br/>xml", "Inner XML text should be valid")
	} else {
		t.Fatal("Messages inner element should be 'body'")
	}

	got, err := xml.MarshalIndent(s, "", "\t")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Got:    %s", got)
	t.Logf("Extect: %s", raw_xml)

	assert(t, string(got) == raw_xml, "Original and Marshaled XML don't match")
}

func assert(t *testing.T, ok bool, msg string) {
	if !ok {
		t.Fatal(msg)
	}
}
