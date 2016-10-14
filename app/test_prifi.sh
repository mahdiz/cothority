#!/usr/bin/env bash

DBG_SHOW=2
# Debug-level for app
DBG_APP=2
DBG_SRV=0
# Uncomment to build in local dir
STATICDIR=test
# Needs 5 clients
NBR=5

. lib/test/libtest.sh
. lib/test/cothorityd.sh

main(){
    startTest
    build
	test Build
	test Setup
    stopTest
}

testSetup(){
	for n in $(seq $NBR); do
		runCfg $n
	done
	testOK runPrifi 1 -nowait trustee
	testOK runPrifi 2 -nowait trustee
	testOK runPrifi 3 -nowait relay
	testOK runPrifi 4 -nowait client
	testOK runPrifi 5 -nowait client
}

testBuild(){
    testOK dbgRun ./prifi --help
}

runPrifi(){
    local D=prifi$1
    shift
    dbgRun ./prifi -d $DBG_APP -c $D/config.toml $@
}

runCfg(){
    echo -e "127.0.0.1:200$1\nprifi$1\n\n" | dbgRun runPrifi $1 setup
}

build(){
    BUILDDIR=$(pwd)
    if [ "$STATICDIR" ]; then
        DIR=$STATICDIR
    else
        DIR=$(mktemp -d)
    fi
    mkdir -p $DIR
    cd $DIR
    echo "Building in $DIR"
    for app in prifi; do
        if [ ! -e $app -o "$BUILD" ]; then
            if ! go build -o $app $BUILDDIR/$app/*.go; then
                fail "Couldn't build $app"
            fi
        fi
    done
    for n in $(seq $NBR); do
    	d=prifi$n
    	rm -rf $d
    	mkdir $d
    	local oshow=$DBG_SHOW
    	DBG_SHOW=0
    	runCfg $n
    	DBG_SHOW=$oshow
    done
}

if [ "$1" -a "$STATICDIR" ]; then
    rm -f $STATICDIR/prifi
fi

main
