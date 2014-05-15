package exml

import (
	"encoding/xml"
	"fmt"
	"strings"
	"testing"

	"gopkg.in/check.v1"
)

// ============================================================================
// Example

const EXAMPLE = `<?xml version="1.0"?>
<address-book name="homies">
    <contact>
        <first-name>Tim</first-name>
        <last-name>Cook</last-name>
        <address>Cupertino</address>
    </contact>
    <contact>
        <first-name>Steve</first-name>
        <last-name>Ballmer</last-name>
        <address>Redmond</address>
    </contact>
    <contact>
        <first-name>Mark</first-name>
        <last-name>Zuckerberg</last-name>
        <address>Menlo Park</address>
    </contact>
</address-book>`

type AddressBook struct {
	Name     string
	Contacts []*Contact
}

type Contact struct {
	FirstName string
	LastName  string
	Address   string
}

func Example() {
	reader := strings.NewReader(EXAMPLE)
	decoder := NewDecoder(reader)

	addressBook := AddressBook{}

	decoder.On("address-book", func(attrs Attrs) {
		addressBook.Name, _ = attrs.Get("name")

		decoder.On("contact", func(attrs Attrs) {
			contact := &Contact{}
			addressBook.Contacts = append(addressBook.Contacts, contact)

			decoder.On("first-name/$text", func(text CharData) {
				contact.FirstName = string(text)
			})

			decoder.On("last-name/$text", func(text CharData) {
				contact.LastName = string(text)
			})

			decoder.On("address/$text", func(text CharData) {
				contact.Address = string(text)
			})
		})
	})

	decoder.Run()

	fmt.Printf("Address book: %s\n", addressBook.Name)
	for _, c := range addressBook.Contacts {
		fmt.Println("-", c.FirstName, c.LastName, "@", c.Address)
	}

	// Output:
	// Address book: homies
	// - Tim Cook @ Cupertino
	// - Steve Ballmer @ Redmond
	// - Mark Zuckerberg @ Menlo Park
}

// ============================================================================
// Tests

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { check.TestingT(t) }

type EXMLSuite struct{}

var _ = check.Suite(&EXMLSuite{})

const ATTRIBUTE = `<?xml version="1.0"?><node attr="node.attr" />`

func (s *EXMLSuite) Test_AttributeReaders(c *check.C) {
	decoder := NewDecoder(strings.NewReader(ATTRIBUTE))
	handlerWasCalled := false

	decoder.On("node", func(attrs Attrs) {
		handlerWasCalled = true

		var attr string
		var ok bool

		attr, ok = attrs.Get("attr")
		c.Assert(attr, check.Equals, "node.attr")
		c.Assert(ok, check.Equals, true)

		attr, ok = attrs.Get("omfglol")
		c.Assert(attr, check.Equals, "")
		c.Assert(ok, check.Equals, false)

		attr = attrs.GetString("attr", "default")
		c.Assert(attr, check.Equals, "node.attr")

		attr = attrs.GetString("omfglol", "default")
		c.Assert(attr, check.Equals, "default")
	})

	decoder.Run()
	c.Assert(handlerWasCalled, check.Equals, true)
}

const SIMPLE = `<?xml version="1.0"?>
<root attr1="root.attr1" attr2="root.attr2">
    <node attr1="node1.attr1" attr2="node1.attr2" />
    <node attr1="node2.attr1" attr2="node2.attr2" />
    <node attr1="node3.attr1" attr2="node3.attr2" />
    <node attr1="node4.attr1" attr2="node4.attr2">
        <subnode attr1="subnode.attr1" attr2="subnode.attr2" />
    </node>
</root>`

func (s *EXMLSuite) Test_Empty(c *check.C) {
	decoder := NewDecoder(strings.NewReader(SIMPLE))
	decoder.Run()
	c.Succeed()
}

func (s *EXMLSuite) Test_Nested(c *check.C) {
	decoder := NewDecoder(strings.NewReader(SIMPLE))

	handler1WasCalled := false
	handler2WasCalled := false
	handler3WasCalled := false

	decoder.On("root", func(attrs Attrs) {
		handler1WasCalled = true
		attr1, _ := attrs.Get("attr1")
		c.Assert(attr1, check.Equals, "root.attr1")
		attr2, _ := attrs.Get("attr2")
		c.Assert(attr2, check.Equals, "root.attr2")

		nodeNum := 0
		decoder.On("node", func(attrs Attrs) {
			handler2WasCalled = true
			nodeNum = nodeNum + 1
			attr1, _ := attrs.Get("attr1")
			c.Assert(attr1, check.Equals, fmt.Sprintf("node%d.attr1", nodeNum))
			attr2, _ := attrs.Get("attr2")
			c.Assert(attr2, check.Equals, fmt.Sprintf("node%d.attr2", nodeNum))

			decoder.On("subnode", func(attrs Attrs) {
				handler3WasCalled = true
				attr1, _ := attrs.Get("attr1")
				c.Assert(attr1, check.Equals, "subnode.attr1")
				attr2, _ := attrs.Get("attr2")
				c.Assert(attr2, check.Equals, "subnode.attr2")
			})
		})
	})

	decoder.Run()

	c.Assert(handler1WasCalled, check.Equals, true)
	c.Assert(handler2WasCalled, check.Equals, true)
	c.Assert(handler3WasCalled, check.Equals, true)
}

func (s *EXMLSuite) Test_Flat(c *check.C) {
	decoder := NewDecoder(strings.NewReader(SIMPLE))

	handler1WasCalled := false
	handler2WasCalled := false
	handler3WasCalled := false

	decoder.On("root", func(attrs Attrs) {
		handler1WasCalled = true
		attr1, _ := attrs.Get("attr1")
		c.Assert(attr1, check.Equals, "root.attr1")
		attr2, _ := attrs.Get("attr2")
		c.Assert(attr2, check.Equals, "root.attr2")
	})

	nodeNum := 0
	decoder.On("root/node", func(attrs Attrs) {
		handler2WasCalled = true
		nodeNum = nodeNum + 1
		attr1, _ := attrs.Get("attr1")
		c.Assert(attr1, check.Equals, fmt.Sprintf("node%d.attr1", nodeNum))
		attr2, _ := attrs.Get("attr2")
		c.Assert(attr2, check.Equals, fmt.Sprintf("node%d.attr2", nodeNum))
	})

	decoder.On("root/node/subnode", func(attrs Attrs) {
		handler3WasCalled = true
		attr1, _ := attrs.Get("attr1")
		c.Assert(attr1, check.Equals, "subnode.attr1")
		attr2, _ := attrs.Get("attr2")
		c.Assert(attr2, check.Equals, "subnode.attr2")
	})

	decoder.Run()

	c.Assert(nodeNum, check.Equals, 4)
	c.Assert(handler1WasCalled, check.Equals, true)
	c.Assert(handler2WasCalled, check.Equals, true)
	c.Assert(handler3WasCalled, check.Equals, true)
}

func (s *EXMLSuite) Test_Mixed1(c *check.C) {
	decoder := NewDecoder(strings.NewReader(SIMPLE))

	handler1WasCalled := false
	handler2WasCalled := false

	nodeNum := 0
	decoder.On("root/node", func(attrs Attrs) {
		handler1WasCalled = true
		nodeNum = nodeNum + 1
		attr1, _ := attrs.Get("attr1")
		c.Assert(attr1, check.Equals, fmt.Sprintf("node%d.attr1", nodeNum))
		attr2, _ := attrs.Get("attr2")
		c.Assert(attr2, check.Equals, fmt.Sprintf("node%d.attr2", nodeNum))

		decoder.On("subnode", func(attrs Attrs) {
			handler2WasCalled = true
			attr1, _ := attrs.Get("attr1")
			c.Assert(attr1, check.Equals, "subnode.attr1")
			attr2, _ := attrs.Get("attr2")
			c.Assert(attr2, check.Equals, "subnode.attr2")
		})
	})

	decoder.Run()

	c.Assert(nodeNum, check.Equals, 4)
	c.Assert(handler1WasCalled, check.Equals, true)
	c.Assert(handler2WasCalled, check.Equals, true)
}

func (s *EXMLSuite) Test_Mixed2(c *check.C) {
	decoder := NewDecoder(strings.NewReader(SIMPLE))

	handler1WasCalled := false
	handler2WasCalled := false

	decoder.On("root", func(attrs Attrs) {
		handler1WasCalled = true
		attr1, _ := attrs.Get("attr1")
		c.Assert(attr1, check.Equals, "root.attr1")
		attr2, _ := attrs.Get("attr2")
		c.Assert(attr2, check.Equals, "root.attr2")

		decoder.On("node/subnode", func(attrs Attrs) {
			handler2WasCalled = true
			attr1, _ := attrs.Get("attr1")
			c.Assert(attr1, check.Equals, "subnode.attr1")
			attr2, _ := attrs.Get("attr2")
			c.Assert(attr2, check.Equals, "subnode.attr2")
		})
	})

	decoder.Run()

	c.Assert(handler1WasCalled, check.Equals, true)
	c.Assert(handler2WasCalled, check.Equals, true)
}

func (s *EXMLSuite) Test_Global(c *check.C) {
	decoder := NewDecoder(strings.NewReader(SIMPLE))

	handlerWasCalled := false
	decoder.On("subnode", func(attrs Attrs) {
		handlerWasCalled = true
		attr1, _ := attrs.Get("attr1")
		c.Assert(attr1, check.Equals, "subnode.attr1")
		attr2, _ := attrs.Get("attr2")
		c.Assert(attr2, check.Equals, "subnode.attr2")
	})

	decoder.Run()

	c.Assert(handlerWasCalled, check.Equals, true)
}

const TEXT = `<?xml version="1.0"?>
<root>
    <node>text content 1</node>
    <node>text content 2</node>
    <node>text content 3</node>
</root>`

func (s *EXMLSuite) Test_Text1(c *check.C) {
	runTextTest1(c, TEXT, "text content %d")
}

func (s *EXMLSuite) Test_Text2(c *check.C) {
	runTextTest2(c, TEXT, "text content %d")
}

func (s *EXMLSuite) Test_Text3(c *check.C) {
	runTextTest3(c, TEXT, "text content %d")
}

func (s *EXMLSuite) Test_Text4(c *check.C) {
	runTextTest4(c, TEXT, "text content %d")
}

const CDATA = `<?xml version="1.0"?>
<root>
    <node><![CDATA[CDATA content 1]]></node>
    <node><![CDATA[CDATA content 2]]></node>
    <node><![CDATA[CDATA content 3]]></node>
</root>`

func (s *EXMLSuite) Test_CDATA1(c *check.C) {
	runTextTest1(c, CDATA, "CDATA content %d")
}

func (s *EXMLSuite) Test_CDATA2(c *check.C) {
	runTextTest2(c, CDATA, "CDATA content %d")
}

func (s *EXMLSuite) Test_CDATA3(c *check.C) {
	runTextTest3(c, CDATA, "CDATA content %d")
}

func (s *EXMLSuite) Test_CDATA4(c *check.C) {
	runTextTest4(c, CDATA, "CDATA content %d")
}

const MIXED = `<?xml version="1.0"?>
<root>
    <node>Text content followed by some <![CDATA[CDATA content 1]]></node>
    <node>Text content followed by some <![CDATA[CDATA content 2]]></node>
    <node>Text content followed by some <![CDATA[CDATA content 3]]></node>
</root>`

func (s *EXMLSuite) Test_MixedContent1(c *check.C) {
	runTextTest1(c, MIXED, "Text content followed by some CDATA content %d")
}

func (s *EXMLSuite) Test_MixedContent2(c *check.C) {
	runTextTest2(c, MIXED, "Text content followed by some CDATA content %d")
}

func (s *EXMLSuite) Test_MixedContent3(c *check.C) {
	runTextTest3(c, MIXED, "Text content followed by some CDATA content %d")
}

func (s *EXMLSuite) Test_MixedContent4(c *check.C) {
	runTextTest4(c, MIXED, "Text content followed by some CDATA content %d")
}

func runTextTest1(c *check.C, data string, expectedFmt string) {
	decoder := NewDecoder(strings.NewReader(data))

	nodeNum := 0
	handlerWasCalled := []bool{false, false, false}

	decoder.On("root", func(attrs Attrs) {
		decoder.On("node", func(attrs Attrs) {
			handlerWasCalled[nodeNum] = true
			nodeNum = nodeNum + 1
			decoder.On("$text", func(text CharData) {
				c.Assert(string(text), check.Equals, fmt.Sprintf(expectedFmt, nodeNum))
			})
		})
	})

	decoder.Run()

	c.Assert(handlerWasCalled[0], check.Equals, true)
	c.Assert(handlerWasCalled[1], check.Equals, true)
	c.Assert(handlerWasCalled[2], check.Equals, true)
}

func runTextTest2(c *check.C, data string, expectedFmt string) {
	decoder := NewDecoder(strings.NewReader(data))

	nodeNum := 0
	handlerWasCalled := []bool{false, false, false}

	decoder.On("root/node", func(attrs Attrs) {
		handlerWasCalled[nodeNum] = true
		nodeNum = nodeNum + 1
		decoder.On("$text", func(text CharData) {
			c.Assert(string(text), check.Equals, fmt.Sprintf(expectedFmt, nodeNum))
		})
	})

	decoder.Run()

	c.Assert(handlerWasCalled[0], check.Equals, true)
	c.Assert(handlerWasCalled[1], check.Equals, true)
	c.Assert(handlerWasCalled[2], check.Equals, true)

}

func runTextTest3(c *check.C, data string, expectedFmt string) {
	decoder := NewDecoder(strings.NewReader(data))

	nodeNum := 0
	handlerWasCalled := []bool{false, false, false}

	decoder.On("root/node/$text", func(text CharData) {
		handlerWasCalled[nodeNum] = true
		nodeNum = nodeNum + 1
		c.Assert(string(text), check.Equals, fmt.Sprintf(expectedFmt, nodeNum))
	})

	decoder.Run()

	c.Assert(handlerWasCalled[0], check.Equals, true)
	c.Assert(handlerWasCalled[1], check.Equals, true)
	c.Assert(handlerWasCalled[2], check.Equals, true)
}

func runTextTest4(c *check.C, data string, expectedFmt string) {
	decoder := NewDecoder(strings.NewReader(data))
	nodeNum := 0

	decoder.On("$text", func(text CharData) {
		nodeNum = nodeNum + 1
		if nodeNum <= 3 {
			c.Assert(string(text), check.Equals, fmt.Sprintf(expectedFmt, nodeNum))
		}
	})

	decoder.Run()
}

const ASSIGN = `<?xml version="1.0"?>
<root>
    <node>Text content</node>
</root>`

func (s *EXMLSuite) Test_Assign(c *check.C) {
	var text string

	decoder := NewDecoder(strings.NewReader(ASSIGN))
	decoder.On("root/node/$text", Assign(&text))
	decoder.Run()

	c.Assert(text, check.Equals, "Text content")
}

const APPEND = `<?xml version="1.0"?>
<root>
    <node>Text content 1</node>
    <node>Text content 2</node>
    <node>Text content 3</node>
</root>`

func (s *EXMLSuite) Test_Append(c *check.C) {
	texts := []string{}

	decoder := NewDecoder(strings.NewReader(APPEND))
	decoder.On("root/node/$text", Append(&texts))
	decoder.Run()

	c.Assert(texts[0], check.Equals, "Text content 1")
	c.Assert(texts[1], check.Equals, "Text content 2")
	c.Assert(texts[2], check.Equals, "Text content 3")
}

const NESTED_TEXT = `<?xml version="1.0"?>
<root>Root text 1<node>Node text</node>Root text 2</root>`

func (s *EXMLSuite) Test_NestedText(c *check.C) {
	texts := []string{}

	decoder := NewDecoder(strings.NewReader(NESTED_TEXT))
	decoder.On("$text", Append(&texts))
	decoder.Run()

	c.Assert(texts[0], check.Equals, "Root text 1")
	c.Assert(texts[1], check.Equals, "Node text")
	c.Assert(texts[2], check.Equals, "Root text 2")
}

const MALFORMED = "<?xml version=\"1.0\"?><root></node>"

func (s *EXMLSuite) Test_Error(c *check.C) {
	decoder := NewDecoder(strings.NewReader(MALFORMED))

	handlerWasCalled := false

	decoder.OnError(func(err error) {
		handlerWasCalled = true
	})

	decoder.Run()

	c.Assert(handlerWasCalled, check.Equals, true)
}

// ============================================================================
// Benchmarks

type SimpleTreeNode struct {
	Attr1 string          `xml:"attr1,attr"`
	Attr2 string          `xml:"attr2,attr"`
	Sub   *SimpleTreeNode `xml:"subnode"`
}

type SimpleTree struct {
	XMLName xml.Name          `xml:"root"`
	Attr1   string            `xml:"attr1,attr"`
	Attr2   string            `xml:"attr2,attr"`
	Nodes   []*SimpleTreeNode `xml:"node"`
}

func Benchmark_UnmarshalSimple(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tree := &SimpleTree{}
		var _ = xml.Unmarshal([]byte(SIMPLE), tree)
	}
}

func Benchmark_UnmarshalText(b *testing.B) {
	runUnmarshalTextBenchmark(b, TEXT)
}

func Benchmark_UnmarshalCDATA(b *testing.B) {
	runUnmarshalTextBenchmark(b, CDATA)
}

func Benchmark_UnmarshalMixed(b *testing.B) {
	runUnmarshalTextBenchmark(b, MIXED)
}

type TextList struct {
	Texts []string `xml:"node"`
}

func runUnmarshalTextBenchmark(b *testing.B, data string) {
	l := &TextList{}
	for i := 0; i < b.N; i++ {
		var _ = xml.Unmarshal([]byte(data), l)
		l.Texts = l.Texts[:0]
	}
}

func Benchmark_DecodeSimple(b *testing.B) {
	reader := strings.NewReader(SIMPLE)
	decoder := NewDecoder(reader)

	for i := 0; i < b.N; i++ {
		decoder.On("root", func(attrs Attrs) {
			tree := &SimpleTree{}
			tree.XMLName = xml.Name{Space: "", Local: "root"}
			tree.Attr1, _ = attrs.Get("attr1")
			tree.Attr2, _ = attrs.Get("attr2")

			decoder.On("node", func(attrs Attrs) {
				node := &SimpleTreeNode{}
				node.Attr1, _ = attrs.Get("attr1")
				node.Attr2, _ = attrs.Get("attr2")
				tree.Nodes = append(tree.Nodes, node)

				decoder.On("subnode", func(attrs Attrs) {
					node.Sub = &SimpleTreeNode{}
					node.Sub.Attr1, _ = attrs.Get("attr1")
					node.Sub.Attr2, _ = attrs.Get("attr2")
				})
			})
		})

		decoder.Run()
		reader.Seek(0, 0)
	}
}

func Benchmark_DecodeText(b *testing.B) {
	runDecodeTextBenchmark(b, TEXT)
}

func Benchmark_DecodeCDATA(b *testing.B) {
	runDecodeTextBenchmark(b, CDATA)
}

func Benchmark_DecodeMixed(b *testing.B) {
	runDecodeTextBenchmark(b, MIXED)
}

func runDecodeTextBenchmark(b *testing.B, data string) {
	reader := strings.NewReader(data)
	decoder := NewDecoder(reader)
	l := &TextList{}

	for i := 0; i < b.N; i++ {
		decoder.On("root/node/$text", func(text CharData) {
			l.Texts = append(l.Texts, string(text))
		})

		decoder.Run()

		reader.Seek(0, 0)
		l.Texts = l.Texts[:0]
	}
}
