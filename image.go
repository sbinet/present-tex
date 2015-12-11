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
				section.Elem[j] = elem
			}
		}
	}
	return err
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
