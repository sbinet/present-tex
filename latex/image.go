// Copyright 2021 The present-tex Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package latex

import (
	"fmt"
	"image"
	"os"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/vp8"
	_ "golang.org/x/image/webp"
)

func inferDims(fname string) (w, h int, err error) {
	f, err := os.Open(fname)
	if err != nil {
		return 0, 0, fmt.Errorf(
			"error opening file [%s]: %w",
			fname,
			err,
		)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return 0, 0, fmt.Errorf(
			"error decoding image file [%s]: %w",
			fname,
			err,
		)
	}
	h = img.Bounds().Dy()
	w = img.Bounds().Dx()

	return w, h, nil
}
