// Copyright (c) 2011 Mikkel Krautz
// The use of this source code is goverened by a BSD-style
// license that can be found in the LICENSE-file.

// Package pdfdraw implements a library for rendering and reading PDFs
package pdfdraw

import (
	"image"
	"os"
)

// RenderOptions specifies a set of options for rendering a page.
type RenderOptions struct {
	// The color to fill the image with before it is drawn.
	FillColor image.RGBAColor

	// Disable anti-aliasing
	NoAA bool
}

// Document represents a PDF document
type Document interface {
	// Get the total number of pages in the Document
	NumPages() int
	// Get a Page
	Page(idx int) Page
}

// Page represents a page in a Document
type Page interface {
	// Get the internal size of the Page
	Size() (width float64, height float64)
	// Render the page to an image.Image
	Render(width int, height int, opts *RenderOptions) image.Image
}

// A BackendOpener is a function that can be used to open a Document using a specific backend
type BackendOpener func(path string) (doc Document, err os.Error)

var backends map[string]BackendOpener

func init() {
	backends = make(map[string]BackendOpener)
}

// Register a new pdfdraw backend
func RegisterBackend(name string, opener BackendOpener) {
	backends[name] = opener
}

// Open a PDF document using the default backend
func Open(path string) (doc Document, err os.Error) {
	for _, opener := range backends {
		doc, err = opener(path)
		return
	}
	return nil, os.NewError("pdfdraw: no available backends")
}

// Get a list of all available backends
func Backends() (names []string) {
	for name, _ := range backends {
		names = append(names, name)
	}
	return
}

// Open a PDF document using a specific backend
func OpenBackend(path string, backend string) (doc Document, err os.Error) {
	opener, ok := backends[backend]
	if !ok {
		return nil, os.NewError("pdfdraw: no such backend")
	}

	return opener(path)
}
