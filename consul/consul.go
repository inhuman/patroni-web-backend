package consul

import (
	"errors"
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/inhuman/consulator/command"
	"github.com/mitchellh/cli"
	"jgit.me/tools/patroni-web-backend/config"
	"jgit.me/tools/patroni-web-backend/utils"
	"log"
	"os"
	"strconv"
)

var Client *api.Client

func Init() error {

	cnf := api.DefaultConfig()

	cnf.Address = config.AppConf.ConsulAddress
	cnf.Datacenter = config.AppConf.ConsulDc

	var err error

	Client, err = api.NewClient(cnf)
	if err != nil {
		return err
	}
	return nil
}

func WriteConfigFromYaml(clusterName string, data []byte) error {

	f, err := utils.CreateTempFile(data, ".yaml")
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()

	if err != nil {
		return err
	}

	fmt.Printf("%s\n", data)

	ui := &cli.BasicUi{Writer: os.Stdout}

	cmd := command.ImportCommand{
		Ui:    ui,
		Purge: true,
	}

	code := cmd.Run([]string{
		"-prefix=" + config.AppConf.ConsulKvPrefix + "/clusters/" + clusterName,
		f.Name(),
	})
	log.Println("prefix:", config.AppConf.ConsulKvPrefix+"/clusters/"+clusterName)
	log.Println("file name:", f.Name())

	if code > 0 {
		return errors.New("can not write to consul kv, with code: " + strconv.Itoa(code))
	}

	return nil
}

func SetClusterCheck(clusterName string, isEnabled bool) error {
	pair := &api.KVPair{
		Key:   config.AppConf.ConsulKvPrefix + "/clusters/" + clusterName + "/checks_enabled",
		Value: []byte(strconv.FormatBool(isEnabled)),
	}

	if _, err := Client.KV().Put(pair, nil); err != nil {
		return err
	}
	return nil
}

func DeleteConfig(clusterName string) error {

	key := config.AppConf.ConsulKvPrefix + "/clusters/" + clusterName

	fmt.Println("prefix:", key)

	_, err := Client.KV().DeleteTree(key, nil)
	return err
}
