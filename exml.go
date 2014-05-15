package exml

import (
	"bytes"
	"encoding/xml"
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
	text        []byte
}

type Decoder struct {
	decoder        *xml.Decoder
	topHandler     *handler
	currentHandler *handler
	errorCallback  ErrorCallback
}

func NewDecoder(r io.Reader) *Decoder {
	topHandler := &handler{callback: nil, subHandlers: nil, parent: nil}
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
				sub = &handler{callback: nil, subHandlers: nil, parent: h}
			}
		} else {
			sub = &handler{callback: callback, subHandlers: nil, parent: h}
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
			d.handleText()
			d.handleToken(t)
		case xml.CharData:
			d.currentHandler.text = append(d.currentHandler.text, t...)
		case xml.EndElement:
			d.handleText()
			if d.currentHandler != d.topHandler {
				d.currentHandler = d.currentHandler.parent
			}
		}
	}
}

func (d *Decoder) handleToken(t xml.StartElement) {
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
}

func (d *Decoder) handleText() {
	h := d.topHandler.subHandlers["$text"]
	if h == nil {
		h = d.currentHandler.subHandlers["$text"]
	}

	text := bytes.TrimSpace(d.currentHandler.text)
	if h != nil && h.callback != nil && len(text) > 0 {
		d.currentHandler.text = d.currentHandler.text[:0]
		h.callback.(func(CharData))(text)
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

func (a Attrs) Get(name string) (string, bool) {
	for _, attr := range a {
		if attr.Name.Local == name {
			return attr.Value, true
		}
	}

	return "", false
}

func (a Attrs) GetString(name string, fallback string) string {
	val, ok := a.Get(name)
	if !ok {
		return fallback
	}

	return val
}
