all: format

format:
	latexindent main.tex -m -y="modifyLineBreaks:removeParagraphLineBreaks:all:1" > main.tmp.tex
	mv main.tmp.tex main.tex
	latexindent main.tex -m -y="modifyLineBreaks:textWrapOptions:columns:80" > main.tmp.tex
	mv main.tmp.tex main.tex

.PHONY: all format
