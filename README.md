present-tex
===========

`present-tex` is a simple command to create a `LaTeX/Beamer` presentation from a [present](https:///golang.org/x/tools/cmd/present) slide deck.

## Installation

```sh
$ go get github.com/sbinet/present-tex
```

## Documentation

Available from [godoc.org](https://godoc.org/github.com/sbinet/present-tex) and from the command-line:

```sh
$ present-tex -h
present-tex - generates LaTeX/Beamer slides from present.

Usage of present-tex:

$ present-tex [input-file [output.tex]]

Examples:

$ present-tex input.slide > out.tex
$ present-tex input.slide out.tex
$ present-tex < input.slide > out.tex

```

## Example

```sh
$ present-tex my.slide > my.tex
$ pdflatex my.tex
```

