package clusters

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync"
)

type NodeDetailed struct {
	Checks *NodeCheck       `json:"checks"`
	Consul *Node            `json:"consul"`
	Info   *json.RawMessage `json:"info"`
	Config *json.RawMessage `json:"config"`
	Common *NodeCommon      `json:"common"`
}

type NodeCommon struct {
	Ips []net.IP
}

type NodeDetailedSynchronizer struct {
	Wg    *sync.WaitGroup
	Mutex *sync.Mutex
}

type NodeCheck struct {
	MasterResponse  int    `json:"master_response"`
	ReplicaResponse int    `json:"replica_response"`
	Message         string `json:"message"`
}

func (conf *Config) FetchNodesDetailed() error {

	nodeDetailedSync := &NodeDetailedSynchronizer{
		Wg:    &sync.WaitGroup{},
		Mutex: &sync.Mutex{},
	}

	for _, cluster := range conf.Clusters {
		for _, node := range cluster.Nodes {
			if cluster.NodesDetailed == nil {
				cluster.NodesDetailed = make(map[string]*NodeDetailed)
			}
			nodeDetailedSync.Wg.Add(1)
			go FetchNodeDetailed(node, cluster, nodeDetailedSync)
		}
	}

	nodeDetailedSync.Wg.Wait()
	return nil
}

func FetchNodeDetailed(node *Node, cluster *ClusterConfig, s *NodeDetailedSynchronizer) {

	nd := NodeDetailed{
		Consul: node,
	}

	if err := nd.addInfo(); err != nil {
		log.Println("fetch node info error:", err)
	}

	if err := nd.addConfig(); err != nil {
		log.Println("fetch node config error:", err)
	}

	if err := nd.addIps(); err != nil {
		log.Println("fetch node ips error:", err)
	}

	s.Mutex.Lock()
	cluster.NodesDetailed[node.Name] = &nd
	s.Mutex.Unlock()

	s.Wg.Done()
}

func (nd *NodeDetailed) addIps() error {

	host, _, err := net.SplitHostPort(nd.Consul.Address)
	if err != nil {
		return err

	}

	addr, err := net.LookupIP(host)
	if err != nil {
		return err
	} else {
		nd.Common = &NodeCommon{
			Ips: addr,
		}
	}

	return nil
}

func (nd *NodeDetailed) addConfig() error {

	response, err := http.Get("http://" + nd.Consul.Address + "/config")
	if err != nil {
		log.Println("/config err:", err)
		nd.Checks.Message = err.Error()

	} else {

		b, err := ioutil.ReadAll(response.Body)
		defer response.Body.Close()
		if err != nil {
			return err
		}

		err = json.Unmarshal(b, &nd.Config)
		if err != nil {
			return err
		}
	}

	return nil
}

func (nd *NodeDetailed) setDefaultInfo() {

	b := json.RawMessage(`{"state": "unreachable", "server_version": 0, "patroni": { "scope": "unknown", "version": "0.0.0" }}`)
	nd.Info = &b
}

func (nd *NodeDetailed) addInfo() error {

	nd.Checks = &NodeCheck{}

	response, err := http.Get("http://" + nd.Consul.Address + "/master")
	if err != nil {
		log.Println("/master err:", err)
		nd.Checks.Message = err.Error()
		nd.setDefaultInfo()

	} else {
		nd.Checks.MasterResponse = response.StatusCode
		// Read body
		b, err := ioutil.ReadAll(response.Body)
		defer response.Body.Close()
		if err != nil {
			return err
		}

		err = json.Unmarshal(b, &nd.Info)
		if err != nil {
			return err
		}
	}

	replicaResponse, err := http.Get("http://" + nd.Consul.Address + "/replica")
	if err != nil {
		log.Println("/replica err:", err)
		nd.Checks.Message = err.Error()

	} else {
		nd.Checks.ReplicaResponse = replicaResponse.StatusCode
	}

	return nil
}
