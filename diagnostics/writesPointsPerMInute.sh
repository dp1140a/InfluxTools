#!/bin/bash
# writesPointsPerMInute.sh

# What: Points written per minute from the last 24 hours, grouped by hour and hostname.
# Why: This allows us to see the number of points hitting each instance per minute.
echo -e "**************************************"
echo -e writesPointsPerMInute.sh
echo -e "**************************************"

influx \
	-host $INFLUXDB_HOST \
	-port $INFLUXDB_PORT \
	-username $INFLUXDB_USER \
	-password $INFLUXDB_PASSWORD \
	-database '_internal' \
	-execute 'SELECT non_negative_derivative(mean("pointReq"), 60s) FROM "_internal"."monitor"."write" WHERE time > now() - 24h GROUP BY time(1h),"hostname" fill(0)' \
	> $OUT_DIR/writesPointsPerMinute.txt

echo "Done!!"
