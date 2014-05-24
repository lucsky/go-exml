/*
The exml package provides an intuitive event based XML parsing API which sits
on top of a standard Go xml.Decoder, greatly simplifying the parsing code
while retaining the raw speed and low memory overhead of the underlying stream
engine, regardless of the size of the input. The package takes care of the
complex tasks of maintaining contexts between event handlers allowing you to
concentrate on dealing with the actual structure of the XML document.
*/
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

// A Decoder wraps an xml.Decoder and maintains the various states
// between the encountered XML nodes during parsing.
type Decoder struct {
	decoder        *xml.Decoder
	topHandler     *handler
	currentHandler *handler
	errorCallback  ErrorCallback
}

// NewDecoder creates a new exml parser reading from r.
func NewDecoder(r io.Reader) *Decoder {
	return NewCustomDecoder(xml.NewDecoder(r))
}

// NewCustomDecoder creates a new exml parser reading from the passed
// xml.Decoder which is useful when you need to configure the underlying
// decoder, when you need to handle non-UTF8 xml documents for example.
func NewCustomDecoder(d *xml.Decoder) *Decoder {
	topHandler := &handler{}
	return &Decoder{
		decoder:        d,
		topHandler:     topHandler,
		currentHandler: topHandler,
	}
}

// On registers a handler for a single tag or for a path.
func (d *Decoder) On(path string, callback TagCallback) {
	h := d.installHandlers(path)
	h.tagCallback = callback
}

// OnTextOf registers a handler for the text content of a single tag or
// for the text content at a certain path.
func (d *Decoder) OnTextOf(path string, callback TextCallback) {
	h := d.installHandlers(path)
	h.textCallback = callback
}

// OnText registers a handler for the text content of the current tag.
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

// OnError registers a global error handler which will be called whenever
// the underlying xml.Decoder reports an error.
func (d *Decoder) OnError(handler ErrorCallback) {
	d.errorCallback = handler
}

// Run starts the parsing process.
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

// Assign is a helper function which returns a text callback that assigns
// the text content of the current tag to the passed variable pointer.
func Assign(v *string) TextCallback {
	return func(c CharData) {
		*v = string(c)
	}
}

// AssignBool is a helper function which returns a text callback that
// assigns the text content of the current tag parsed as a bool to the
// passed variable pointer. The accepted text values correspond to the
// one accepted by the strconv.ParseBool() function. The fallback parameter
// value is used when the parsing of the text content fails.
func AssignBool(v *bool, fallback bool) TextCallback {
	return func(c CharData) {
		val, err := strconv.ParseBool(string(c))
		if err == nil {
			*v = val
		} else {
			*v = fallback
		}
	}
}

// AssignFloat is a helper function which returns a text callback that
// assigns the text content of the current tag parsed as a float to the
// passed variable pointer. The accepted text values correspond to the
// one accepted by the strconv.ParseFloat() function. The fallback parameter
// value is used when the parsing of the text content fails.
func AssignFloat(v *float64, bitsize int, fallback float64) TextCallback {
	return func(c CharData) {
		val, err := strconv.ParseFloat(string(c), bitsize)
		if err == nil {
			*v = val
		} else {
			*v = fallback
		}
	}
}

// AssignInt is a helper function which returns a text callback that
// assigns the text content of the current tag parsed as an int to the
// passed variable pointer. The accepted text values correspond to the
// one accepted by the strconv.ParseInt() function. The fallback parameter
// value is used when the parsing of the text content fails.
func AssignInt(v *int64, base int, bitsize int, fallback int64) TextCallback {
	return func(c CharData) {
		val, err := strconv.ParseInt(string(c), base, bitsize)
		if err == nil {
			*v = val
		} else {
			*v = fallback
		}
	}
}

// AssignUInt is a helper function which returns a text callback that
// assigns the text content of the current tag parsed as an uint to the
// passed variable pointer. The accepted text values correspond to the
// one accepted by the strconv.ParseUint() function. The fallback parameter
// value is used when the parsing of the text content fails.
func AssignUInt(v *uint64, base int, bitsize int, fallback uint64) TextCallback {
	return func(c CharData) {
		val, err := strconv.ParseUint(string(c), base, bitsize)
		if err == nil {
			*v = val
		} else {
			*v = fallback
		}
	}
}

// Append is a helper function which returns a text callback that appends
// the text content of the current tag to the passed strings slice pointer.
func Append(a *[]string) TextCallback {
	return func(c CharData) {
		*a = append(*a, string(c))
	}
}

// AppendBool is a helper function which returns a text callback that appends
// the text content of the current tag parsed as a bool to the passed slice
// pointer. The accepted text values correspond to the one accepted by the
// strconv.ParseBool() function. The fallback parameter value is used when
// the parsing of the text content fails.
func AppendBool(a *[]bool, fallback bool) TextCallback {
	return func(c CharData) {
		var val bool
		AssignBool(&val, fallback)(c)
		*a = append(*a, val)
	}
}

// AppendFloat is a helper function which returns a text callback that appends
// the text content of the current tag parsed as a float to the passed slice
// pointer. The accepted text values correspond to the one accepted by the
// strconv.ParseFloat() function. The fallback parameter value is used when
// the parsing of the text content fails.
func AppendFloat(a *[]float64, bitsize int, fallback float64) TextCallback {
	return func(c CharData) {
		var val float64
		AssignFloat(&val, bitsize, fallback)(c)
		*a = append(*a, val)
	}
}

// AppendInt is a helper function which returns a text callback that appends
// the text content of the current tag parsed as an int to the passed slice
// pointer. The accepted text values correspond to the one accepted by the
// strconv.ParseInt() function. The fallback parameter value is used when
// the parsing of the text content fails.
func AppendInt(a *[]int64, base int, bitsize int, fallback int64) TextCallback {
	return func(c CharData) {
		var val int64
		AssignInt(&val, base, bitsize, fallback)(c)
		*a = append(*a, val)
	}
}

// AppendUInt is a helper function which returns a text callback that appends
// the text content of the current tag parsed as an uint to the passed slice
// pointer. The accepted text values correspond to the one accepted by the
// strconv.ParseUint() function. The fallback parameter value is used when
// the parsing of the text content fails.
func AppendUInt(a *[]uint64, base int, bitsize int, fallback uint64) TextCallback {
	return func(c CharData) {
		var val uint64
		AssignUInt(&val, base, bitsize, fallback)(c)
		*a = append(*a, val)
	}
}

type CharData xml.CharData
type Attrs []xml.Attr

// Get returns the value of the requested attribute and true when the
// attributes exists, or an empty string and false when it doesn't.
func (a Attrs) Get(name string) (string, bool) {
	for _, attr := range a {
		if attr.Name.Local == name {
			return attr.Value, true
		}
	}

	return "", false
}

// GetString returns the value of the requested attribute when it exists
// or the passed fallback value when it doesn't.
func (a Attrs) GetString(name string, fallback string) string {
	val, ok := a.Get(name)
	if !ok {
		return fallback
	}

	return val
}

// GetBool returns the value of the requested attribute as a bool when it
// exists or the passed fallback value when it doesn't. The accepted values
// corresponds to the one accepted by the strconv.ParseBool() function.
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

// GetFloat returns the value of the requested attribute as a float when it
// exists or the passed fallback value when it doesn't. The additional
// parameters and the accepted values corresponds to the one accepted by
// the strconv.ParseFloat() function.
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

// GetInt returns the value of the requested attribute as an int when it
// exists or the passed fallback value when it doesn't. The additional
// parameters and the accepted values corresponds to the one accepted by
// the strconv.ParseInt() function.
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

// GetUInt returns the value of the requested attribute as an uint when it
// exists or the passed fallback value when it doesn't. The additional
// parameters and the accepted values corresponds to the one accepted by
// the strconv.ParseUint() function.
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
