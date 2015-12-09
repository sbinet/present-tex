.PHONY: all

all: test.slide _code/hello.go
	present-tex ./test.slide >| ./test.tex
	pdflatex -shell-escape test.tex


