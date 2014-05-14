package exml

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"strings"
)

type Callback interface{}
type ElemCallback func(Attrs)
type TextCallback func(CharData)
type ErrorCallback func(error)

type handler struct {
	callback    Callback
	subHandlers map[string]*handler
	parent      *handler
}

type Decoder struct {
	decoder        *xml.Decoder
	topHandler     *handler
	currentHandler *handler
	text           []byte
	errorCallback  ErrorCallback
}

func NewDecoder(r io.Reader) *Decoder {
	topHandler := &handler{nil, nil, nil}
	return &Decoder{
		decoder:        xml.NewDecoder(r),
		topHandler:     topHandler,
		currentHandler: topHandler,
	}
}

func (d *Decoder) On(event string, callback Callback) {
	events := strings.Split(event, "/")
	depth := len(events) - 1
	h := d.currentHandler

	for i, ev := range events {
		var sub *handler
		if i < depth {
			sub = h.subHandlers[ev]
			if sub == nil {
				sub = &handler{nil, nil, h}
			}
		} else {
			sub = &handler{callback, nil, h}
		}

		if h.subHandlers == nil {
			h.subHandlers = make(map[string]*handler)
		}

		h.subHandlers[ev] = sub
		h = sub
	}
}

func (d *Decoder) OnError(handler ErrorCallback) {
	d.errorCallback = handler
}

func (d *Decoder) Run() {
	for {
		token, err := d.decoder.Token()
		if token == nil {
			if d.errorCallback != nil {
				d.errorCallback(err)
			}
			break
		}

		switch t := token.(type) {
		case xml.StartElement:
			h := d.topHandler.subHandlers[t.Name.Local]
			if h == nil && d.currentHandler != d.topHandler {
				h = d.currentHandler.subHandlers[t.Name.Local]
			}
			if h != nil {
				h.parent = d.currentHandler
				d.currentHandler = h
				if h.callback != nil {
					h.callback.(func(Attrs))(t.Attr)
				}
			}
			break
		case xml.CharData:
			d.text = append(d.text, t...)
			break
		case xml.EndElement:
			text := bytes.TrimSpace(d.text)
			if len(text) > 0 {
				h := d.topHandler.subHandlers["$text"]
				if h == nil {
					h = d.currentHandler.subHandlers["$text"]
				}
				if h != nil && h.callback != nil {
					h.callback.(func(CharData))(text)
				}
			}
			if d.currentHandler != d.topHandler {
				d.currentHandler = d.currentHandler.parent
			}
			d.text = d.text[:0]
			break
		}
	}
}

func Assign(slot *string) func(CharData) {
	return func(c CharData) {
		*slot = string(c)
	}
}

func Append(a *[]string) func(CharData) {
	return func(c CharData) {
		*a = append(*a, string(c))
	}
}

type Attrs []xml.Attr
type CharData xml.CharData

var AttributeNotFoundError = errors.New("attribute not found")

func (a Attrs) Get(name string) (string, error) {
	for _, attr := range a {
		if attr.Name.Local == name {
			return attr.Value, nil
		}
	}

	return "", AttributeNotFoundError
}
