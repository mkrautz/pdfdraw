// Copyright (c) 2011 Mikkel Krautz
// The use of this source code is goverened by a BSD-style
// license that can be found in the LICENSE-file.

package pdfdraw

import (
	"image/png"
	"os"
	"testing"
)

func TestRenderFirstPage(t *testing.T) {
	doc, err := Open("R-intro.pdf")
	if err != nil {
		t.Fatal(err)
	}
	defer doc.Close()

	page := doc.Page(0)
	img := page.Render(1024, 1448, nil)

	f, err := os.Create("test.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	err = png.Encode(f, img)
	if err != nil {
		t.Fatal(err)
	}
}
