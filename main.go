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
	"go/build"
	"html/template"
	"io"
	"log"
	"os"
	"path"
	"strings"

	"golang.org/x/tools/present"
)

const (
	basePkg         = "github.com/sbinet/present-tex"
	basePathMessage = `
By default, present-tex locates the slide template files and associated
static content by looking for a %q package
in your Go workspaces (GOPATH).
You may use the -base flag to specify an alternate location.
`
)

func printf(format string, args ...interface{}) (int, error) {
	return fmt.Fprintf(os.Stderr, format, args...)
}

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

	tmpldir := ""
	flag.StringVar(&tmpldir, "base", "", "base path for slide templates")

	flag.Parse()

	if tmpldir == "" {
		p, err := build.Default.Import(basePkg, "", build.FindOnly)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't find present-tex files: %v\n", err)
			fmt.Fprintf(os.Stderr, basePathMessage, basePkg)
			os.Exit(1)
		}
		tmpldir = path.Join(p.Dir, "templates")
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
		printf("input:  [%s]...\n", input)

		r = f
		w = os.Stdout

	case 2:

		input = flag.Arg(0)
		f, err := os.Open(input)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		printf("input:  [%s]...\n", input)

		output = flag.Arg(1)
		printf("output: [%s]...\n", output)

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

	doc, err := present.Parse(r, input, 0)
	if err != nil {
		log.Fatal(err)
	}

	tmpl, err := initTemplates(tmpldir)
	if err != nil {
		log.Fatal(err)
	}

	buf := new(bytes.Buffer)
	err = doc.Render(buf, tmpl)
	if err != nil {
		log.Fatal(err)
	}

	out := unescapeHTML(buf.Bytes())

	_, err = w.Write(out)
	if err != nil {
		log.Fatalf("could not fill output: %v\n", err)
	}
}

func unescapeHTML(data []byte) []byte {
	out := make([]byte, len(data))
	copy(out, data)
	for _, r := range []struct {
		old string
		new string
	}{
		{
			old: "&lt;",
			new: "<",
		},
		{
			old: "&gt;",
			new: ">",
		},
		{
			old: "&#43;",
			new: "+",
		},
		{
			old: "&#34;",
			new: `"`,
		},
		{
			old: "&#39;",
			new: "'",
		},
		{
			old: "&quot;",
			new: `"`,
		},
		{
			old: "&amp;",
			new: "&",
		},
		{
			old: "&nbsp;",
			new: " ",
		},
	} {
		out = bytes.Replace(out, []byte(r.old), []byte(r.new), -1)
	}
	return out
}

func initTemplates(base string) (*template.Template, error) {
	fname := path.Join(base, "beamer.tmpl")
	tmpl := template.New("").Funcs(funcs).Delims("<<", ">>")
	_, err := tmpl.ParseFiles(fname)
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
		"~", `\~{}`,
		"#", `\#`,
		"{", `\{`,
		"}", `\}`,
		`\`, `\textbackslash`,
	)
	tex2 := strings.NewReplacer(
		"_", `\_`,
	)

	funcs["style"] = func(s string) string {
		s = tex1.Replace(s)
		s = string(renderStyle(s))
		s = tex2.Replace(s)
		return s
	}

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
