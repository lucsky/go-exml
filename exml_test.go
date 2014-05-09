package exml

import (
	"encoding/xml"
	"fmt"
	"strings"
	"testing"

	"gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { check.TestingT(t) }

type EXMLSuite struct{}

var _ = check.Suite(&EXMLSuite{})

const SIMPLE = `<?xml version="1.0"?>
<root attr1="root.attr1" attr2="root.attr2">
    <node attr1="node1.attr1" attr2="node1.attr2" />
    <node attr1="node2.attr1" attr2="node2.attr2" />
    <node attr1="node3.attr1" attr2="node3.attr2" />
    <node attr1="node4.attr1" attr2="node4.attr2">
        <subnode attr1="subnode.attr1" attr2="subnode.attr2" />
    </node>
</root>`

func (s *EXMLSuite) Test_Nested(c *check.C) {
	decoder := NewDecoder(strings.NewReader(SIMPLE))

	handler1WasCalled := false
	handler2WasCalled := false
	handler3WasCalled := false

	decoder.On("root", func(attrs Attrs, text CharData) {
		handler1WasCalled = true
		attr1, _ := attrs.Get("attr1")
		c.Assert(attr1, check.Equals, "root.attr1")
		attr2, _ := attrs.Get("attr2")
		c.Assert(attr2, check.Equals, "root.attr2")

		nodeNum := 0
		decoder.On("node", func(attrs Attrs, text CharData) {
			handler2WasCalled = true
			nodeNum = nodeNum + 1
			attr1, _ := attrs.Get("attr1")
			c.Assert(attr1, check.Equals, fmt.Sprintf("node%d.attr1", nodeNum))
			attr2, _ := attrs.Get("attr2")
			c.Assert(attr2, check.Equals, fmt.Sprintf("node%d.attr2", nodeNum))

			decoder.On("subnode", func(attrs Attrs, text CharData) {
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

	decoder.On("root", func(attrs Attrs, text CharData) {
		handler1WasCalled = true
		attr1, _ := attrs.Get("attr1")
		c.Assert(attr1, check.Equals, "root.attr1")
		attr2, _ := attrs.Get("attr2")
		c.Assert(attr2, check.Equals, "root.attr2")
	})

	nodeNum := 0
	decoder.On("root/node", func(attrs Attrs, text CharData) {
		handler2WasCalled = true
		nodeNum = nodeNum + 1
		attr1, _ := attrs.Get("attr1")
		c.Assert(attr1, check.Equals, fmt.Sprintf("node%d.attr1", nodeNum))
		attr2, _ := attrs.Get("attr2")
		c.Assert(attr2, check.Equals, fmt.Sprintf("node%d.attr2", nodeNum))
	})

	decoder.On("root/node/subnode", func(attrs Attrs, text CharData) {
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
	decoder.On("root/node", func(attrs Attrs, text CharData) {
		handler1WasCalled = true
		nodeNum = nodeNum + 1
		attr1, _ := attrs.Get("attr1")
		c.Assert(attr1, check.Equals, fmt.Sprintf("node%d.attr1", nodeNum))
		attr2, _ := attrs.Get("attr2")
		c.Assert(attr2, check.Equals, fmt.Sprintf("node%d.attr2", nodeNum))

		decoder.On("subnode", func(attrs Attrs, text CharData) {
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

	decoder.On("root", func(attrs Attrs, text CharData) {
		handler1WasCalled = true
		attr1, _ := attrs.Get("attr1")
		c.Assert(attr1, check.Equals, "root.attr1")
		attr2, _ := attrs.Get("attr2")
		c.Assert(attr2, check.Equals, "root.attr2")

		decoder.On("node/subnode", func(attrs Attrs, text CharData) {
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

func runTextTest1(c *check.C, data string, expectedFmt string) {
	decoder := NewDecoder(strings.NewReader(data))

	nodeNum := 0
	handlerWasCalled := []bool{false, false, false}

	decoder.On("root", func(attrs Attrs, text CharData) {
		decoder.On("node", func(attrs Attrs, text CharData) {
			handlerWasCalled[nodeNum] = true
			nodeNum = nodeNum + 1
			decoder.On("$text", func(attrs Attrs, text CharData) {
				c.Assert(len(attrs), check.Equals, 0)
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

	decoder.On("root/node", func(attrs Attrs, text CharData) {
		handlerWasCalled[nodeNum] = true
		nodeNum = nodeNum + 1
		decoder.On("$text", func(attrs Attrs, text CharData) {
			c.Assert(len(attrs), check.Equals, 0)
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

	decoder.On("root/node/$text", func(attrs Attrs, text CharData) {
		handlerWasCalled[nodeNum] = true
		nodeNum = nodeNum + 1
		c.Assert(len(attrs), check.Equals, 0)
		c.Assert(string(text), check.Equals, fmt.Sprintf(expectedFmt, nodeNum))
	})

	decoder.Run()

	c.Assert(handlerWasCalled[0], check.Equals, true)
	c.Assert(handlerWasCalled[1], check.Equals, true)
	c.Assert(handlerWasCalled[2], check.Equals, true)
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
	for i := 0; i < b.N; i++ {
		l := &TextList{}
		var _ = xml.Unmarshal([]byte(data), l)
	}
}

func Benchmark_DecodeSimple(b *testing.B) {
	reader := strings.NewReader(SIMPLE)
	decoder := NewDecoder(reader)

	for i := 0; i < b.N; i++ {
		var tree *SimpleTree

		decoder.On("root", func(attrs Attrs, text CharData) {
			tree = &SimpleTree{}
			tree.XMLName = xml.Name{"", "root"}
			tree.Attr1, _ = attrs.Get("attr1")
			tree.Attr2, _ = attrs.Get("attr2")

			decoder.On("node", func(attrs Attrs, text CharData) {
				node := &SimpleTreeNode{}
				node.Attr1, _ = attrs.Get("attr1")
				node.Attr2, _ = attrs.Get("attr2")
				tree.Nodes = append(tree.Nodes, node)

				decoder.On("subnode", func(attrs Attrs, text CharData) {
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
	reader := strings.NewReader(SIMPLE)
	decoder := NewDecoder(reader)

	for i := 0; i < b.N; i++ {
		l := &TextList{}
		decoder.On("root/node/$text", func(attrs Attrs, text CharData) {
			l.Texts = append(l.Texts, string(text))
		})

		decoder.Run()
		reader.Seek(0, 0)
	}
}
