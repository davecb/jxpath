DIRS=./src/pathExpr ./src/xml ./src/json ./src/trace \
     ./src/lexer ./src/token ./src/jxpath
FILES=${shell find ${DIRS} -type f  | egrep -v 'RCS|.iml|.idea'}

all:
	@echo "ci, co or rcsout, here"
	#echo ${FILES}


cil:
	ci -l ${FILES}

ci:
	ci -u ${FILES}

co:
	co -l ${FILES}

out: # rcsout
	-for i in ${DIRS}; do \
		echo -n "$$i: " ; \
		(cd $$i; rcsout); \
	done; \
	echo ""

rcsl:
	rcs -l ${FILES}

diff:
	rcsdiff ${FILES}

log: # report log message in use
	cat src/jxpath/RCS/.main.go.checkin-comment
