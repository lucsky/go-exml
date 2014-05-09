# go-exml

The **go-exml** package provides an intuitive event based XML parsing API which sits on top of a standard Go ```encoding/xml/Decoder```, greatly simplifying the parsing code while retaining the raw speed and low memory overhead of the underlying stream engine, regardless of the size of the input. The module takes care of the complex tasks of maintaining contexts between event handlers allowing you to concentrate on dealing with the actual structure of the XML document.

# Installation

```go get github.com/lucsky/go-exml```

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

# License

Code is under the [BSD 2 Clause (NetBSD) license][license].

[license]:https://github.com/lucsky/go-exml/tree/master/LICENSE
