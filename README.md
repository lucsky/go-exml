# go-exml

The **go-exml** package provides an intuitive event based XML parsing API which sits on top of a standard Go ```encoding/xml/Decoder```, greatly simplifying the parsing code while retaining the raw speed and low memory overhead of the underlying stream engine, regardless of the size of the input. The module takes care of the complex tasks of maintaining contexts between event handlers allowing you to concentrate on dealing with the actual structure of the XML document.

# Installation

```go get github.com/lucsky/go-exml```

or

```go get http://gopkg.in/lucsky/go-exml.v1```

# Usage

The best way to illustrate how **go-exml** makes parsing very easy is to look at actual examples. Consider the following contrived sample document:

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

Here is a way to parse it into an array of contact objects using **go-exml**:

```go
package main

import (
    "fmt"
    "os"

    "github.com/lucsky/go-exml"
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
                decoder.On("$text", func(text exml.CharData) {
                    contact.FirstName = string(text)
                })
            })

            decoder.On("last-name", func(attrs exml.Attrs) {
                decoder.On("$text", func(text exml.CharData) {
                    contact.LastName = string(text)
                })
            })

            decoder.On("address", func(attrs exml.Attrs) {
                decoder.On("$text", func(text exml.CharData) {
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

To reduce the amount and depth of event callbacks that you have to write, **go-exml** provides **stacked events**. Here's how you can simplify the last example using stacked events:

```go
...
...

contact := &Contact{}

decoder.On("first-name/$text", func(text exml.CharData) {
    contact.FirstName = string(text)
})

decoder.On("last-name/$text", func(text exml.CharData) {
    contact.LastName = string(text)
})

decoder.On("address/$text", func(text exml.CharData) {
    contact.Address = string(text)
})

...
...
```

Finally, since using nodes text content to initialize struct fields is a pretty frequent task, **go-exml** provides a shortcut to make it shorter to write. Let's revisit the previous example and use this shortcut:

```go
...
...

contact := &Contact{}
decoder.On("first-name/$text", decoder.Assign(&contact.FirstName))
decoder.On("last-name/$text", decoder.Assign(&contact.LastName))
decoder.On("address/$text", decoder.Assign(&contact.Address))

...
...
```

Another shortcut allows to accumulate text content from various nodes to a single slice:

```go
...
...

info := []string{}
decoder.On("first-name/$text", decoder.Append(&info))
decoder.On("last-name/$text", decoder.Append(&info))
decoder.On("address/$text", decoder.Append(&info))

...
...
```

# Benchmarks

The included benchmarks show that **go-exml** can be *massively* faster than standard unmarshaling and the difference would most likely be even greater for bigger inputs.

```shell
% go test -bench . -benchmem
OK: 16 passed
PASS
Benchmark_UnmarshalSimple      50000         55643 ns/op        6141 B/op        128 allocs/op
Benchmark_UnmarshalText       100000         23253 ns/op        3522 B/op         64 allocs/op
Benchmark_UnmarshalCDATA      100000         24455 ns/op        3611 B/op         64 allocs/op
Benchmark_UnmarshalMixed      100000         29596 ns/op        4120 B/op         70 allocs/op
Benchmark_DecodeSimple       5000000           375 ns/op          57 B/op          3 allocs/op
Benchmark_DecodeText         5000000           541 ns/op         123 B/op          5 allocs/op
Benchmark_DecodeCDATA        5000000           541 ns/op         123 B/op          5 allocs/op
Benchmark_DecodeMixed        5000000           541 ns/op         123 B/op          5 allocs/op
ok      github.com/lucsky/go-exml   23.966s
```

# License

Code is under the [BSD 2 Clause (NetBSD) license][license].

[license]:https://github.com/lucsky/go-exml/tree/master/LICENSE
