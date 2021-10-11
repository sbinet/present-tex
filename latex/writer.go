// Copyright 2021 The present-tex Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package latex

import (
	"strings"

	"github.com/yuin/goldmark/util"
)

type writer struct{}

func newWriter() writer { return writer{} }

func (w *writer) RawWrite(wbuf util.BufWriter, src []byte) {
	// FIXME(sbinet)
	_, _ = wbuf.Write(src)

	//	n := 0
	//	l := len(src)
	//	for i := 0; i < l; i++ {
	//		v := escapeLaTeX(src[i])
	//		if v != nil {
	//			_, _ = wbuf.Write(src[i-n : i])
	//			n = 0
	//			_, _ = wbuf.Write(v)
	//			continue
	//		}
	//		n++
	//	}
	//	if n != 0 {
	//		_, _ = wbuf.Write(src[l-n:])
	//	}
}

func (w *writer) Write(wbuf util.BufWriter, src []byte) {
	// FIXME(sbinet)
	w.RawWrite(wbuf, src)
}

// UTF8 replaces UTF8 code sequences with their LaTeX equivalent.
func UTF8(s string) string {
	return latexUTF8.Replace(s)
}

// escapeLaTeX escapes characters that should be escaped in LaTeX text.
func escapeLaTeX(src []byte) []byte {
	return []byte(latexRepl.Replace(string(src)))
}

var (
	latexRepl = strings.NewReplacer(
		"-->", `$\Rightarrow$ `,
		"<--", `$\Leftarrow$ `,
		"->", `$\rightarrow$ `,
		"<-", `$\leftarrow$ `,
		"⇒", `$\Rightarrow$ `,
		"—", `---`,

		`\`, `\textbackslash`,
		"_", `\_`,
		"&", `\&`,
		"$", `\$`,
		"^", `\^{}`,
		"%", `\%`,
		"~", `$\sim$`,
		"#", `\#`,
		"{", `\{`,
		"}", `\}`,
		"é", `\'e`,
		"è", `\`+"`e",
		"à", `\`+"`a",
		"ù", `\`+"`u",
		"â", `\^a`,
		"ê", `\^e`,
		"î", `\^i`,
		"ô", `\^o`,
		"û", `\^u`,
		"ŷ", `\^y`,
		"ä", `\"a`,
		"ë", `\"e`,
		"ï", `\"i`,
		"ö", `\"o`,
		"ü", `\"u`,
		"ÿ", `\"y`,
		"ç", `\c{c}`,
		"Ç", `\c{C}`,
		"æ", `\ae `,
		"œ", `\oe `,
	)
	latexUTF8 = strings.NewReplacer(
		"⇒", `$\Rightarrow$ `,
		"—", `---`,

		"é", `\'e`,
		"è", `\`+"`e",
		"à", `\`+"`a",
		"ù", `\`+"`u",
		"â", `\^a`,
		"ê", `\^e`,
		"î", `\^i`,
		"ô", `\^o`,
		"û", `\^u`,
		"ŷ", `\^y`,
		"ä", `\"a`,
		"ë", `\"e`,
		"ï", `\"i`,
		"ö", `\"o`,
		"ü", `\"u`,
		"ÿ", `\"y`,
		"ç", `\c{c}`,
		"Ç", `\c{C}`,
		"æ", `\ae `,
		"œ", `\oe `,
	)
)
