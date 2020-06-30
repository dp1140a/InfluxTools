#!/bin/bash
# hintedHandoffQueueThroughput.sh

# What: Hinted-handoff queue throughput per minute size on disk from the last 24 hours, grouped by hour, and all tags.
# Why: This will tell us how fast (or slow) the hinted-handoff queue is draining (or filling up)
echo -e "**************************************"
echo -e hintedHandoffQueueThroughput.sh
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
# echo $INFLUXDB_USER : $INFLUXDB_PASSWORD
influx \
	-host $INFLUXDB_HOST \
	-port $INFLUXDB_PORT \
	-username $INFLUXDB_USER \
	-password $INFLUXDB_PASSWORD \
	-database '_internal' \
	-execute 'SELECT derivative(max("queueBytes"), 60s) FROM "_internal"."monitor"."hh_processor" WHERE time > now() - 24h GROUP BY time(1h),* fill(0)' \
	> $OUT_DIR/hintedHandoffQueueThroughput.txt

echo "Done!!"
