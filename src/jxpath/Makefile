#
# Convenience makefile for go development
#
all: clean deploy regress

deploy:
	go install
run: # just check it's built
	../../bin/jxpath --help; : 

race:
	go test -race

clean:
	-go clean -i
	-rm ../../pkg/linux_amd64/*.a
	-rm ../../bin/jxpath

regress:
	../../bin/jxpath /timelord /galaxy <universe.xml

# Stretch goal test: get this to work: requires attributes
# jxpath /html/body/div/div[7]/div[2]/script <smileybarry.xml 

bench:
	go test -bench BenchmarkDrWho # 2>/dev/null

bencht:
	go test -bench BenchmarkLexer #2>/dev/null

NODES=15
PROFILE=-cpuprofile
#PROFILE=-blockprofile
#PROFILE=-memprofile
profile:
	go test -bench BenchmarkEndToEnd -run=NONE ${PROFILE}=profile.out
	go tool pprof -text -nodecount=${NODES} jxpath.test profile.out


COVER=cover.out
cover:
	go test -run=Test[JX].* -coverprofile=${COVER}
	go tool cover -html=${COVER}
