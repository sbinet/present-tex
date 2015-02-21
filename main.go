// Copyright 2015 The present-tex Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"log"
	"os"
	"path"
	"strings"

	"golang.org/x/tools/present"
)

func printf(format string, args ...interface{}) (int, error) {
	return fmt.Fprintf(os.Stderr, format, args...)
}

func main() {
	flag.Parse()
	input := flag.Arg(0)
	output := input
	if flag.NArg() > 1 {
		output = flag.Arg(1)
	} else {
		output = input
		if strings.HasSuffix(output, ".slide") {
			output = output[:len(output)-len(".slide")] + ".pdf"
		} else {
			output += ".pdf"
		}
	}
	printf("input:  [%s]...\n", input)
	printf("output: [%s]...\n", output)

	f, err := os.Open(input)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	doc, err := present.Parse(f, input, 0)
	if err != nil {
		log.Fatal(err)
	}

	printf("doc:\ntitle: %q\nsub: %q\ntime: %v\nauthors: %v\ntags: %v\n",
		doc.Title, doc.Subtitle, doc.Time, doc.Authors,
		doc.Tags,
	)
	/*
		for _, section := range doc.Sections {
			printf("--- section %v %q---\n", section.Number, section.Title)
			for _, elem := range section.Elem {
				switch elem := elem.(type) {
				default:
					printf("%#v\n", elem)
				case present.Code:
					printf("code: %s\n", string(elem.Raw))
				}
			}
		}
	*/

	tmpl, err := initTemplates("templates")
	if err != nil {
		log.Fatal(err)
	}

	buf := new(bytes.Buffer)
	err = doc.Render(buf, tmpl)
	if err != nil {
		log.Fatal(err)
	}

	out := bytes.Replace(buf.Bytes(), []byte(`&#34;`), []byte(`"`), -1)
	os.Stdout.Write(out)
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
