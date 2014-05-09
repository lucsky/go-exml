package exml

import (
	"encoding/xml"
	"fmt"
	"os"
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
	decoder.On("root", func(attrs Attrs, text xml.CharData) {
		handler1WasCalled = true
		attr1, _ := attrs.Get("attr1")
		c.Assert(attr1, check.Equals, "root.attr1")
		attr2, _ := attrs.Get("attr2")
		c.Assert(attr2, check.Equals, "root.attr2")

		nodeNum := 0
		decoder.On("node", func(attrs Attrs, text xml.CharData) {
			handler2WasCalled = true
			nodeNum = nodeNum + 1
			attr1, _ := attrs.Get("attr1")
			c.Assert(attr1, check.Equals, fmt.Sprintf("node%d.attr1", nodeNum))
			attr2, _ := attrs.Get("attr2")
			c.Assert(attr2, check.Equals, fmt.Sprintf("node%d.attr2", nodeNum))

			decoder.On("subnode", func(attrs Attrs, text xml.CharData) {
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

	decoder.On("root", func(attrs Attrs, text xml.CharData) {
		handler1WasCalled = true
		attr1, _ := attrs.Get("attr1")
		c.Assert(attr1, check.Equals, "root.attr1")
		attr2, _ := attrs.Get("attr2")
		c.Assert(attr2, check.Equals, "root.attr2")
	})

	nodeNum := 0
	decoder.On("root/node", func(attrs Attrs, text xml.CharData) {
		handler2WasCalled = true
		nodeNum = nodeNum + 1
		attr1, _ := attrs.Get("attr1")
		c.Assert(attr1, check.Equals, fmt.Sprintf("node%d.attr1", nodeNum))
		attr2, _ := attrs.Get("attr2")
		c.Assert(attr2, check.Equals, fmt.Sprintf("node%d.attr2", nodeNum))
	})

	decoder.On("root/node/subnode", func(attrs Attrs, text xml.CharData) {
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
	decoder.On("root/node", func(attrs Attrs, text xml.CharData) {
		handler1WasCalled = true
		nodeNum = nodeNum + 1
		attr1, _ := attrs.Get("attr1")
		c.Assert(attr1, check.Equals, fmt.Sprintf("node%d.attr1", nodeNum))
		attr2, _ := attrs.Get("attr2")
		c.Assert(attr2, check.Equals, fmt.Sprintf("node%d.attr2", nodeNum))

		decoder.On("subnode", func(attrs Attrs, text xml.CharData) {
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

	decoder.On("root", func(attrs Attrs, text xml.CharData) {
		handler1WasCalled = true
		attr1, _ := attrs.Get("attr1")
		c.Assert(attr1, check.Equals, "root.attr1")
		attr2, _ := attrs.Get("attr2")
		c.Assert(attr2, check.Equals, "root.attr2")

		decoder.On("node/subnode", func(attrs Attrs, text xml.CharData) {
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

	decoder.On("root", func(attrs Attrs, text xml.CharData) {
		decoder.On("node", func(attrs Attrs, text xml.CharData) {
			handlerWasCalled[nodeNum] = true
			nodeNum = nodeNum + 1
			decoder.On("$text", func(attrs Attrs, text xml.CharData) {
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

	decoder.On("root/node", func(attrs Attrs, text xml.CharData) {
		handlerWasCalled[nodeNum] = true
		nodeNum = nodeNum + 1
		decoder.On("$text", func(attrs Attrs, text xml.CharData) {
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

	decoder.On("root/node/$text", func(attrs Attrs, text xml.CharData) {
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
