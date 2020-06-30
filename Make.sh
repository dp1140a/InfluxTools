#!/usr/bin/env bash

# Go parameters
GOCMD=go
GOBUILD="$GOCMD build"
GOCLEAN="$GOCMD clean"
GOTEST="$GOCMD test"
GOGET="$GOCMD get"
BUILD_DIR="build"

LD_FLAGS="-s -w"
GOOS=$(uname)
GOARCH="amd64"

function build(){
	echo "Building $1 for $GOOS-$GOARCH"
  export GOOS=$GOOS
	BINARY_NAME=$1_$(echo $GOOS | awk '{print tolower($0)}')
	SRC="./cmd/$1"
	echo $BINARY_NAME
  $($GOBUILD -o $BUILD_DIR/$BINARY_NAME -ldflags="$LD_FLAGS" $SRC)
  echo "Done Building $1"
}

function clean(){
    echo "Cleaning"
    rm -f $BUILD_DIR/*

}

function all(){
    clean
    build clusterInfo
    build rebalancer
    build underReplicated
    echo "All binaries are in the $BUILD_DIR folder."
    echo "Done"
}

function dist(){
	all
	cp -R profiles/. build
	cp -R diagnostics/. build
}

function usage(){
  echo "Make.sh will build or install InfluxTools"
  echo "Usage: $0 [tool] [-p platform]"
  echo " "
  echo "Tool:"
  echo "  all       		build all tools"
  echo "  rebalancer     	builds the rebalancer tool"
  echo "  clusterInfo     	builds the clusterInfo tool"
  echo "  underReplicated   builds the underreplication tool"
  echo "  diagnostics		builds the diagnostics tools"
  echo "  usage|help    displays what you are reading right now"
  echo " "
  echo "Parameters:"
  echo "  -p      Platform to compile for (darwin|linux|windows|all). Note: all tools will only be compiled for the amd64 (64bit) architecture"
  echo " "
  echo "TODOS:"
  exit 0
}

#######################
###		MAIN		###
#######################
# Check if GOlang is installed
if ! [ -x "$(command -v go)" ]; then
  echo 'Error: go is not installed. Please install Golang and run again' >&2
  exit 1
fi

subcommand=$1
shift
while getopts ":p:" opt; do
  case ${opt} in
    p)
      PLATFORM=$OPTARG
      if [[ "$PLATFORM" =~ ^(darwin|linux|windows)$ ]];
      then
        echo -e "Platform set to $PLATFORM"
        GOOS=$PLATFORM
      else
        echo -e "$PLATFORM is not a valid platform option (darwin|linux|windows)"
        exit 1
      fi
      ;;
    \?)
      echo "Invalid Option: $OPTARG" 1>&2
      usage
      ;;
    :)
      echo "Invalid option: $OPTARG requires an argument" 1>&2
      exit 1
  esac
done
shift $((OPTIND -1))

case $subcommand in
  all)
    all
    ;;

  clean)
    clean
    ;;

  clusterInfo)
    build clusterInfo
    ;;

  rebalancer)
    build rebalancer
    ;;

  underReplicated)
    build underReplicated
    ;;

  *)
    usage
    ;;
esac
