// Copyright 2015 The present-tex Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"net/url"
	"strings"
)

func renderLink(href, text string) string {
	text = renderFont(text)
	if text == "" {
		text = href
	}
	return fmt.Sprintf(`\myblue{\href{%s}{\texttt{%s}}}`, href, text)
}

// parseInlineLink parses an inline link at the start of s, and returns
// a rendered HTML link and the total length of the raw inline link.
// If no inline link is present, it returns all zeroes.
func parseInlineLink(s string) (link string, length int) {
	if !strings.HasPrefix(s, "[[") {
		return
	}
	end := strings.Index(s, "]]")
	if end == -1 {
		return
	}
	urlEnd := strings.Index(s, "]")
	rawURL := strings.Replace(s[2:urlEnd], `\&`, "&", -1)
	const badURLChars = `<>"{}|\^[] ` + "`" // per RFC2396 section 2.4.3
	if strings.ContainsAny(rawURL, badURLChars) {
		return
	}
	if urlEnd == end {
		simpleUrl := ""
		url, err := url.Parse(rawURL)
		if err == nil {
			// If the URL is http://foo.com, drop the http://
			// In other words, render [[http://golang.org]] as:
			//   <a href="http://golang.org">golang.org</a>
			if strings.HasPrefix(rawURL, url.Scheme+"://") {
				simpleUrl = strings.TrimPrefix(rawURL, url.Scheme+"://")
			} else if strings.HasPrefix(rawURL, url.Scheme+":") {
				simpleUrl = strings.TrimPrefix(rawURL, url.Scheme+":")
			}
		}
		return renderLink(rawURL, simpleUrl), end + 2
	}
	if s[urlEnd:urlEnd+2] != "][" {
		return
	}
	text := s[urlEnd+2 : end]
	return renderLink(rawURL, text), end + 2
}
