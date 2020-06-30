package cluster

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/dp1140a/InfluxTools/net"
)

type Cluster struct {
	Nodes  Nodes
	Shards map[string]*Shard
}

type Nodes struct {
	Data []Datanode `json:"data"`
	Meta []Metanode `json:"meta"`
}

type Metanode struct {
	ID         int    `json:"id"`
	Addr       string `json:"addr"`
	HTTPScheme string `json:"httpScheme"`
	TCPAddr    string `json:"tcpAddr"`
	Version    string `json:"version"`
}

type Datanode struct {
	ID         int    `json:"id"`
	HTTPAddr   string `json:"httpAddr"`
	HTTPScheme string `json:"httpScheme"`
	Status     string `json:"status"`
	TCPAddr    string `json:"tcpAddr"`
	Version    string `json:"version"`
	Shards     []string
}

type Shard struct {
	ID              string    `json:"id"`
	Database        string    `json:"database"`
	RetentionPolicy string    `json:"retention-policy"`
	ReplicaN        int       `json:"replica-n"`
	ShardGroupID    string    `json:"shard-group-id"`
	StartTime       time.Time `json:"start-time"`
	EndTime         time.Time `json:"end-time"`
	ExpireTime      time.Time `json:"expire-time"`
	TruncatedAt     time.Time `json:"truncated-at"`
	Owners          []struct {
		ID      string `json:"id"`
		TCPAddr string `json:"tcpAddr"`
	} `json:"owners"`
	DiskBytes int
	Path      string
	Closed    bool
}

func init() {
	log.SetFlags(log.LstdFlags)
}

func NewCluster(metaNodeURL string) *Cluster {
	c := &Cluster{}
	c.buildNodeInfo(metaNodeURL)
	c.buildShardInfo(metaNodeURL)
	//c.getShardSizes()

	return c
}

func (cluster *Cluster) buildNodeInfo(metaNodeURL string) {
	body := net.MakeRequest(metaNodeURL+"/show-cluster", net.GET, "")

	nodes := Nodes{}
	jsonErr := json.Unmarshal(body, &nodes)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	cluster.Nodes = nodes
}

func (cluster *Cluster) buildShardInfo(metaNodeURL string) {
	body := net.MakeRequest(metaNodeURL+"/show-shards", net.GET, "")
	var shardsSlice []*Shard
	now := time.Now()

	jsonErr := json.Unmarshal(body, &shardsSlice)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	shardsMap := make(map[string]*Shard)
	nodeMap := make(map[string][]string)
	for i := 0; i < len(shardsSlice); i++ {
		shardsMap[shardsSlice[i].ID] = shardsSlice[i]
		for _, owner := range shardsSlice[i].Owners {
			nodeMap[owner.ID] = append(nodeMap[owner.ID], shardsSlice[i].ID)
		}

		if shardsMap[shardsSlice[i].ID].EndTime.Before(now) {
			shardsMap[shardsSlice[i].ID].Closed = true
		} else {
			shardsMap[shardsSlice[i].ID].Closed = false
		}
	}

	for idx, node := range cluster.Nodes.Data {
		cluster.Nodes.Data[idx].Shards = nodeMap[strconv.Itoa(node.ID)]
	}

	cluster.Shards = shardsMap
}

func (cluster Cluster) GetShardSizes() {
	myExp := regexp.MustCompile(`(?:shard:[\/\w:]+)`)
	for _, node := range cluster.Nodes.Data {
		b := net.MakeRequest("http://"+node.HTTPAddr+"/debug/vars", net.GET, "")
		c := make(map[string]interface{})
		json.Unmarshal(b, &c)
		shardNames := myExp.FindAllString(string(b), -1)
		for _, shardName := range shardNames {
			var shardId string
			var shardPath string
			var shardSize int
			for key, shard := range c[shardName].(map[string]interface{}) {
				switch key {
				case "tags":
					tags := shard.(map[string]interface{})
					shardId = tags["id"].(string)
					shardPath = tags["path"].(string)

				case "values":
					values := shard.(map[string]interface{})
					shardSize = int(values["diskBytes"].(float64))

				}
			}
			cluster.Shards[shardId].Path = shardPath
			cluster.Shards[shardId].DiskBytes = shardSize
		}
	}
}

func (cluster Cluster) GetDataNodes() []Datanode {
	return cluster.Nodes.Data
}

func (cluster Cluster) GetMetaNodes() []Metanode {
	return cluster.Nodes.Meta
}

func (cluster Cluster) GetShards() map[string]*Shard {
	return cluster.Shards
}

func (cluster Cluster) GetUnderReplcatedShards() map[string]Shard {
	var underShards = make(map[string]Shard)
	for _, shard := range cluster.Shards {
		if len(shard.Owners) < shard.ReplicaN {
			underShards[shard.ID] = *shard
		}
	}

	return underShards
}

func (cluster Cluster) PrintCluster() {
	clusterJSON, _ := json.MarshalIndent(cluster, "", "\t")
	fmt.Println(string(clusterJSON))
}

func (cluster Cluster) PrintShards() {
	shardJSON, _ := json.MarshalIndent(cluster.Shards, "", "\t")
	fmt.Println(string(shardJSON))
}

func (cluster Cluster) PrintNodes() {
	nodesJSON, _ := json.MarshalIndent(cluster.Nodes, "", "\t")
	fmt.Println(string(nodesJSON))
}

func (cluster Cluster) PrintUnderReplicatedShards() {
	if len(cluster.GetUnderReplcatedShards()) > 0 {
		for _, shard := range cluster.GetUnderReplcatedShards() {
			fmt.Printf("Shard %v:\n", shard.ID)
			fmt.Printf("\tDesired: %d\n", shard.ReplicaN)
			fmt.Printf("\tActual: %d %v\n", len(shard.Owners), shard.Owners)
		}
	} else {
		fmt.Println("There are no under replicated shards")
	}

}
