package blockchain

import (
	"io/ioutil"
	"log"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type blockChainConfig struct {
	Defaults struct {
		StakeReward int64 `yaml:"stakeReward"`
	} `yaml:"defaults"`
	Storage struct {
		Type string `yaml:"type"`
		Path string `yaml:"path"`
	} `yaml:"storage"`
}

var ConfigBaseDir string
var ChainConfig blockChainConfig

func LoadConfig() error {

	yamlFile, err := ioutil.ReadFile(ConfigBaseDir + "/config.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
		return err
	}
	err = yaml.Unmarshal(yamlFile, &ChainConfig)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
		return err
	}

	return nil
}

func ReadNodeByAddress(address string) (*Node, error) {
	var node *Node
	yamlFile, err := ioutil.ReadFile(ChainConfig.Storage.Path + "/nodes/" + address)
	if err != nil {
		logrus.Error("yamlFile.Get err   #%v ", err)
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, &node)
	if err != nil {
		logrus.Error("Unmarshal: %v", err)
		return nil, err
	}

	return node, nil
}
