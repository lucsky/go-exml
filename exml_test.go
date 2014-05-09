package exml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { check.TestingT(t) }

type EXMLSuite struct{}

var _ = check.Suite(&EXMLSuite{})

func (s *EXMLSuite) Test_Nested(c *check.C) {
	reader, err := os.Open("test_files/simple.xml")
	if err != nil {
		c.Error(err)
		c.FailNow()
	}
	defer reader.Close()

	handler1WasCalled := false
	handler2WasCalled := false
	handler3WasCalled := false

	decoder := NewDecoder(reader)
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
	reader, err := os.Open("test_files/simple.xml")
	if err != nil {
		c.Error(err)
		c.FailNow()
	}
	defer reader.Close()

	handler1WasCalled := false
	handler2WasCalled := false
	handler3WasCalled := false

	decoder := NewDecoder(reader)

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
	reader, err := os.Open("test_files/simple.xml")
	if err != nil {
		c.Error(err)
		c.FailNow()
	}
	defer reader.Close()

	handler1WasCalled := false
	handler2WasCalled := false

	decoder := NewDecoder(reader)

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
	reader, err := os.Open("test_files/simple.xml")
	if err != nil {
		c.Error(err)
		c.FailNow()
	}
	defer reader.Close()

	handler1WasCalled := false
	handler2WasCalled := false

	decoder := NewDecoder(reader)

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

func (s *EXMLSuite) Test_Text1(c *check.C) {
	runTextTest1(c, "test_files/text.xml", "text content %d")
}

func (s *EXMLSuite) Test_Text2(c *check.C) {
	runTextTest2(c, "test_files/text.xml", "text content %d")
}

func (s *EXMLSuite) Test_Text3(c *check.C) {
	runTextTest3(c, "test_files/text.xml", "text content %d")
}

func (s *EXMLSuite) Test_CDATA1(c *check.C) {
	runTextTest1(c, "test_files/cdata.xml", "CDATA content %d")
}

func (s *EXMLSuite) Test_CDATA2(c *check.C) {
	runTextTest2(c, "test_files/cdata.xml", "CDATA content %d")
}

func (s *EXMLSuite) Test_CDATA3(c *check.C) {
	runTextTest3(c, "test_files/cdata.xml", "CDATA content %d")
}

func (s *EXMLSuite) Test_MixedContent1(c *check.C) {
	runTextTest1(c, "test_files/mixed.xml", "Text content followed by some CDATA content %d")
}

func (s *EXMLSuite) Test_MixedContent2(c *check.C) {
	runTextTest2(c, "test_files/mixed.xml", "Text content followed by some CDATA content %d")
}

func (s *EXMLSuite) Test_MixedContent3(c *check.C) {
	runTextTest3(c, "test_files/mixed.xml", "Text content followed by some CDATA content %d")
}

func runTextTest1(c *check.C, testFile string, expectedFmt string) {
	reader, err := os.Open(testFile)
	if err != nil {
		c.Error(err)
		c.FailNow()
	}
	defer reader.Close()

	decoder := NewDecoder(reader)
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

func runTextTest2(c *check.C, testFile string, expectedFmt string) {
	reader, err := os.Open(testFile)
	if err != nil {
		c.Error(err)
		c.FailNow()
	}
	defer reader.Close()

	decoder := NewDecoder(reader)
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

func runTextTest3(c *check.C, testFile string, expectedFmt string) {
	reader, err := os.Open(testFile)
	if err != nil {
		c.Error(err)
		c.FailNow()
	}
	defer reader.Close()

	decoder := NewDecoder(reader)
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

func (s *EXMLSuite) Test_Error(c *check.C) {
	reader := strings.NewReader("<?xml version=\"1.0\"?><root></node>")
	decoder := NewDecoder(reader)
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
	data, _ := ioutil.ReadFile("test_files/simple.xml")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tree := &SimpleTree{}
		var _ = xml.Unmarshal(data, tree)
	}
}

func Benchmark_UnmarshalText(b *testing.B) {
	runUnmarshalTextBenchmark(b, "test_files/text.xml")
}

func Benchmark_UnmarshalCDATA(b *testing.B) {
	runUnmarshalTextBenchmark(b, "test_files/cdata.xml")
}

func Benchmark_UnmarshalMixed(b *testing.B) {
	runUnmarshalTextBenchmark(b, "test_files/mixed.xml")
}

type TextList struct {
	Texts []string `xml:"node"`
}

func runUnmarshalTextBenchmark(b *testing.B, filename string) {
	data, _ := ioutil.ReadFile(filename)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l := &TextList{}
		var _ = xml.Unmarshal(data, l)
	}
}

func Benchmark_DecodeSimple(b *testing.B) {
	data, _ := ioutil.ReadFile("test_files/simple.xml")
	reader := bytes.NewReader(data)
	decoder := NewDecoder(reader)
	b.ResetTimer()

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
	runDecodeTextBenchmark(b, "test_files/text.xml")
}

func Benchmark_DecodeCDATA(b *testing.B) {
	runDecodeTextBenchmark(b, "test_files/cdata.xml")
}

func Benchmark_DecodeMixed(b *testing.B) {
	runDecodeTextBenchmark(b, "test_files/mixed.xml")
}

func runDecodeTextBenchmark(b *testing.B, filename string) {
	data, _ := ioutil.ReadFile(filename)
	reader := bytes.NewReader(data)
	decoder := NewDecoder(reader)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l := &TextList{}
		decoder.On("root/node/$text", func(attrs Attrs, text CharData) {
			l.Texts = append(l.Texts, string(text))
		})

		decoder.Run()
		reader.Seek(0, 0)
	}
}
