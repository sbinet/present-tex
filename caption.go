// Copyright 2015 The present-tex Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"golang.org/x/tools/present"
)

func parseCaptions(doc *present.Doc) error {
	var err error
	for i := range doc.Sections {
		section := &doc.Sections[i]
		var captions []int
		for j := range section.Elem {
			elem := section.Elem[j]
			switch elem.(type) {
			default:
				continue
			case present.Caption:
				captions = append(captions, j)
			}
		}
		for j := len(captions) - 1; j >= 0; j-- {
			idx := captions[j]
			section.Elem = append(section.Elem[:idx], section.Elem[idx+1:]...)
		}
	}
	return err
}

func parseCaption(elem *present.Caption) error {
	var err error
	elem.Text = renderFont(elem.Text)
	return err
}
