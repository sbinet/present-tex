// Copyright 2015 The present-tex Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Command present-tex generates LaTeX/Beamer slides from present.
//
// Usage of present-tex:
//
//   $ present-tex [options] [input-file [output.tex]]
//
// Examples:
//   $ present-tex input.slide > out.tex
//   $ present-tex input.slide out.tex
//   $ present-tex < input.slide > out.tex
//
// Options:
//   -base="": base path for slide templates
package main

import (
	"bytes"
	"flag"
	"fmt"
	"html"
	"html/template"
	"io"
	"io/fs"
	"log"
	"os"
	"strings"

	"golang.org/x/tools/present"
)

var (
	tmpl        *template.Template // beamer template
	hasCode     = false            // whether the .slide has a .code or .play directive
	beamerTheme = flag.String("beamer-theme", "default", "Beamer theme to use (e.g: Berkeley, Madrid, ...)")
	dpi         = flag.Int("dpi", 72, "DPI resolution to use for PDF")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `%[1]s - generates LaTeX/Beamer slides from present.

Usage of %[1]s:

$ %[1]s [options] [input-file [output.tex]]

Examples:

$ %[1]s input.slide > out.tex
$ %[1]s input.slide out.tex
$ %[1]s < input.slide > out.tex

Options:
`,
			os.Args[0],
		)
		flag.PrintDefaults()
	}

	log.SetFlags(0)
	log.SetPrefix("present-tex: ")

	tmpldirFlag := flag.String("base", "", "base path for slide templates")

	flag.Parse()

	var tmpldir = func() fs.FS {
		o, err := fs.Sub(tmplFS, "templates")
		if err != nil {
			log.Fatalf("could not locate embedded 'templates' directory: %+v", err)
		}
		return o
	}()

	if *tmpldirFlag != "" {
		tmpldir = os.DirFS(*tmpldirFlag)
	}

	var (
		r      io.Reader
		w      io.Writer
		input  = "stdin"
		output = "stdout"
	)

	switch flag.NArg() {
	case 0:
		r = os.Stdin
		w = os.Stdout
	case 1:
		input = flag.Arg(0)
		f, err := os.Open(input)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		log.Printf("input:  [%s]...\n", input)

		r = f
		w = os.Stdout

	case 2:

		input = flag.Arg(0)
		f, err := os.Open(input)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		log.Printf("input:  [%s]...\n", input)

		output = flag.Arg(1)
		log.Printf("output: [%s]...\n", output)

		tex, err := os.Create(output)
		if err != nil {
			log.Fatalf("could not create output file [%s]: %v\n", output, err)
		}
		defer func() {
			err = tex.Close()
			if err != nil {
				log.Fatalf("could not close output file [%s]: %v\n", output, err)
			}
		}()

		r = f
		w = tex

	default:
		flag.Usage()
		os.Exit(2)
	}

	err := xmain(w, r, input, tmpldir)
	if err != nil {
		log.Fatalf("could not run present-tex: %+v", err)
	}
}

func xmain(w io.Writer, r io.Reader, input string, tmpldir fs.FS) error {
	ctx := present.Context{
		ReadFile: os.ReadFile,
		Render:   renderAsLaTeX,
	}

	doc, err := ctx.Parse(r, input, 0)
	if err != nil {
		return fmt.Errorf("could not parse input document: %w", err)
	}

	err = parseImages(doc)
	if err != nil {
		return fmt.Errorf("could not parse images: %w", err)
	}

	err = parseCode(doc)
	if err != nil {
		return fmt.Errorf("could not parse code fragments: %w", err)
	}

	tmpl, err = initTemplates(tmpldir)
	if err != nil {
		return fmt.Errorf("could not parse templates: %w", err)
	}

	buf := new(bytes.Buffer)
	err = doc.Render(buf, tmpl)
	if err != nil {
		return fmt.Errorf("could not render document: %w", err)
	}

	out := []byte(html.UnescapeString(buf.String()))

	_, err = w.Write(out)
	if err != nil {
		return fmt.Errorf("could not fill output: %w", err)
	}

	return nil
}

func initTemplates(root fs.FS) (*template.Template, error) {
	tmpl := template.New("").Funcs(funcs).Delims("<<", ">>")
	_, err := tmpl.ParseFS(root, "beamer.tmpl")
	if err != nil {
		return nil, err
	}

	return tmpl, err
}

// renderElem implements the elem template function, used to render
// sub-templates.
func renderElem(t *template.Template, e present.Elem) (template.HTML, error) {
	var data interface{} = e
	if s, ok := e.(present.Section); ok {
		data = struct {
			present.Section
			Template *template.Template
		}{s, t}
	}
	return execTemplate(t, e.TemplateName(), data)
}

var (
	funcs = template.FuncMap{}
)

func init() {
	funcs["elem"] = renderElem
	funcs["stringFromBytes"] = func(raw []byte) string { return string(raw) }
	funcs["join"] = func(lines []string) string { return strings.Join(lines, "\n") }
	funcs["nodot"] = func(s string) string {
		if strings.HasPrefix(s, ".") {
			return s[1:]
		}
		return s
	}
	tex1 := strings.NewReplacer(
		"&", `\&`,
		"$", `\$`,
		"^", `\^{}`,
		"%", `\%`,
		"~", `$\sim$`,
		"#", `\#`,
		"{", `\{`,
		"}", `\}`,
		`\`, `\textbackslash`,
		"é", `\'e`,
		"è", `\`+"`e",
		"à", `\`+"`a",
		"ù", `\`+"`u",
		"â", `\^a`,
		"ê", `\^e`,
		"î", `\^i`,
		"ô", `\^o`,
		"û", `\^u`,
		"ä", `\"a`,
		"ë", `\"e`,
		"ï", `\"i`,
		"ü", `\"u`,
		"ÿ", `\"y`,
		"ç", `\c{c}`,
		"Ç", `\c{C}`,
		"œ", `\oe `,
		">", `$>$`,
		"<", `$<$`,
		">=", `$\geq$`,
		"<=", `$\leq$`,
		"->", `$\rightarrow$`,
		"=>", `$\Rightarrow$`,
	)
	tex2 := strings.NewReplacer(
		"_", `\_`,
	)

	style := func(s string) string {
		s = tex1.Replace(s)
		s = string(renderStyle(s))
		s = tex2.Replace(s)
		return s
	}
	funcs["style"] = style

	funcs["beamerTheme"] = func() string {
		return *beamerTheme
	}

	funcs["hasCode"] = func() bool {
		return hasCode
	}

	funcs["pdfAuthor"] = func(authors []present.Author) string {
		out := make([]string, 0, len(authors))
		for _, a := range authors {
			name, _, _ := parseAuthor(a)
			if name == "" {
				continue
			}
			out = append(out, fmt.Sprintf("pdfauthor={%s},%%\n", style(name)))
		}
		return strings.Join(out, "  ")
	}

	funcs["texAuthor"] = func(authors []present.Author) string {
		const hdr = "\\parbox{0.26\\textwidth}{\n\t\\texorpdfstring\n\t  {\n\t\t\\centering\n"
		out := make([]string, 0, len(authors))
		shorts := make([]string, 0, len(authors))
		for _, a := range authors {
			elems := renderAuthor(a)
			if len(elems) == 0 {
				continue
			}
			name := elems[0]
			if name == "" {
				continue
			}
			if len(shorts) > 0 {
				shorts = append(shorts, "\\&")
			}
			shorts = append(shorts, name)
			if len(out) > 0 {
				out = append(out, "\\and %\n")
			}
			out = append(out, hdr)
			for _, elem := range elems {
				out = append(out, "\t\t"+elem+` \\`+"\n")
			}
			out = append(out, "\t  }\n\t{"+elems[0]+"}\n}\n")
		}
		if len(out) > 0 {
			out = append([]string{"\\author[" + strings.Join(shorts, " ") + "]{\n"}, out...)
			out = append(out, "}\n")
		}
		return strings.Join(out, " ")
	}
}

func renderAuthor(author present.Author) []string {
	var elems []string
	if len(author.Elem) == 0 {
		return elems
	}
	for _, e := range author.Elem {
		str, err := renderElem(tmpl, e)
		if err != nil {
			log.Fatal(err)
		}
		elem := html.UnescapeString(string(str))
		elem = strings.Trim(elem, "\n")
		elems = append(elems, elem)
	}
	return elems
}

func parseAuthor(author present.Author) (name string, inst string, mail present.Link) {
	elems := author.TextElem()
	if len(elems) == 0 {
		return
	}
	getLines := func(i int) []string {
		lines := elems[i].(present.Text).Lines
		return lines
	}

	name = strings.TrimSpace(getLines(0)[0])
	if name == "" {
		return
	}

	if len(elems) > 1 {
		inst = strings.TrimSpace(getLines(1)[0])
	}
	for _, elem := range author.Elem {
		link, ok := elem.(present.Link)
		if !ok {
			continue
		}
		if strings.Contains(link.Label, "@") {
			mail = link
		}
	}
	return
}

// execTemplate is a helper to execute a template and return the output as a
// template.HTML value.
func execTemplate(t *template.Template, name string, data interface{}) (template.HTML, error) {
	b := new(bytes.Buffer)
	err := t.ExecuteTemplate(b, name, data)
	if err != nil {
		return "", err
	}
	return template.HTML(b.String()), nil
}
