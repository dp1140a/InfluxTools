#!/usr/bin/env bash

INFLUXDB_USER="foo"
INFLUXDB_PASSWORD="bar"
INFLUXDB_HOST="localhost"
INFLUXDB_PORT="8086"
BASE_NAME=$(date '+%Y%m%d-%H%M%S')
OUT_DIR="output"

while [[ $# -gt 1 ]]
do
key="$1"

case $key in
    -u|--username)
    INFLUXDB_USER="$2"
    shift # past argument
    ;;

    -p|--password)
    INFLUXDB_PASSWORD="$2"
    shift # past argument
    ;;

     -h|--host)
    INFLUXDB_HOST="$2"
    shift # past argument
    ;;

     -x|--port)
    INFLUXDB_PORT="$2"
    shift # past argument
    ;;

    -o|--outdir)
    OUT_DIR="$2"
    shift # past argument
    ;;

    *)
            # unknown option
    ;;
esac
shift # past argument or value
done

echo -e $OUT_DIR
if [ ! -d "$OUT_DIR" ]; then
  mkdir $OUT_DIR
else
  rm -f $OUT_DIR/*.txt
fi

source countsSeriesAndMeasurements.sh -u $INFLUXDB_USER -p $INFLUXDB_PASSWORD -o $OUT_DIR -h $INFLUXDB_HOST -x $INFLUXDB_PORT
source hintedHandoffQueueSize.sh -u $INFLUXDB_USER -p $INFLUXDB_PASSWORD -o $OUT_DIR -h $INFLUXDB_HOST -x $INFLUXDB_PORT
source hintedHandoffQueueThroughput.sh -u $INFLUXDB_USER -p $INFLUXDB_PASSWORD -o $OUT_DIR -h $INFLUXDB_HOST -x $INFLUXDB_PORT
source readWriteResponse.sh -u $INFLUXDB_USER -p $INFLUXDB_PASSWORD -o $OUT_DIR -h $INFLUXDB_HOST -x $INFLUXDB_PORT
source shardSizeOnDisk.sh -u $INFLUXDB_USER -p $INFLUXDB_PASSWORD -o $OUT_DIR -h $INFLUXDB_HOST -x $INFLUXDB_PORT
source writesHttpReqPerMinute.sh -u $INFLUXDB_USER -p $INFLUXDB_PASSWORD -o $OUT_DIR -h $INFLUXDB_HOST -x $INFLUXDB_PORT
source writesPointsPerMInute.sh -u $INFLUXDB_USER -p $INFLUXDB_PASSWORD -o $OUT_DIR -h $INFLUXDB_HOST -x $INFLUXDB_PORT

tar -czf ${BASE_NAME}.tar.gz $OUT_DIR

echo -e "Done.  Please send the tarball, along with your logs, to InfluxData Support."
