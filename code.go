package main

import (
	"golang.org/x/tools/present"
)

func parseCode(doc *present.Doc) error {
	var err error
	for i := range doc.Sections {
		section := &doc.Sections[i]
		for _, elem := range section.Elem {
			switch elem.(type) {
			default:
				continue
			case present.Code:
				hasCode = true
				return err
			}
		}
	}
	return err
}
