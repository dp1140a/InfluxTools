# Influx Tools
This is a set of tools to help with operation and support of InfluxEnterprise

Liscense: MIT

Maintainer: Dave Patton

## Rebalancer
A tool for rebalancing your cluster

#### Options
* url – The host and port of the cluster metanode to use (http://localhost:8086)
* debug – Print verbose output to console (false)
* threshold – Threshold for shard count variance on a node (.10)
* autoreplicate – Automatically replicate under replicated shards (false)
* autoExecute – Automatically execute the rebalancing plan (false)

#### Usage:

```bash
./rebalancer [options]
```

## ClusterInfo
A tool for displaying cluster information in JSON.  Will display cluster nodes, and shards, as well as shard ownership.


#### Options
* url – The host and port of the cluster metanode to use (http://localhost:8086)

#### Usage:

```bash
./clusterInfo [options]
```

## UnderReplicated
A tool for displaying any underreplicated shards. Currently this tool will just inform you if you have any under-replicated shards as shown below:

```bash
Shard 150:
    Desired: 3
    Actual: 2 [{5 data2:8088} {6 data3:8088}]

Process finished with exit code 0
```

#### Options
* url – The host and port of the cluster metanode to use (http://localhost:8086)

#### Usage:

```bash
./underReplicated [options]
```

## Profiling
Tools for gathering and displaying Go profile information

## Support Bundle
A set of scripts that will run diagnostics queries on your installation of influxdb.  You can run the scripts individually or run the diagnostics script to run them all.  If you run them all the script will also place the output into a tarball for upload to Influxdata support

#### Scripts
* Count Series and Measurements:
    * What: Series and measurement counts from the last 24 hours, grouped by hour, database, and hostname.
    * Why: This allows us to see the current series and measurement count, as well as how much the series count has grown per hour over the last 24 hours
* Hinted Handoff Queue Size:
    * What: Hinted-handoff queue size from the last 24 hours, grouped by hour, and all tags.
    * Why: This will tell us how large the hinted-handoff queue was for each hour, broken out by target server.
* Hinted Handoff Queue Throughput:
    * What: Hinted-handoff queue throughput per minute size on disk from the last 24 hours, grouped by hour, and all tags.
    * Why: This will tell us how fast (or slow) the hinted-handoff queue is draining (or filling up)
* Read Write Response:
    * What: Hinted-handoff queue size from the last 24 hours, grouped by hour, and all tags.
    * Why: This will tell us how large the hinted-handoff queue was for each hour, broken out by target server.
* Shard Size on Disk:
    * What: Shard size on disk from the last 24 hours, grouped by hour, database, shard ID, and hostname.
    * Why: This will tell us the approximate database size per hour, which can be used for identifying "hot spots" in a cluster or very large shards.
* Writes Http Req Per Minute:
    * What: HTTP write requests per minute from the last 24 hours, grouped by hour and hostname.
    * Why: By reviewing this output, along with the points per minute throughput, it should allow us to get an estimate for the batch size and request volume per instance
* Writes Points Per Minute:
    * What: Points written per minute from the last 24 hours, grouped by hour and hostname.
    * Why: This allows us to see the number of points hitting each instance per minute.

#### Options (for each script)
* -u |--username – The InfluxDB user to use
* -p | --password – The InfluxDB user password
* -h | --host – The InfluxDB host to run the queries against
* -x | --port – The InfluxDB port to use
* -o | --outdir – The directory of the output to use


#### Usage:

```bash
./[scriptName] [options]
```

## ToDo
* Provide ability for underreplicated to replicate shard in addition to just display
* REFACTOR, Refactor, refactor!
* A great dashboard definition for monitoring your Influx Enterprise Cluster.
* Integrate the query profiler tool
* Handle over-replicated shards.
* Make the cluster rebalancer smarter to take shard size into account.
* Dockerize the profile tool for easier use
* Add more tools