// Copyright 2021 The present-tex Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package latex

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// Renderer renders a CommonMark document as LaTeX-Beamer.
type Renderer struct {
	dpi   int
	w     writer
	funcs map[ast.NodeKind]renderFunc
}

type renderFunc func(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error)

// New returns a new Renderer.
func New(dpi int) *Renderer {
	r := &Renderer{
		dpi:   dpi,
		w:     newWriter(),
		funcs: make(map[ast.NodeKind]renderFunc),
	}

	// table
	// r.register(tast.KindTable, renderTable)
	// r.register(tast.KindTableHeader, renderTableHeader)
	// r.register(tast.KindTableRow, renderTableRow)
	// r.register(tast.KindTableCell, renderTableCell)

	// blocks
	r.register(ast.KindDocument, r.renderDocument)
	r.register(ast.KindHeading, r.renderHeading)
	r.register(ast.KindBlockquote, r.renderBlockquote)
	r.register(ast.KindCodeBlock, r.renderCodeBlock)
	r.register(ast.KindFencedCodeBlock, r.renderFencedCodeBlock)
	r.register(ast.KindHTMLBlock, r.renderHTMLBlock)
	r.register(ast.KindList, r.renderList)
	r.register(ast.KindListItem, r.renderListItem)
	r.register(ast.KindParagraph, r.renderParagraph)
	r.register(ast.KindTextBlock, r.renderTextBlock)
	r.register(ast.KindThematicBreak, r.renderThematicBreak)
	// inlines
	r.register(ast.KindAutoLink, r.renderAutoLink)
	r.register(ast.KindCodeSpan, r.renderCodeSpan)
	r.register(ast.KindEmphasis, r.renderEmphasis)
	r.register(ast.KindImage, r.renderImage)
	r.register(ast.KindLink, r.renderLink)
	r.register(ast.KindRawHTML, r.renderRawHTML)
	r.register(ast.KindText, r.renderText)
	r.register(ast.KindString, r.renderString)

	// strikethrough
	//	r.register(tast.KindStrikethrough, r.renderStrikethrough)

	return r
}

// Render renders a PDF doc
func (r *Renderer) Render(w io.Writer, source []byte, node ast.Node) error {
	writer, ok := w.(util.BufWriter)
	if !ok {
		writer = bufio.NewWriter(w)
		defer writer.Flush()
	}

	err := ast.Walk(node, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		var (
			s   = ast.WalkContinue
			err error
		)
		f := r.funcs[node.Kind()]
		if f != nil {
			s, err = f(writer, source, node, entering)
		}
		return s, err
	})

	if err != nil {
		return err
	}

	return nil
}

// AddOptions adds given option to this renderer.
func (r *Renderer) AddOptions(...renderer.Option) {}

// register a new node render func
func (r *Renderer) register(kind ast.NodeKind, v renderFunc) {
	if r.funcs == nil {
		r.funcs = map[ast.NodeKind]renderFunc{}
	}

	r.funcs[kind] = v
}

var (
	_ renderer.Renderer = (*Renderer)(nil)
)

func (r *Renderer) writeLines(w util.BufWriter, source []byte, n ast.Node) {
	l := n.Lines().Len()
	for i := 0; i < l; i++ {
		line := n.Lines().At(i)
		r.w.RawWrite(w, line.Value(source))
	}
}

// // GlobalAttributeFilter defines attribute names which any elements can have.
// var GlobalAttributeFilter = util.NewBytesFilter(
// 	[]byte("accesskey"),
// 	[]byte("autocapitalize"),
// 	[]byte("class"),
// 	[]byte("contenteditable"),
// 	[]byte("contextmenu"),
// 	[]byte("dir"),
// 	[]byte("draggable"),
// 	[]byte("dropzone"),
// 	[]byte("hidden"),
// 	[]byte("id"),
// 	[]byte("itemprop"),
// 	[]byte("lang"),
// 	[]byte("slot"),
// 	[]byte("spellcheck"),
// 	[]byte("style"),
// 	[]byte("tabindex"),
// 	[]byte("title"),
// 	[]byte("translate"),
// )

func (r *Renderer) renderDocument(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// nothing to do
	return ast.WalkContinue, nil
}

// // HeadingAttributeFilter defines attribute names which heading elements can have
// var HeadingAttributeFilter = GlobalAttributeFilter

var headings = [...]string{
	1: "\n\\section{",
	2: "\n\\subsection{",
	3: "\n\\subsubsection{",
	4: "\n\\paragraph{",
	5: "\n\\subparagraph{",
	6: "\n\\textbf{",
}

func (r *Renderer) renderHeading(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Heading)
	if entering {
		_, _ = w.WriteString(headings[n.Level])
		//	if n.Attributes() != nil {
		//		RenderAttributes(w, node, HeadingAttributeFilter)
		//	}
		_, _ = w.WriteString("}\n")
	}
	return ast.WalkContinue, nil
}

// // BlockquoteAttributeFilter defines attribute names which blockquote elements can have
// var BlockquoteAttributeFilter = GlobalAttributeFilter.Extend(
// 	[]byte("cite"),
// )

func (r *Renderer) renderBlockquote(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		if n.Attributes() != nil {
			_, _ = w.WriteString("\n\\begin{quotation}")
			// RenderAttributes(w, n, BlockquoteAttributeFilter)
			_ = w.WriteByte('\n')
		} else {
			_, _ = w.WriteString("\n\\begin{quotation}\n")
		}
	} else {
		_, _ = w.WriteString("\n\\end{quotation}\n")
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderCodeBlock(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString("\n\\begin{verbatim}\n") // FIXME(sbinet): handle languages w/ lstlisting
		r.writeLines(w, source, n)
	} else {
		_, _ = w.WriteString("\n\\end{verbatim}\n")
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderFencedCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.FencedCodeBlock)
	if entering {
		_, _ = w.WriteString("\n\\begin{minted}")
		lang := n.Language(source)
		if lang == nil {
			lang = []byte("text")
		}
		_, _ = w.WriteString("{")
		r.w.Write(w, lang)
		_, _ = w.WriteString("}\n")
		r.writeLines(w, source, n)
	} else {
		_, _ = w.WriteString("\\end{minted}\n")
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderHTMLBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.HTMLBlock)
	if entering {
		_, _ = w.WriteString("\n\\begin{verbatim}\n") // FIXME(sbinet)
		l := n.Lines().Len()
		for i := 0; i < l; i++ {
			line := n.Lines().At(i)
			_, _ = w.Write(line.Value(source))
		}
	} else {
		if n.HasClosure() {
			closure := n.ClosureLine
			_, _ = w.Write(closure.Value(source))
		}
		_, _ = w.WriteString("\n\\end{verbatim}\n")
	}
	return ast.WalkContinue, nil
}

// // ListAttributeFilter defines attribute names which list elements can have.
// var ListAttributeFilter = GlobalAttributeFilter.Extend(
// 	[]byte("start"),
// 	[]byte("reversed"),
// 	[]byte("type"),
// )

func (r *Renderer) renderList(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.List)
	tag := "itemize"
	if n.IsOrdered() {
		tag = "enumerate"
	}
	if entering {
		_, _ = w.WriteString("\n\\begin{")
		_, _ = w.WriteString(tag)
		_, _ = w.WriteString("}\n")
		//	if n.Attributes() != nil {
		//		RenderAttributes(w, n, ListAttributeFilter)
		//	}
	} else {
		_, _ = w.WriteString("\\end{")
		_, _ = w.WriteString(tag)
		_, _ = w.WriteString("}\n")
	}
	return ast.WalkContinue, nil
}

// // ListItemAttributeFilter defines attribute names which list item elements can have.
// var ListItemAttributeFilter = GlobalAttributeFilter.Extend(
// 	[]byte("value"),
// )

func (r *Renderer) renderListItem(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString("\\item ")
		//		fc := n.FirstChild()
		//		if fc != nil {
		//			if _, ok := fc.(*ast.TextBlock); !ok {
		//				_ = w.WriteByte('\n')
		//			}
		//		}
	}
	return ast.WalkContinue, nil
}

// // ParagraphAttributeFilter defines attribute names which paragraph elements can have.
// var ParagraphAttributeFilter = GlobalAttributeFilter

func (r *Renderer) renderParagraph(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		_, _ = w.WriteString("\n\n")
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderTextBlock(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		if n.FirstChild() != nil {
			_ = w.WriteByte('\n')
		}
	}
	return ast.WalkContinue, nil
}

// // ThematicAttributeFilter defines attribute names which hr elements can have.
// var ThematicAttributeFilter = GlobalAttributeFilter.Extend(
// 	[]byte("align"),   // [Deprecated]
// 	[]byte("color"),   // [Not Standardized]
// 	[]byte("noshade"), // [Deprecated]
// 	[]byte("size"),    // [Deprecated]
// 	[]byte("width"),   // [Deprecated]
// )

func (r *Renderer) renderThematicBreak(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	_, _ = w.WriteString("\n\\vspace{1em}\n\\hrule\n\\vspace{1em}")
	//	if n.Attributes() != nil {
	//		RenderAttributes(w, n, ThematicAttributeFilter)
	//	}
	_, _ = w.WriteString("\n")
	return ast.WalkContinue, nil
}

//// LinkAttributeFilter defines attribute names which link elements can have.
//var LinkAttributeFilter = GlobalAttributeFilter.Extend(
//	[]byte("download"),
//	// []byte("href"),
//	[]byte("hreflang"),
//	[]byte("media"),
//	[]byte("ping"),
//	[]byte("referrerpolicy"),
//	[]byte("rel"),
//	[]byte("shape"),
//	[]byte("target"),
//)

func (r *Renderer) renderAutoLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.AutoLink)
	if !entering {
		_ = w.WriteByte('}')
		return ast.WalkContinue, nil
	}
	_, _ = w.WriteString("\\colhref{")
	url := n.URL(source)
	if n.AutoLinkType == ast.AutoLinkEmail && !bytes.HasPrefix(bytes.ToLower(url), []byte("mailto:")) {
		_, _ = w.WriteString("mailto:")
	}
	_, _ = w.Write(util.EscapeHTML(util.URLEscape(url, false)))
	_, _ = w.WriteString("}{")
	return ast.WalkContinue, nil
}

// // CodeAttributeFilter defines attribute names which code elements can have.
// var CodeAttributeFilter = GlobalAttributeFilter

func (r *Renderer) renderCodeSpan(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString("\\texttt{")
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			segment := c.(*ast.Text).Segment
			value := segment.Value(source)
			if bytes.HasSuffix(value, []byte("\n")) {
				r.w.RawWrite(w, value[:len(value)-1])
				if c != n.LastChild() {
					r.w.RawWrite(w, []byte(" "))
				}
			} else {
				r.w.RawWrite(w, value)
			}
		}
		return ast.WalkSkipChildren, nil
	}
	_, _ = w.WriteString("}")
	return ast.WalkContinue, nil
}

// // EmphasisAttributeFilter defines attribute names which emphasis elements can have.
// var EmphasisAttributeFilter = GlobalAttributeFilter

func (r *Renderer) renderEmphasis(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Emphasis)
	if entering {
		tag := `\emph{`
		if n.Level == 2 {
			tag = `\textbf{`
		}
		_, _ = w.WriteString(tag)
		//	if n.Attributes() != nil {
		//		RenderAttributes(w, n, EmphasisAttributeFilter)
		//	}
	} else {
		_, _ = w.WriteString("}")
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Link)
	if entering {
		_, _ = w.WriteString("\\colhref{")
		_, _ = w.Write(util.EscapeHTML(util.URLEscape(n.Destination, true)))
		_, _ = w.WriteString("}{\\texttt{")
	} else {
		_, _ = w.WriteString("}}")
	}
	return ast.WalkContinue, nil
}

// // ImageAttributeFilter defines attribute names which image elements can have.
// var ImageAttributeFilter = GlobalAttributeFilter.Extend(
// 	[]byte("align"),
// 	[]byte("border"),
// 	[]byte("crossorigin"),
// 	[]byte("decoding"),
// 	[]byte("height"),
// 	[]byte("importance"),
// 	[]byte("intrinsicsize"),
// 	[]byte("ismap"),
// 	[]byte("loading"),
// 	[]byte("referrerpolicy"),
// 	[]byte("sizes"),
// 	[]byte("srcset"),
// 	[]byte("usemap"),
// 	[]byte("width"),
// )

var imageFilter = util.NewBytesFilter(
	[]byte("height"),
	[]byte("width"),
)

func (r *Renderer) renderImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Image)
	_, _ = w.WriteString("\\begin{figure}[h]\n")
	_, _ = w.WriteString("\\begin{center}\n")
	_, _ = w.WriteString("\\includegraphics[")
	switch attrs := n.Attributes(); attrs {
	case nil:
		width, height, err := inferDims(string(n.Destination))
		if err != nil {
			return ast.WalkStop, err
		}
		// FIXME(sbinet): shouldn't this be inches 'in' instead of 'cm' ?
		// FIXME(sbinet): shouldn't we use a floating point value ?
		_, _ = w.WriteString(fmt.Sprintf("width=%dcm,", int(float64(width)/float64(r.dpi))))
		_, _ = w.WriteString(fmt.Sprintf("height=%dcm", int(float64(height)/float64(r.dpi))))
	default:
		nn := 0
		for _, attr := range attrs {
			if !imageFilter.Contains(attr.Name) {
				if !bytes.HasPrefix(attr.Name, dataPrefix) {
					continue
				}
			}
			if nn > 0 {
				_ = w.WriteByte(',')
			}
			_, _ = w.Write(attr.Name)
			_, _ = w.WriteString(`=`)
			// TODO: convert numeric values to strings
			_, _ = w.Write(util.EscapeHTML(attr.Value.([]byte)))
			nn++
		}
	}
	_, _ = w.WriteString("]{")
	_, _ = w.Write(util.EscapeHTML(util.URLEscape(n.Destination, true)))
	//	if n.Attributes() != nil {
	//		RenderAttributes(w, n, ImageAttributeFilter)
	//	}
	_, _ = w.WriteString("}\n")
	_, _ = w.WriteString("\\end{center}\n")
	_, _ = w.WriteString("\\end{figure}\n")
	return ast.WalkSkipChildren, nil
}

func (r *Renderer) renderRawHTML(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkSkipChildren, nil
	}
	n := node.(*ast.RawHTML)
	l := n.Segments.Len()
	for i := 0; i < l; i++ {
		segment := n.Segments.At(i)
		_, _ = w.Write(segment.Value(source))
	}
	return ast.WalkSkipChildren, nil
}

func (r *Renderer) renderText(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Text)
	segment := n.Segment
	if n.IsRaw() {
		r.w.RawWrite(w, segment.Value(source))
	} else {
		r.w.Write(w, escapeLaTeX(segment.Value(source)))
		if n.SoftLineBreak() {
			_ = w.WriteByte('\n')
		}
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderString(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.String)
	if n.IsCode() {
		_, _ = w.Write(n.Value)
	} else {
		if n.IsRaw() {
			r.w.RawWrite(w, n.Value)
		} else {
			r.w.Write(w, n.Value)
		}
	}
	return ast.WalkContinue, nil
}

var dataPrefix = []byte("data-")

// RenderAttributes renders given node's attributes.
// You can specify attribute names to render by the filter.
// If filter is nil, RenderAttributes renders all attributes.
func RenderAttributes(w util.BufWriter, node ast.Node, filter util.BytesFilter) {
	for _, attr := range node.Attributes() {
		if filter != nil && !filter.Contains(attr.Name) {
			if !bytes.HasPrefix(attr.Name, dataPrefix) {
				continue
			}
		}
		_, _ = w.WriteString(" ")
		_, _ = w.Write(attr.Name)
		_, _ = w.WriteString(`="`)
		// TODO: convert numeric values to strings
		_, _ = w.Write(util.EscapeHTML(attr.Value.([]byte)))
		_ = w.WriteByte('"')
	}
}

// A Writer interface writes textual contents to a writer.
type Writer interface {
	// Write writes the given source to writer with resolving references and unescaping
	// backslash escaped characters.
	Write(writer util.BufWriter, source []byte)

	// RawWrite writes the given source to writer without resolving references and
	// unescaping backslash escaped characters.
	RawWrite(writer util.BufWriter, source []byte)
}

type defaultWriter struct {
}

func escapeRune(writer util.BufWriter, r rune) {
	if r < 256 {
		v := util.EscapeHTMLByte(byte(r))
		if v != nil {
			_, _ = writer.Write(v)
			return
		}
	}
	_, _ = writer.WriteRune(util.ToValidRune(r))
}

func (d *defaultWriter) RawWrite(writer util.BufWriter, source []byte) {
	n := 0
	l := len(source)
	for i := 0; i < l; i++ {
		v := util.EscapeHTMLByte(source[i])
		if v != nil {
			_, _ = writer.Write(source[i-n : i])
			n = 0
			_, _ = writer.Write(v)
			continue
		}
		n++
	}
	if n != 0 {
		_, _ = writer.Write(source[l-n:])
	}
}

func (d *defaultWriter) Write(writer util.BufWriter, source []byte) {
	escaped := false
	var ok bool
	limit := len(source)
	n := 0
	for i := 0; i < limit; i++ {
		c := source[i]
		if escaped {
			if util.IsPunct(c) {
				d.RawWrite(writer, source[n:i-1])
				n = i
				escaped = false
				continue
			}
		}
		if c == '&' {
			pos := i
			next := i + 1
			if next < limit && source[next] == '#' {
				nnext := next + 1
				if nnext < limit {
					nc := source[nnext]
					// code point like #x22;
					if nnext < limit && nc == 'x' || nc == 'X' {
						start := nnext + 1
						i, ok = util.ReadWhile(source, [2]int{start, limit}, util.IsHexDecimal)
						if ok && i < limit && source[i] == ';' {
							v, _ := strconv.ParseUint(util.BytesToReadOnlyString(source[start:i]), 16, 32)
							d.RawWrite(writer, source[n:pos])
							n = i + 1
							escapeRune(writer, rune(v))
							continue
						}
						// code point like #1234;
					} else if nc >= '0' && nc <= '9' {
						start := nnext
						i, ok = util.ReadWhile(source, [2]int{start, limit}, util.IsNumeric)
						if ok && i < limit && i-start < 8 && source[i] == ';' {
							v, _ := strconv.ParseUint(util.BytesToReadOnlyString(source[start:i]), 0, 32)
							d.RawWrite(writer, source[n:pos])
							n = i + 1
							escapeRune(writer, rune(v))
							continue
						}
					}
				}
			} else {
				start := next
				i, ok = util.ReadWhile(source, [2]int{start, limit}, util.IsAlphaNumeric)
				// entity reference
				if ok && i < limit && source[i] == ';' {
					name := util.BytesToReadOnlyString(source[start:i])
					entity, ok := util.LookUpHTML5EntityByName(name)
					if ok {
						d.RawWrite(writer, source[n:pos])
						n = i + 1
						d.RawWrite(writer, entity.Characters)
						continue
					}
				}
			}
			i = next - 1
		}
		if c == '\\' {
			escaped = true
			continue
		}
		escaped = false
	}
	d.RawWrite(writer, source[n:])
}

// DefaultWriter is a default implementation of the Writer.
var DefaultWriter = &defaultWriter{}

var bDataImage = []byte("data:image/")
var bPng = []byte("png;")
var bGif = []byte("gif;")
var bJpeg = []byte("jpeg;")
var bWebp = []byte("webp;")
var bJs = []byte("javascript:")
var bVb = []byte("vbscript:")
var bFile = []byte("file:")
var bData = []byte("data:")

// IsDangerousURL returns true if the given url seems a potentially dangerous url,
// otherwise false.
func IsDangerousURL(url []byte) bool {
	if bytes.HasPrefix(url, bDataImage) && len(url) >= 11 {
		v := url[11:]
		if bytes.HasPrefix(v, bPng) || bytes.HasPrefix(v, bGif) ||
			bytes.HasPrefix(v, bJpeg) || bytes.HasPrefix(v, bWebp) {
			return false
		}
		return true
	}
	return bytes.HasPrefix(url, bJs) || bytes.HasPrefix(url, bVb) ||
		bytes.HasPrefix(url, bFile) || bytes.HasPrefix(url, bData)
}
