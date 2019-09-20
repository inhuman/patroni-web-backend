package clusters

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/inhuman/consul-kv-mapper"
	"jgit.me/tools/patroni-web-backend/config"
	"jgit.me/tools/patroni-web-backend/db"
	"jgit.me/tools/patroni-web-backend/utils"
	"sort"
	"strconv"
	"time"
)

type Config struct {
	Clusters []*ClusterConfig `json:"clusters"`
}

type ClusterConfig struct {
	Name          string                   `json:"name"`
	Nodes         []*Node                  `json:"nodes"`
	NodesDetailed map[string]*NodeDetailed `json:"nodes_detailed"`
	Scheduled     *Scheduled               `json:"scheduled"`
	Logs          []*ServiceLog            `json:"logs"`
	Memos         []*db.Memo               `json:"memos"`
	Dc            string                   `json:"dc"`
	ChecksEnabled bool                     `json:"checks_enabled"`
}

type ServiceLog struct {
	Service    string            `json:"service"`
	Platform   string            `json:"platform"`
	AccessData map[string]string `json:"access_data"`
}

type Scheduled struct {
	Switchover db.ClusterSwitchover `json:"switchover"`
}

type Node struct {
	Address     string        `json:"address"`
	Dc          string        `json:"dc"`
	Name        string        `json:"name"`
	Auth        string        `json:"auth"`
	ClusterName string        `json:"cluster_name"`
	Memos       []*db.Memo    `json:"memos"`
	Logs        []*ServiceLog `json:"logs"`
}

func FetchClusters(client *api.Client) *Config {

	conf := &Config{}

	confMap, err := consul_kv_mapper.BuildMap(client, config.AppConf.ConsulKvPrefix+"/clusters")
	if err != nil {
		fmt.Printf("fetch clusters from kv error: %s\n", err)
	}

	conf.buildConfig(confMap)
	conf.addClusterScheduledActions()
	conf.addMemos()

	conf.SortClusters()

	return conf
}

func FetchConsulConfig(client *api.Client) string {

	confMap, err := consul_kv_mapper.BuildMap(client, config.AppConf.ConsulKvPrefix+"/clusters")
	if err != nil {
		fmt.Printf("fetch clusters from kv error: %s\n", err)
	}

	yml := ""
	offset := ""

	processChildren(confMap, &yml, offset)
	return yml
}

func processChildren(confMap *consul_kv_mapper.MapType, yml *string, offset string) {
	newLine := "\n"
	valueStr := ""

	if (confMap != nil) && (len(confMap.Children) > 0) {
		for name, child := range confMap.Children {

			nameStr := fmt.Sprintf("%s", name)

			if len(child.Children) > 0 {
				valueStr = offset + nameStr + ": " + newLine
			} else {
				valueStr = offset + nameStr + ": " + fmt.Sprintf("%s", child.Value) + newLine
			}

			*yml += valueStr

			if len(child.Children) > 0 {
				processChildren(child, yml, offset+"  ")
			}
		}
	}
}

func (conf *Config) buildConfig(confMap *consul_kv_mapper.MapType) {

	if confMap != nil {
		for clusterName := range confMap.Children {

			if utils.StringInSlice(string(clusterName), []string{"fake-cluster", "liar-cluster"}) {
				continue
			}

			clusterMap := confMap.Children[clusterName]

			cluster := ClusterConfig{
				Name: string(clusterName),
			}

			if clusterMap.Children["nodes"] != nil {
				for nodeName := range clusterMap.Children["nodes"].Children {

					nodeMap := clusterMap.Children["nodes"].Children[nodeName]

					node := &Node{
						Name:        string(nodeName),
						ClusterName: string(clusterName),
					}

					if nodeMap.Children["address"] != nil {
						node.Address = string(nodeMap.Children["address"].Value)
					}

					if nodeMap.Children["auth"] != nil {
						node.Auth = string(nodeMap.Children["auth"].Value)
					}

					if nodeMap.Children["logs"] != nil {
						node.Logs = buildLogConfig(nodeMap.Children["logs"])
					}

					if clusterMap.Children["dc"] != nil {
						node.Dc = string(clusterMap.Children["dc"].Value)
					}

					if cluster.Nodes == nil {
						cluster.Nodes = []*Node{}
					}
					cluster.Nodes = append(cluster.Nodes, node)
				}
			}

			if clusterMap.Children["logs"] != nil {
				cluster.Logs = buildLogConfig(clusterMap.Children["logs"])
			}

			if clusterMap.Children["dc"] != nil {
				cluster.Dc = string(clusterMap.Children["dc"].Value)
			}

			cluster.ChecksEnabled = buildCheckConfig(clusterMap.Children["checks_enabled"])

			conf.Clusters = append(conf.Clusters, &cluster)
		}
	}
}

func buildLogConfig(confMap *consul_kv_mapper.MapType) []*ServiceLog {

	l := []*ServiceLog{}

	for serviceName := range confMap.Children {

		for platformName := range confMap.Children[serviceName].Children {

			platformAccessData := confMap.Children[serviceName].Children[platformName]

			serviceLog := &ServiceLog{
				Service:  string(serviceName),
				Platform: string(platformName),
			}

			m := make(map[string]string)

			for paramName := range platformAccessData.Children {
				if platformAccessData.Children[paramName].Value != "" {
					m[string(paramName)] = string(platformAccessData.Children[paramName].Value)
				}
			}

			serviceLog.AccessData = m

			l = append(l, serviceLog)
		}

	}
	return l
}

func buildCheckConfig(confMap *consul_kv_mapper.MapType) bool {

	if confMap != nil {
		v, err := strconv.ParseBool(string(confMap.Value))
		if err == nil {
			return v
		}
	}
	return true
}

func (conf *Config) addMemos() {

	for _, cluster := range conf.Clusters {

		cluster.Memos = db.GetMemosById(cluster.Name)

		for _, node := range cluster.Nodes {
			node.Memos = db.GetMemosById(cluster.Name + "-" + node.Name)
		}
	}
}

func (conf *Config) addClusterScheduledActions() {
	for _, cluster := range conf.Clusters {

		scheduledSwitchover := &db.ClusterSwitchover{
			ClusterName: cluster.Name,
		}

		scheduledSwitchover.GetLastScheduledSwitchover()

		nilTime := time.Time{}

		if scheduledSwitchover.ScheduledAt != nilTime {

			if cluster.Scheduled == nil {
				cluster.Scheduled = &Scheduled{}
			}
			cluster.Scheduled.Switchover = *scheduledSwitchover
		}
	}
}

func (c *Config) SortClusters() {
	sort.Slice(c.Clusters[:], func(i, j int) bool {
		return c.Clusters[i].Name < c.Clusters[j].Name
	})
}
