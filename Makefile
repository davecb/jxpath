DIRS=./src/pathExpr ./src/xml ./src/json ./src/trace ./src/jxpath
FILES=${shell find ${DIRS} -type f  | grep -v 'RCS'}

all:
	@echo "ci, co or rcsout, here"
	# echo ${FILES}


cil:
	ci -l ${FILES}

ci:
	ci -u ${FILES}

co:
	co -l ${FILES}

rcsout:
	for i in ${DIRS}; do \
		(cd $$i; rcsout); \
	done

rcsl:
	rcs -l ${FILES}

diff:
	rcsdiff ${FILES}
