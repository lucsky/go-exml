package exml

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"strings"
)

type Handler interface{}
type ElemHandler func(Attrs)
type TextHandler func(CharData)
type ErrorHandler func(error)

type Decoder struct {
	decoder      *xml.Decoder
	handlers     map[string]Handler
	errorHandler ErrorHandler
	events       []string
	text         *bytes.Buffer
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		decoder:  xml.NewDecoder(r),
		handlers: make(map[string]Handler),
		events:   []string{"/"},
		text:     new(bytes.Buffer),
	}
}

func (d *Decoder) On(event string, handler Handler) {
	fullEvent := strings.Join(append(d.events, event), "/")
	d.handlers[fullEvent] = handler
}

func (d *Decoder) OnError(handler ErrorHandler) {
	d.errorHandler = handler
}

func (d *Decoder) Run() {
	for {
		token, err := d.decoder.Token()
		if token == nil {
			if d.errorHandler != nil {
				d.errorHandler(err)
			}
			break
		}

		switch t := token.(type) {
		case xml.StartElement:
			d.text.Reset()
			d.events = append(d.events, t.Name.Local)
			handler := d.getHandler()
			if handler != nil {
				handler.(func(Attrs))(t.Attr)
			}
			break
		case xml.CharData:
			d.text.Write(t)
			break
		case xml.EndElement:
			numPop := 1
			if d.text.Len() > 0 {
				numPop = 2
				d.events = append(d.events, "$text")
				handler := d.getHandler()
				if handler != nil {
					handler.(func(CharData))(d.text.Bytes())
				}
			}
			d.events = d.events[:len(d.events)-numPop]
			break
		}
	}
}

func (d *Decoder) getHandler() Handler {
	fullEvent := strings.Join(d.events, "/")
	return d.handlers[fullEvent]
}

type Attrs []xml.Attr
type CharData xml.CharData

func (a Attrs) Get(name string) (string, error) {
	for _, attr := range a {
		if attr.Name.Local == name {
			return attr.Value, nil
		}
	}

	return "", errors.New("attribute not found")
}
