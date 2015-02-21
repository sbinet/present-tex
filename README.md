present-tex
===========

`present-tex` is a simple command to create a `LaTeX/Beamer` presentation from a [present](https:///golang.org/x/tools/cmd/present) slide deck.

## Installation

```sh
$ go get github.com/sbinet/present-tex
```

## Example

```sh
$ present-tex my.slide > my.tex
$ pdflatex my.tex
```

