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

			decoder.On("first-name/$text", func(text exml.CharData) {
				contact.FirstName = string(text)
			})

			decoder.On("last-name/$text", func(text exml.CharData) {
				contact.LastName = string(text)
			})

			decoder.On("address/$text", func(text exml.CharData) {
				contact.Address = string(text)
			})
		})
	})

	decoder.Run()

	fmt.Printf("Address book: %s\n", addressBook.Name)
	for _, c := range addressBook.Contacts {
		fmt.Printf("- %s %s @ %s\n", c.FirstName, c.LastName, c.Address)
	}
}
