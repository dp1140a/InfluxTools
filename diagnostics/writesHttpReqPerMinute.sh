#!/bin/bash
# writesHttpReqPerMinute.sh

# What: HTTP write requests per minute from the last 24 hours, grouped by hour and hostname.
# Why: By reviewing this output, along with the points per minute throughput, it should allow us to get an estimate for the batch size and request volume per instance
echo -e "**************************************"
echo -e writesHttpReqPerMinute.sh
echo -e "**************************************"

influx \
	-host $INFLUXDB_HOST \
	-port $INFLUXDB_PORT \
	-username $INFLUXDB_USER \
	-password $INFLUXDB_PASSWORD \
	-database '_internal' \
	-execute 'SELECT non_negative_derivative(mean("writeReq"), 60s) FROM "_internal"."monitor"."httpd" WHERE time > now() - 24h GROUP BY time(1h),"hostname" fill(0)' \
	> $OUT_DIR/writesHttpReqPerMinute.txt

echo "Done!!"
