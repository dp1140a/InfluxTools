package main

import (
	"flag"
	"github.com/dp1140a/InfluxTools/cluster"
)

var metaNodeURL = flag.String("url", "http://localhost:8091", "The host and port of the cluster metanode to use.")

func main() {
	flag.Parse()
	c := cluster.NewCluster(*metaNodeURL)
	c.PrintUnderReplicatedShards()
}
