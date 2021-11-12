// Copyright 2021 The present-tex Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"strings"

	"github.com/sbinet/present-tex/latex"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"golang.org/x/tools/present"
)

var replacer = strings.NewReplacer(
	"-->", `$\Rightarrow$ `,
	"<--", `$\Leftarrow$ `,
	"->", `$\rightarrow$ `,
	"<-", `$\leftarrow$ `,
	"⇒", `$\Rightarrow$ `,
	"—", `---`,
	"±", `$\pm$`,
)

func renderAsLaTeX(input []byte) (present.Elem, error) {
	md := goldmark.New(goldmark.WithRenderer(latex.New(*dpi)))
	reader := text.NewReader(input)
	doc := md.Parser().Parse(reader)
	err := fixupMarkdown(doc)
	if err != nil {
		return nil, err
	}

	var b strings.Builder
	if err := md.Renderer().Render(&b, input, doc); err != nil {
		return nil, err
	}
	return Latex{Latex: replacer.Replace(b.String())}, nil
}

func fixupMarkdown(n ast.Node) error {
	err := ast.Walk(n, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			switch n := n.(type) {
			case *ast.Link:
				n.SetAttributeString("target", []byte("_blank"))
				// https://developers.google.com/web/tools/lighthouse/audits/noopener
				n.SetAttributeString("rel", []byte("noopener"))
			}
		}
		return ast.WalkContinue, nil
	})

	if err != nil {
		return fmt.Errorf("could not fixup markdown: %w", err)
	}

	return nil
}

type Latex struct {
	Cmd   string // original command from present source
	Latex string
}

func (s Latex) PresentCmd() string { return s.Cmd }
func (Latex) TemplateName() string { return "latex" }

var _ present.Elem = (*Latex)(nil)
