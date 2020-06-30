package main

/**
References:
https://community.hortonworks.com/articles/44148/hdfs-balancer-3-cluster-balancing-algorithm.html
*/
import (
	"flag"
	"fmt"
	"github.com/dp1140a/InfluxTools/cluster"
	"github.com/dp1140a/InfluxTools/net"
	"github.com/dp1140a/InfluxTools/util"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"
)

type PlanStep struct {
	FromNode string
	ToNode   string
	shardId  string
}

type ShardStats struct {
	IdealShardsPerNode int
	OverThreshold      int
	UnderThreshold     int
}

type NodeShards struct {
	NodeId    int
	TCPAddr   string
	Shards    []string
	IdealDiff int
}

const (
	OVER           = "over"
	ABOVE          = "above"
	IDEAL          = "ideal"
	BELOW          = "below"
	UNDER          = "under"
	REQ_PAUSE_TIME = time.Duration(1) * time.Second // Pause for 1 second
)

var metaNodeURL = flag.String("url", "http://10.0.1.215:8091", "The host and port of the cluster metanode to use.")
var verbose = flag.Bool("debug", false, "Print verbose output to console")
var threshold = flag.Float64("threshold", .10, "Threshold for shard count variance on a node")
var autoReplicate = flag.Bool("autoreplicate", false, "Automatically replicate under replicated shards")
var autoExecute = flag.Bool("autoExecute", false, "Automatically execute the rebalancing plan")

var zeta *rand.Rand
var shardStats = ShardStats{} //Keep shard stats at this level since they are calcd once and referred to often

func init() {
	seed := rand.NewSource(time.Now().Unix())
	zeta = rand.New(seed) // initialize local pseudorandom generator
}

func createPlan(nodeShards []NodeShards) []PlanStep {
	plan := make([]PlanStep, 0)
	var step = PlanStep{}
	curState := bucketize(nodeShards)

	fmt.Println(curState)
	/**
	Shard Pairing:
	Over-Utilized -> Under-Utilized
	Over-Utilized -> Below-Average
	Above-Average -> Under-Utilized
	*/
	for isbalancingNeeded(curState) {
		//balance := 0
		//for balance < 3 {
		fmt.Println("Current State: ", curState)
		switch {
		case len(curState[OVER]) > 0 && len(curState[UNDER]) > 0:
			from := curState[OVER]
			to := curState[UNDER]
			step = pair(from, to)
			curState[OVER] = from
			curState[UNDER] = to

		case len(curState[OVER]) > 0 && len(curState[BELOW]) > 0:
			step = pair(curState[OVER], curState[BELOW])

		case len(curState[ABOVE]) > 0 && len(curState[UNDER]) > 0:
			step = pair(curState[ABOVE], curState[UNDER])

		}

		if step == (PlanStep{}) {
			//Empty PlanStep do nothing
		} else {
			plan = append(plan, step)
		}

		tShard := []NodeShards{}
		for _, bucket := range curState {
			for _, node := range bucket {
				tShard = append(tShard, node)
			}
		}

		curState = bucketize(tShard)

		fmt.Println("Current State: ", curState)
		fmt.Println(isbalancingNeeded(curState))
		//balance++
	}

	return plan
}

func pair(from []NodeShards, to []NodeShards) PlanStep {
	fmt.Println("From: ", from)
	fmt.Println("To: ", to)
	var planStep = PlanStep{}
	badShard := true

	for badShard {
		fromIdx := zeta.Intn(len(from[0].Shards))
		potentialShard := from[0].Shards[fromIdx]
		fmt.Printf("Potential Shard %v from node %v to node %v\n", potentialShard, from[0].TCPAddr, to[0].TCPAddr)
		if isShardFreeToCopy(potentialShard, to[0].Shards) {
			fmt.Printf("Shard %v is free to copy\n", potentialShard)

			//add shard to plan
			planStep.shardId = potentialShard
			planStep.FromNode = from[0].TCPAddr
			planStep.ToNode = to[0].TCPAddr

			//update datanodes:
			//	1. remove shard id from "from" shard list in cur State
			from[0] = removeShardFromList(potentialShard, from[0])
			//	2. add shard to "to" node shard list in curState
			to[0].Shards = append(to[0].Shards, potentialShard)
			to[0].IdealDiff = to[0].IdealDiff + 1
			badShard = false
			fmt.Println(from)
			fmt.Println(to)
		} else {
			fmt.Printf("Shard %v is NOT free to copy\n", potentialShard)
		}
	}

	return planStep
}

func removeShardFromList(shardId string, nodeShard NodeShards) NodeShards {

	for i, id := range nodeShard.Shards {
		if id == shardId {
			nodeShard.Shards = append(nodeShard.Shards[:i], nodeShard.Shards[i+1:]...)
			nodeShard.IdealDiff = nodeShard.IdealDiff - 1
			break
		}
	}
	return nodeShard
}

func mapNodeShards(datanodes []cluster.Datanode) []NodeShards {
	nodeShards := make([]NodeShards, len(datanodes))
	for i, node := range datanodes {
		nodeShards[i] = NodeShards{
			node.ID,
			node.TCPAddr,
			node.Shards,
			0,
		}
		fmt.Printf("Node %v has %v shards: %v \n", node.TCPAddr, len(node.Shards), node.Shards)
	}

	return nodeShards
}

func bucketize(nodeShards []NodeShards) map[string][]NodeShards {
	var buckets = make(map[string][]NodeShards)
	buckets[OVER] = make([]NodeShards, 0)
	buckets[ABOVE] = make([]NodeShards, 0)
	buckets[IDEAL] = make([]NodeShards, 0)
	buckets[BELOW] = make([]NodeShards, 0)
	buckets[UNDER] = make([]NodeShards, 0)

	for i, node := range nodeShards {
		nodeShards[i].IdealDiff = len(node.Shards) - shardStats.IdealShardsPerNode
		shardCount := len(nodeShards[i].Shards)

		ovrundr := UNDER
		if nodeShards[i].IdealDiff > 0 {
			ovrundr = OVER
		}
		fmt.Printf("Node %v: %d shard(s) %v\n", node.TCPAddr, nodeShards[i].IdealDiff, ovrundr)

		switch {

		case shardCount >= shardStats.OverThreshold:
			buckets[OVER] = append(buckets[OVER], nodeShards[i])

		case shardCount > shardStats.IdealShardsPerNode && shardStats.UnderThreshold < shardStats.OverThreshold:
			buckets[ABOVE] = append(buckets[ABOVE], nodeShards[i])

		case shardCount == shardStats.IdealShardsPerNode:
			buckets[IDEAL] = append(buckets[IDEAL], nodeShards[i])

		case shardCount < shardStats.IdealShardsPerNode && shardCount >= shardStats.UnderThreshold:
			buckets[BELOW] = append(buckets[BELOW], nodeShards[i])

		case shardCount <= shardStats.UnderThreshold:
			buckets[UNDER] = append(buckets[UNDER], nodeShards[i])

		}
	}

	return buckets
}

func getShardCount(shards map[string]*cluster.Shard) int {
	shardCount := 0
	for _, shard := range shards {
		shardCount += shard.ReplicaN
	}

	return shardCount
}

func calcNodeShardStats(datanodes []cluster.Datanode, totalShards int) ShardStats {
	idealShardsPerNode := int(math.Floor(float64(totalShards) / float64(len(datanodes))))
	overThreshold := int(math.Floor(float64(idealShardsPerNode) + (float64(idealShardsPerNode) * *threshold)))
	underThreshold := int(math.Ceil(float64(idealShardsPerNode) - (float64(idealShardsPerNode) * *threshold)))

	fmt.Printf("There should be %d total shards.\n", totalShards)
	fmt.Printf("Each node should have %d shards\n", idealShardsPerNode)
	fmt.Println("Overthrehold is: ", overThreshold)
	fmt.Println("Underthreshold is: ", underThreshold)

	return ShardStats{
		idealShardsPerNode,
		overThreshold,
		underThreshold,
	}
}

func getDatanodeById(nodeId int, datanodes []cluster.Datanode) cluster.Datanode {
	for idx, node := range datanodes {
		if nodeId == node.ID {
			return datanodes[idx]
		}
	}

	return cluster.Datanode{}
}

//Over-Utilized -> Under-Utilized
//Over-Utilized -> Below-Average
//Above-Average -> Under-Utilized
func isbalancingNeeded(buckets map[string][]NodeShards) bool {
	if len(buckets[OVER]) > 0 && len(buckets[UNDER]) > 0 {
		return true
	} else if len(buckets[OVER]) > 0 && len(buckets[BELOW]) > 0 {
		return true
	} else if len(buckets[ABOVE]) > 0 && len(buckets[UNDER]) > 0 {
		return true
	} else {
		return false
	}
}

func isShardFreeToCopy(shardId string, shardList []string) bool {
	fmt.Printf("Looking for shard %v in %v\n", shardId, shardList)
	doesHave := true
	for _, id := range shardList {
		if id == shardId {
			doesHave = false
			break
		}
	}

	return doesHave
}

func truncateShards() {
	fmt.Printf("Truncating Shards\n")
	//fmt.Printf("Using MetaNode URL: %v\n", *metaNodeURL)
	body := net.MakeRequest(*metaNodeURL+"/truncate-shards", net.POST, "delay=1m0s")

	if string(body) == "204" {
		fmt.Println("Truncate Shards: OK")
	}

}

//copy-shard dest=data2%3A8088&shard=24&src=data1%3A8088
// remove-shard shard=24&src=data1%3A8088
func doRebalance(plan []PlanStep) {
	fmt.Printf("Moving Shards\n")
	for _, step := range plan {
		copyShard(step)
		removeShard(step)
	}
}

func removeShard(planStep PlanStep) {
	fmt.Printf("Removing Shard %v from %v\n", planStep.shardId, planStep.FromNode)
	// Encoded to shard=161&src=data2%3A8088
	qStr := "shard=" + planStep.shardId + "&src=" + strings.Replace(planStep.FromNode, ":", "%3A", -1)
	//fmt.Println(qStr)
	body := net.MakeRequest(*metaNodeURL+"/remove-shard?"+qStr, net.POST, "")

	if string(body) == "204" {
		fmt.Println("Remove shard OK")
	} else {
		fmt.Println("oops something went wrong")
		fmt.Println(string(body[:]))
	}
}

func copyShard(planStep PlanStep) {
	fmt.Printf("Copying Shard %v from %v to %v\n", planStep.shardId, planStep.FromNode, planStep.ToNode)
	// Encoded to dest=data2%3A8088&shard=27&src=data1%3A8088
	qStr := "dest=" + strings.Replace(planStep.ToNode, ":", "%3A", -1) + "&shard=" + planStep.shardId + "&src=" + strings.Replace(planStep.FromNode, ":", "%3A", -1)
	//fmt.Println(qStr)
	body := net.MakeRequest(*metaNodeURL+"/copy-shard?"+qStr, net.POST, "")

	if string(body) == "204" {
		fmt.Println("Shard Copy OK")
	} else {
		fmt.Println("oops something went wrong")
		fmt.Println(string(body[:]))
	}
}

func copyUnderReplicatedShard(underReplicatedShards map[string]cluster.Shard, datanodes []cluster.Datanode) {
	fmt.Println("")
	//TODO: Handle case where underreplication is more than one ie: should have 3 copies but only has 1
	for _, shard := range underReplicatedShards {
		diff := shard.ReplicaN - len(shard.Owners)
		fmt.Printf("Shard %s\n-----------------\n", shard.ID)
		fmt.Printf("Shard %s needs to be copied %d times\n", shard.ID, diff)
		for i := 1; i <= diff; i++ {
			badCopy := true
			nodeIdx := 0
			for badCopy {
				if isShardFreeToCopy(shard.ID, datanodes[nodeIdx].Shards) {
					fmt.Printf("Copying shard %v to %v.\n", shard.ID, datanodes[nodeIdx].TCPAddr)

					copyShard(PlanStep{
						shard.Owners[rand.Intn(len(shard.Owners))].TCPAddr,
						datanodes[nodeIdx].TCPAddr,
						shard.ID,
					})
					fmt.Println("")
					time.Sleep(REQ_PAUSE_TIME)
					badCopy = false
				} else {
					fmt.Printf("Shard %v is NOT free to copy to %v\n", shard.ID, datanodes[nodeIdx].TCPAddr)
					nodeIdx++
				}
			}
		}
	}
}

func askForConfirmation() bool {
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		fmt.Println(err)
	}
	okayResponses := []string{"y", "Y", "yes", "Yes", "YES"}
	nokayResponses := []string{"n", "N", "no", "No", "NO"}
	if util.ContainsString(okayResponses, response) {
		return true
	} else if util.ContainsString(nokayResponses, response) {
		return false
	} else {
		fmt.Println("Please type yes or no and then press enter:")
		return askForConfirmation()
	}
}

func printPlan(plan []PlanStep) {
	for _, step := range plan {
		fmt.Printf("Move shard %v from %v to %v\n", step.shardId, step.FromNode, step.ToNode)
	}
}

func main() {

	/**
	Rebalance to Create Space
	From https://docs.influxdata.com/enterprise_influxdb/v1.5/guides/rebalance/#rebalance-procedure-1-rebalance-a-cluster-to-create-space
	1. Truncate hot shards: /truncate-shards
	2. ID Cold Shards:
		Look at: "end-time"
	3. Copy Cold Shards
	4. Confirm copied shards
	5. Remove Unnecessary Cold Shards
	6. Confirm the Rebalance

	Rebalance to Increase Availbility
	From https://docs.influxdata.com/enterprise_influxdb/v1.5/guides/rebalance/#rebalance-procedure-2-rebalance-a-cluster-to-increase-availability
	1. Update Retention Policy
	2. Truncate hot shards: /truncate-shards
	3. ID Cold Shards:
		Look at: "end-time"
	4. Copy Cold Shards
	5. Confirm the Rebalance
	*/
	flag.Parse()
	fmt.Println("Starting Influx Cluster Rebalancer")
	fmt.Printf("Using Meta Node Address %s\n", *metaNodeURL)
	if *verbose == false {
		fmt.Println("No verbose output.  Set verbose flag to true to see additional detail")
	}

	fmt.Println("Gathering Cluster Information")
	c := cluster.NewCluster(*metaNodeURL) // Create new cluster
	if *verbose {
		c.PrintCluster()
	}

	// Check for and handle under replicated shards
	// TODO: If we have under-replicated shards we should handle it.  Otherwise our shard count will be screwed up
	underReplicatedShards := c.GetUnderReplcatedShards()
	if len(underReplicatedShards) > 0 {
		fmt.Println("There are under replicated shards")

		c.PrintUnderReplicatedShards()

		if *autoReplicate == true {
			fmt.Println("Replicate flag is on.  We will automatically replicate under replicated shards.")
			copyUnderReplicatedShard(underReplicatedShards, c.GetDataNodes())

		} else {
			fmt.Println("Replicate flag is off.  Since there are under replicated shards you will need to fix this before continuing with a rebalance.")
			fmt.Println("Do you wish to replicate underreplicated shards? [Y/N]")
			if askForConfirmation() {
				fmt.Println("Ok. Starting Replication.")
				copyUnderReplicatedShard(underReplicatedShards, c.GetDataNodes())
			} else {
				fmt.Println("Exiting")
				os.Exit(0)
			}
		}
	}

	truncateShards()                                                                // Truncate Shards
	shardStats = calcNodeShardStats(c.GetDataNodes(), getShardCount(c.GetShards())) // Get Shard Stats
	plan := createPlan(mapNodeShards(c.GetDataNodes()))                             // Create the rebalance plan

	//Execute the plan if not empty
	if len(plan) == 0 {
		fmt.Println("Cluster is already balanced.  Nothing to do.")
		os.Exit(0)
	} else {
		fmt.Println("Rebalance Plan:")
		fmt.Printf("The plan has %d steps.\n", len(plan))
		printPlan(plan)
		if *autoExecute == true {
			fmt.Println("Execute flag is on.  We will automatically execute the plan.")
			doRebalance(plan)
		} else {
			fmt.Println("Do you wish to execute the plan? [Y/N]")
			if askForConfirmation() {
				fmt.Println("Got it. Executing plan.")
				doRebalance(plan)
			} else {
				fmt.Println("Exiting")
				os.Exit(0)
			}
		}
	}
}
