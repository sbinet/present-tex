// Copyright 2015 The present-tex Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"golang.org/x/tools/present"
)

func parseCode(doc *present.Doc) error {
	var err error
	for i := range doc.Sections {
		section := &doc.Sections[i]
		for ii, elem := range section.Elem {
			switch elem := elem.(type) {
			default:
				continue
			case present.Code:
				hasCode = true
				if elem.Ext == ".cxx" {
					elem.Ext = ".cpp"
				}
				section.Elem[ii] = elem
			}
		}
	}
	return err
}
