#!/bin/bash
# hintedHandoffQueueSize.sh

# What: Hinted-handoff queue size from the last 24 hours, grouped by hour, and all tags.
# Why: This will tell us how large the hinted-handoff queue was for each hour, broken out by target server.
echo -e "**************************************"
echo -e batchSize.sh
echo -e "**************************************"

INFLUXDB_USER="foo"
INFLUXDB_PASSWORD="bar"
OUT_DIR="output"
INFLUXDB_HOST="localhost"
INFLUXDB_PORT="8086"

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
    -o|--outdir)
    OUT_DIR="$2"
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

    *)
            # unknown option
    ;;
esac
shift # past argument or value
done

echo -e "Running query"
influx \
	-host $INFLUXDB_HOST \
	-port $INFLUXDB_PORT \
	-username $INFLUXDB_USER \
	-password $INFLUXDB_PASSWORD \
	-database '_internal' \
	-execute 'SELECT non_negative_derivative(max("writePointsOk"), 1s) / non_negative_derivative(max("writeReqOk"), 1s) FROM "monitor"."shard" WHERE time > now()-1d GROUP BY time(1m), "database", "id" fill(0)' \
	> $OUT_DIR/batchSize.txt

echo "Done!!"
