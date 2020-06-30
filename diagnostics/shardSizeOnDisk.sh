#!/bin/bash
# shardSizeOnDisk.sh

# What: Shard size on disk from the last 24 hours, grouped by hour, database, shard ID, and hostname.
# Why: This will tell us the approximate database size per hour, which can be used for identifying "hot spots" in a cluster or very large shards.
echo -e "**************************************"
echo -e shardSizeOnDisk.sh
echo -e "**************************************"

influx \
	-host $INFLUXDB_HOST \
	-port $INFLUXDB_PORT \
	-username $INFLUXDB_USER \
	-password $INFLUXDB_PASSWORD \
	-database '_internal' \
	-execute 'SELECT last(diskBytes) AS diskBytes FROM "_internal"."monitor"."shard" WHERE time > now() - 24h GROUP BY time(1h),"id","database","hostname" fill(0)' \
	> $OUT_DIR/shardSizeOnDisk.txt

echo "Done!!"
