.PHONY: all

all: test.slide _code/hello.go
	present-tex -beamer-theme=Madrid ./test.slide ./test.tex
	pdflatex -shell-escape test.tex
	pdflatex -shell-escape test.tex
	@/bin/rm test.aux test.nav test.out test.snm test.toc test.vrb test.log
