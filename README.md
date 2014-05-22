# exml [![Build Status](https://drone.io/github.com/lucsky/go-exml/status.png)](https://drone.io/github.com/lucsky/go-exml/latest)

The **exml** package provides an intuitive event based XML parsing API which sits on top of a standard Go ```encoding/xml/Decoder```, greatly simplifying the parsing code while retaining the raw speed and low memory overhead of the underlying stream engine, regardless of the size of the input. The module takes care of the complex tasks of maintaining contexts between event handlers allowing you to concentrate on dealing with the actual structure of the XML document.

# Installation

**HEAD:**

```go get github.com/lucsky/go-exml```

**v3:**

```go get http://gopkg.in/lucsky/go-exml.v3```

The third version of **exml** provides compile time callback safety at the cost of an **API CHANGE**. Ad hoc ```$text``` events have been replaced by the specific ```OnText``` and ```OnTextOf``` event registration methods.

**v2:**

```go get http://gopkg.in/lucsky/go-exml.v2```

The second version of **exml** has a better implementation based on a dynamic handler tree, allowing global events (see example below), having lower memory usage and also being faster.

**v1:**

```go get http://gopkg.in/lucsky/go-exml.v1```

Initial (and naive) implementation based on a flat list of absolute event paths.

# Usage

The best way to illustrate how **exml** makes parsing very easy is to look at actual examples. Consider the following contrived sample document:

```xml
<?xml version="1.0"?>
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
</address-book>
```

Here is a way to parse it into an array of contact objects using **exml**:

```go
package main

import (
    "fmt"
    "os"

    "gopkg.in/lucsky/go-exml.v3"
)

type AddressBook struct {
    Name     string
    Contacts []*Contact
}

type Contact struct {
    FirstName string
    LastName  string
    Address   string
}

func main() {
    reader, _ := os.Open("example.xml")
    defer reader.Close()

    addressBook := AddressBook{}
    decoder := exml.NewDecoder(reader)

    decoder.On("address-book", func(attrs exml.Attrs) {
        addressBook.Name, _ = attrs.Get("name")

        decoder.On("contact", func(attrs exml.Attrs) {
            contact := &Contact{}
            addressBook.Contacts = append(addressBook.Contacts, contact)

            decoder.On("first-name", func(attrs exml.Attrs) {
                decoder.OnText(func(text exml.CharData) {
                    contact.FirstName = string(text)
                })
            })

            decoder.On("last-name", func(attrs exml.Attrs) {
                decoder.OnText(func(text exml.CharData) {
                    contact.LastName = string(text)
                })
            })

            decoder.On("address", func(attrs exml.Attrs) {
                decoder.OnText(func(text exml.CharData) {
                    contact.Address = string(text)
                })
            })
        })
    })

    decoder.Run()

    fmt.Printf("Address book: %s\n", addressBook.Name)
    for _, c := range addressBook.Contacts {
        fmt.Printf("- %s %s @ %s\n", c.FirstName, c.LastName, c.Address)
    }
}
```

To reduce the amount and depth of event callbacks that you have to write, **exml** allows you to register handlers on **events paths**:

```go
decoder.OnTextOf("address-book/contact/first-name", func(text exml.CharData) {
    fmt.Println("First name: ", string(text))
})

// This works too:
decoder.On("address-book/contact", func(attrs exml.Attrs) {
    decoder.OnTextOf(last-name", func(text exml.CharData) {
        fmt.Println("Last name: ", string(text))
    })
})

// And this as well:
decoder.On("address-book/contact/address", func(attrs exml.Attrs) {
    decoder.OnText(func(text exml.CharData) {
        fmt.Println("Address: ", string(text))
    })
})
```

Finally, since using nodes text content to initialize struct fields is a pretty frequent task, **exml** provides a shortcut to make it shorter to write. Let's revisit our address book example and use this shortcut:

```go
contact := &Contact{}
decoder.OnTextOf("first-name", exml.Assign(&contact.FirstName))
decoder.OnTextOf("last-name", exml.Assign(&contact.LastName))
decoder.OnTextOf("address", exml.Assign(&contact.Address))
```

Another shortcut allows to accumulate text content from various nodes to a single slice:

```go
info := []string{}
decoder.OnTextOf("first-name", decoder.Append(&info))
decoder.OnTextOf("last-name", decoder.Append(&info))
decoder.OnTextOf("address", decoder.Append(&info))
```

The second version (aka v2) of **exml** introduced global events which allow to register a top level handler that would be picked up at any level whenever a corresponding XML node is encountered. For example, this snippet would allow to print all text nodes regardless of their depth and parent tag:

```go
decoder := exml.NewDecoder(reader)
decoder.OnText(func(text CharData) {
    fmt.Println(string(text))
})
```

# API

The full API is visible at the **exml** [gopkg.in][gopkg] page.

# Benchmarks

The included benchmarks show that **exml** can be *massively* faster than standard unmarshaling and the difference would most likely be even greater for bigger inputs.

```shell
% go test -bench . -benchmem
OK: 23 passed
PASS
Benchmark_UnmarshalSimple      50000         57156 ns/op        6138 B/op        128 allocs/op
Benchmark_UnmarshalText       100000         22423 ns/op        3452 B/op         61 allocs/op
Benchmark_UnmarshalCDATA      100000         23460 ns/op        3483 B/op         61 allocs/op
Benchmark_UnmarshalMixed      100000         28807 ns/op        4034 B/op         67 allocs/op
Benchmark_DecodeSimple       5000000           388 ns/op          99 B/op          3 allocs/op
Benchmark_DecodeText         5000000           485 ns/op         114 B/op          3 allocs/op
Benchmark_DecodeCDATA        5000000           485 ns/op         114 B/op          3 allocs/op
Benchmark_DecodeMixed        5000000           487 ns/op         114 B/op          3 allocs/op
ok      github.com/lucsky/go-exml   11.194s
```

# License

Code is under the [BSD 2 Clause (NetBSD) license][license].

[license]:https://github.com/lucsky/go-exml/tree/master/LICENSE
[gopkg]:http://gopkg.in/lucsky/go-exml.v3
