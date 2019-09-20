// Copyright 2015 The present-tex Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"golang.org/x/tools/present"
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
					return err
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
		return err
	}

	return parseCaptions(doc)
}

func parseImage(elem *present.Image) error {
	var err error

	if elem.Height == 0 || elem.Width == 0 {
		f, err := os.Open(elem.URL)
		if err != nil {
			return fmt.Errorf(
				"error opening file [%s]: %v",
				elem.URL,
				err,
			)
		}
		defer f.Close()
		img, _, err := image.Decode(f)
		if err != nil {
			return fmt.Errorf(
				"error decoding image file [%s]: %v",
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
