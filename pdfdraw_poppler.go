// Copyright (c) 2011 Mikkel Krautz
// The use of this source code is goverened by a BSD-style
// license that can be found in the LICENSE-file.

package pdfdraw

/*
#cgo pkg-config: poppler-glib cairo
#include <poppler.h>
#include <cairo.h>
*/
import "C"

import (
	"errors"
	"image"
	"image/color"
	"io/ioutil"
	"reflect"

	"unsafe"
)

type popplerDocument struct {
	doc  *C.PopplerDocument
	data []byte // keep for gc
}

type popplerPage struct {
	page *C.PopplerPage
}

func init() {
	RegisterBackend("poppler", popplerOpenDoc)
}

func popplerOpenDoc(path string) (Document, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	doc := C.poppler_document_new_from_data((*C.char)(unsafe.Pointer(&data[0])), C.int(len(data)), nil, nil)
	if doc == nil {
		return nil, errors.New("unable to open file")
	}
	return &popplerDocument{doc, data}, nil
}

func (doc *popplerDocument) Close() error {
	return nil
}

func (doc *popplerDocument) NumPages() int {
	return int(C.poppler_document_get_n_pages(doc.doc))
}

func (doc *popplerDocument) Page(idx int) (page Page) {
	pp := new(popplerPage)
	pp.page = C.poppler_document_get_page(doc.doc, C.int(idx))
	return pp
}

func (page *popplerPage) Size() (width, height float64) {
	var (
		dw, dh C.double
	)
	C.poppler_page_get_size(page.page, &dw, &dh)
	return float64(dw), float64(dh)
}

func (page *popplerPage) Render(width int, height int, opts *RenderOptions) image.Image {
	surface := C.cairo_image_surface_create(C.CAIRO_FORMAT_ARGB32, C.int(width), C.int(height))
	defer C.cairo_surface_destroy(surface)

	ctx := C.cairo_create(surface)
	defer C.cairo_destroy(ctx)

	ow, oh := page.Size()
	fw := float64(width)
	fh := float64(height)
	sw, sh := float64(fw/ow), float64(fh/oh)
	C.cairo_scale(ctx, C.double(sw), C.double(sh))

	fillColor := color.RGBA{255, 255, 255, 255}
	if opts != nil {
		fillColor = opts.FillColor
	}
	C.cairo_set_source_rgba(ctx, C.double(float64(fillColor.R)/float64(255)),
		C.double(float64(fillColor.G)/float64(255)),
		C.double(float64(fillColor.B)/float64(255)),
		C.double(float64(fillColor.A)/float64(255)))
	C.cairo_rectangle(ctx, 0, 0, C.double(width), C.double(height))
	C.cairo_fill(ctx)

	if opts != nil && opts.NoAA {
		C.cairo_set_antialias(ctx, C.CAIRO_ANTIALIAS_NONE)
	}

	C.poppler_page_render_for_printing(page.page, ctx)
	dataPtr := C.cairo_image_surface_get_data(surface)
	data := *(*[]uint8)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(dataPtr)),
		Len:  width * height * 4,
		Cap:  width * height * 4,
	}))

	nrgba := image.NewRGBA(image.Rect(0, 0, width, height))
	copy(nrgba.Pix, data)

	// cairo bgra -> go rgba
	for i := 0; i < len(data); i += 4 {
		nrgba.Pix[i], nrgba.Pix[i+2] = nrgba.Pix[i+2], nrgba.Pix[i]
	}

	return nrgba

}
