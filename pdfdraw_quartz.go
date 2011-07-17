// Copyright (c) 2011 Mikkel Krautz
// The use of this source code is goverened by a BSD-style
// license that can be found in the LICENSE-file.

package pdfdraw

/*
#cgo CFLAGS: -x objective-c 
#cgo LDFLAGS: -framework Quartz -framework QuartzCore -framework CoreFoundation -framework ApplicationServices
#include <Quartz/Quartz.h>

static CGPDFDocumentRef OpenDoc(char *path) {
	CFStringRef pdffn = CFStringCreateWithCStringNoCopy(NULL, path, kCFStringEncodingUTF8,  NULL);
	CFURLRef url = CFURLCreateWithFileSystemPath(NULL, pdffn, kCFURLPOSIXPathStyle, NO);
    CGPDFDocumentRef doc = CGPDFDocumentCreateWithURL(url);
    CFRelease(url);
	return doc;
}

static void ReleaseDoc(CGPDFDocumentRef doc) {
	CFRelease(doc);
}

static CGContextRef CreateBitmapContext(int w, int h, void *data) {
	return CGBitmapContextCreateWithData(data, w, h, 8, 4*w, CGColorSpaceCreateDeviceRGB(), kCGImageAlphaPremultipliedLast | kCGBitmapByteOrder32Big, NULL, NULL);
}

static void ReleaseBitmapContext(CGContextRef ctx) {
	CFRelease(ctx);
}

static void FillContextWithColor(CGContextRef ctx, int w, int h, float r, float g, float b, float a) {
	CGFloat color[4] = { r, g, b, a };
	CGContextSetFillColor(ctx, color);
	CGRect drawRect = CGRectMake(0, 0, w, h);
	CGContextFillRect(ctx, drawRect);
}

static void RenderPageToContext(CGContextRef ctx, CGPDFPageRef page, int w, int h) {
	CGRect boxRect = CGPDFPageGetBoxRect(page, kCGPDFArtBox);
	CGRect drawRect = CGRectMake(0, 0, w, h);
	CGContextSaveGState(ctx);
	CGContextTranslateCTM(ctx, drawRect.origin.x, drawRect.origin.y);
	CGContextScaleCTM(ctx, drawRect.size.width / boxRect.size.width, drawRect.size.height / boxRect.size.height);
	CGContextTranslateCTM(ctx, -boxRect.origin.x, -boxRect.origin.y);
	CGContextDrawPDFPage(ctx, page);
	CGContextRestoreGState(ctx);
}
*/
import "C"

import (
	"image"
	"os"
	"reflect"
	"runtime"
	"unsafe"
)

type quartzDocument struct {
	doc C.CGPDFDocumentRef
}

type quartzPage struct {
	page C.CGPDFPageRef
}

func init() {
	RegisterBackend("quartz", openQuartzDoc)
}

func openQuartzDoc(path string) (doc Document, err os.Error) {
	qd := new(quartzDocument)
	qd.doc = C.OpenDoc(C.CString(path))
	if qd.doc == nil {
		return nil, os.NewError("unable to open pdf file")
	}
	runtime.SetFinalizer(qd, func(qd *quartzDocument) {
		C.ReleaseDoc(qd.doc)
	})
	return qd, nil
}

func (doc *quartzDocument) NumPages() int {
	return int(C.CGPDFDocumentGetNumberOfPages(doc.doc))
}

func (doc *quartzDocument) Page(idx int) Page {
	qp := new(quartzPage)
	qp.page = C.CGPDFDocumentGetPage(doc.doc, C.size_t(idx+1))
	return qp
}

func (page *quartzPage) Size() (width, height float64) {
	rect := C.CGPDFPageGetBoxRect(page.page, C.kCGPDFArtBox)
	x, y, w, h := float64(rect.origin.x), float64(rect.origin.y), float64(rect.size.width), float64(rect.size.height)
	return x - w, y - h
}

func (page *quartzPage) Render(width int, height int, opts *RenderOptions) image.Image {
	w, h := C.int(width), C.int(height)

	rgba := image.NewRGBA(width, height)
	_, ptr := unsafe.Reflect(rgba.Pix)
	hdrp := (*reflect.SliceHeader)(ptr)
	ctx := C.CreateBitmapContext(w, h, unsafe.Pointer(hdrp.Data))
	defer C.ReleaseBitmapContext(ctx)

	fillColor := image.RGBAColor{255, 255, 255, 255}
	if opts != nil {
		fillColor = opts.FillColor
	}
	C.FillContextWithColor(ctx, w, h,
		C.float(float32(fillColor.R)/255.0),
		C.float(float32(fillColor.G)/255.0),
		C.float(float32(fillColor.B)/255.0),
		C.float(float32(fillColor.A)/255.0))

	C.RenderPageToContext(ctx, page.page, w, h)
	return rgba
}
