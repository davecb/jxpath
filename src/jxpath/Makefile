#
# Convenience makefile for go development
#
all: clean deploy regress

deploy:
	go install
run: # just check it's built
	../../bin/jxpath --help; : 

test:
	go test

clean:
	-go clean -i
	-rm ../../pkg/linux_amd64/*.a
	-rm ../../bin/jxpath

regress:
	../../bin/jxpath /timelord /galaxy <universe.xml

# Stress test: get this to work
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
	go test -bench BenchmarkDrWho -run=NONE ${PROFILE}=profile.out
	go tool pprof -text -nodecount=${NODES} jxpath.test profile.out
