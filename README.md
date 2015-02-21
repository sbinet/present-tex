present-tex
===========

`present-tex` is a simple command to create a `LaTeX/Beamer` presentation from a [golang.org/x/tools/cmd/present](present) slide deck.

## Installation

```go
$ go get github.com/sbinet/present-tex
```

## Example

```sh
$ present-tex my.slide > my.tex
$ pdflatex my.tex
```

