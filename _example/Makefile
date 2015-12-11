.PHONY: all

all: test.slide _code/hello.go
	present-tex ./test.slide ./test.tex
	pdflatex -shell-escape test.tex
	@/bin/rm test.aux test.nav test.out test.snm test.toc test.vrb
