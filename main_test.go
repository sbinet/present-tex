// Copyright 2021 The present-tex Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"io/fs"
	"os"
	"os/exec"
	"testing"
)

func TestConvert(t *testing.T) {
	var tmpldir = func() fs.FS {
		o, err := fs.Sub(tmplFS, "templates")
		if err != nil {
			t.Fatalf("could not locate embedded 'templates' directory: %+v", err)
		}
		return o
	}()

	err := os.Chdir("testdata")
	if err != nil {
		t.Fatalf("could not chdir to testdata: %+v", err)
	}

	for _, tc := range []struct {
		input string
		want  string
	}{
		{
			input: "talk.slide",
			want:  "talk_golden.tex",
		},
	} {
		t.Run("", func(t *testing.T) {
			r, err := os.ReadFile(tc.input)
			if err != nil {
				t.Fatalf("could not read input file %q: %+v", tc.input, err)
			}

			want, err := os.ReadFile(tc.want)
			if err != nil {
				t.Fatalf("could not read golden file %q: %+v", tc.want, err)
			}

			w := new(bytes.Buffer)
			err = xmain(w, bytes.NewReader(r), tc.input, tmpldir)
			if err != nil {
				t.Fatalf("could not process document: %+v", err)
			}

			if got := w.Bytes(); !bytes.Equal(got, want) {
				_ = os.WriteFile(tc.input+".tex", got, 0644)
				out, _ := exec.Command("diff", "-urN", tc.input+".tex", tc.want).CombinedOutput()
				t.Fatalf("output documents differ: %q:\n%s", tc.input, out)
			}
		})
	}
}
