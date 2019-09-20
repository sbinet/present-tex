// Copyright 2015 The present-tex Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"image"
	"os"
	"os/exec"
	"strings"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/vp8"
	_ "golang.org/x/image/webp"

	"golang.org/x/tools/present"
	"golang.org/x/xerrors"
)

type Image struct {
	present.Image
	HasCaption bool
	Caption    present.Caption
}

func parseImages(doc *present.Doc) error {
	var err error
	for i := range doc.Sections {
		section := &doc.Sections[i]
		for j := range section.Elem {
			elem := section.Elem[j]
			switch elem := elem.(type) {
			default:
				continue
			case present.Image:
				err = parseImage(&elem)
				if err != nil {
					return xerrors.Errorf("could not parse image %w: %w", elem.URL, err)
				}
				img := Image{Image: elem}
				if j+1 < len(section.Elem) {
					if elem, ok := section.Elem[j+1].(present.Caption); ok {
						err = parseCaption(&elem)
						if err != nil {
							return err
						}
						img.HasCaption = true
						img.Caption = elem
					}
				}
				section.Elem[j] = img
			}
		}
	}

	if err != nil {
		return xerrors.Errorf("could not parse images: %w", err)
	}

	return parseCaptions(doc)
}

func parseImage(elem *present.Image) error {
	var err error

	if strings.HasSuffix(elem.URL, ".svg") {
		oname := elem.URL[:len(elem.URL)-len(".svg")] + "_svg.png"
		err := exec.Command("convert", elem.URL, oname).Run()
		if err != nil {
			return xerrors.Errorf(
				"could not convert SVG image %q to PNG: %w",
				elem.URL, err,
			)
		}
		elem.URL = oname
	}

	if elem.Height == 0 || elem.Width == 0 {
		f, err := os.Open(elem.URL)
		if err != nil {
			return xerrors.Errorf(
				"error opening file [%s]: %w",
				elem.URL,
				err,
			)
		}
		defer f.Close()
		img, _, err := image.Decode(f)
		if err != nil {
			return xerrors.Errorf(
				"error decoding image file [%s]: %w",
				elem.URL,
				err,
			)
		}
		h := img.Bounds().Dy()
		w := img.Bounds().Dx()

		switch {
		case elem.Height == 0 && elem.Width == 0:
			elem.Height = h
			elem.Width = w
		case elem.Height == 0 && elem.Width != 0:
			// rescale, keeping ratio
			ratio := float64(elem.Width) / float64(w)
			elem.Height = int(float64(h) * ratio)
		case elem.Height != 0 && elem.Width == 0:
			// rescale, keeping ratio
			ratio := float64(elem.Height) / float64(h)
			elem.Width = int(float64(w) * ratio)

		}

	}

	// rescale height/width to a (default=72) DPI resolution
	// height and width are now in inches.
	elem.Height /= *dpi
	elem.Width /= *dpi

	return err
}
