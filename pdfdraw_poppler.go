// Copyright (c) 2011 Mikkel Krautz
// The use of this source code is goverened by a BSD-style
// license that can be found in the LICENSE-file.

package pdfdraw

/*
#cgo pkg-config: poppler-glib cairo
#include <stdlib.h>
#include <poppler.h>
#include <cairo.h>

static unsigned char getbyte(unsigned char *buf, int idx) {
	return buf[idx];
}

static char *path_to_uri(char *path) {
	GError *err = NULL;
	gchar *absfn = NULL;

	if (g_path_is_absolute(path))
		absfn = g_strdup(path);
	else {
		gchar *tmp = g_get_current_dir();
		absfn = g_build_filename(tmp, path, NULL);
		free(tmp);
	}

	return (char *) g_filename_to_uri(absfn, NULL, &err);
}
*/
import "C"

import (
	"errors"
	"image"
	"image/color"

	"unsafe"
)

type popplerDocument struct {
	doc *C.PopplerDocument
}

type popplerPage struct {
	page *C.PopplerPage
}

func init() {
	C.g_type_init()
	RegisterBackend("poppler", popplerOpenDoc)
}

func popplerOpenDoc(path string) (doc Document, err error) {
	uri := C.path_to_uri(C.CString(path))
	if uri == nil {
		return nil, errors.New("unable to convert path to uri")
	}
	defer C.free(unsafe.Pointer(uri))

	pd := new(popplerDocument)
	pd.doc = C.poppler_document_new_from_file(uri, nil, nil)
	if pd == nil {
		return nil, errors.New("unable to open file")
	}
	return pd, nil
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
	data := C.cairo_image_surface_get_data(surface)
	nrgba := image.NewNRGBA(image.Rect(0, 0, width, height))
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			nrgba.SetNRGBA(x, y, color.NRGBA{
				R: uint8(C.getbyte(data, C.int(x*4+4*y*width+2))),
				G: uint8(C.getbyte(data, C.int(x*4+4*y*width+1))),
				B: uint8(C.getbyte(data, C.int(x*4+4*y*width+0))),
				A: uint8(C.getbyte(data, C.int(x*4+4*y*width+3))),
			})
		}
	}

	return nrgba
}
