package exml

import (
	"bytes"
	"encoding/xml"
	"io"
	"strconv"
	"strings"
)

type TagCallback func(Attrs)
type TextCallback func(CharData)
type ErrorCallback func(error)

type handler struct {
	tagCallback   TagCallback
	textCallback  TextCallback
	subHandlers   map[string]*handler
	parentHandler *handler
	text          []byte
}

type Decoder struct {
	decoder        *xml.Decoder
	topHandler     *handler
	currentHandler *handler
	errorCallback  ErrorCallback
}

func NewDecoder(r io.Reader) *Decoder {
	topHandler := &handler{}
	return &Decoder{
		decoder:        xml.NewDecoder(r),
		topHandler:     topHandler,
		currentHandler: topHandler,
	}
}

func (d *Decoder) On(path string, callback TagCallback) {
	h := d.installHandlers(path)
	h.tagCallback = callback
}

func (d *Decoder) OnTextOf(path string, callback TextCallback) {
	h := d.installHandlers(path)
	h.textCallback = callback
}

func (d *Decoder) OnText(callback TextCallback) {
	d.currentHandler.textCallback = callback
}

func (d *Decoder) installHandlers(path string) *handler {
	events := strings.Split(path, "/")
	depth := len(events) - 1
	h := d.currentHandler

	var sub *handler
	for i, ev := range events {
		if i < depth {
			sub = h.subHandlers[ev]
			if sub == nil {
				sub = &handler{parentHandler: h}
			}
		} else {
			sub = &handler{parentHandler: h}
		}

		if h.subHandlers == nil {
			h.subHandlers = make(map[string]*handler)
		}

		h.subHandlers[ev] = sub
		h = sub
	}

	return sub
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
			d.handleTag(t)
		case xml.CharData:
			d.currentHandler.text = append(d.currentHandler.text, t...)
		case xml.EndElement:
			d.handleText()
			if d.currentHandler != d.topHandler {
				d.currentHandler = d.currentHandler.parentHandler
			}
		}
	}
}

func (d *Decoder) handleTag(t xml.StartElement) {
	h := d.topHandler.subHandlers[t.Name.Local]
	if h == nil && d.currentHandler != d.topHandler {
		h = d.currentHandler.subHandlers[t.Name.Local]
	}

	if h != nil {
		h.parentHandler = d.currentHandler
		d.currentHandler = h
		if h.tagCallback != nil {
			h.tagCallback(t.Attr)
		}
	}
}

func (d *Decoder) handleText() {
	text := bytes.TrimSpace(d.currentHandler.text)
	d.currentHandler.text = d.currentHandler.text[:0]
	if d.currentHandler.textCallback != nil && len(text) > 0 {
		d.currentHandler.textCallback(text)
	}
}

func Assign(v *string) TextCallback {
	return func(c CharData) {
		*v = string(c)
	}
}

func Append(a *[]string) TextCallback {
	return func(c CharData) {
		*a = append(*a, string(c))
	}
}

type CharData xml.CharData
type Attrs []xml.Attr

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

func (a Attrs) GetBool(name string, fallback bool) bool {
	strVal, ok := a.Get(name)
	if !ok {
		return fallback
	}

	val, err := strconv.ParseBool(strVal)
	if err != nil {
		return fallback
	}

	return val
}

func (a Attrs) GetFloat(name string, bitsize int, fallback float64) float64 {
	strVal, ok := a.Get(name)
	if !ok {
		return fallback
	}

	val, err := strconv.ParseFloat(strVal, bitsize)
	if err != nil {
		return fallback
	}

	return val
}

func (a Attrs) GetInt(name string, base int, bitsize int, fallback int64) int64 {
	strVal, ok := a.Get(name)
	if !ok {
		return fallback
	}

	val, err := strconv.ParseInt(strVal, base, bitsize)
	if err != nil {
		return fallback
	}

	return val
}

func (a Attrs) GetUInt(name string, base int, bitsize int, fallback uint64) uint64 {
	strVal, ok := a.Get(name)
	if !ok {
		return fallback
	}

	val, err := strconv.ParseUint(strVal, base, bitsize)
	if err != nil {
		return fallback
	}

	return val
}
