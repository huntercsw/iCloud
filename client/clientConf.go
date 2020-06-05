package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
)

const CONF_NAME = "./clientConf.xml"

type ClientConf struct {
	ExportIp   string   `xml:"exportIp"`
	Etcd       []string `xml:"etcd"`  // etcd ip:port
	Mongo      string   `xml:"mongo"` // mondoDB
	Log        string   `xml:"log"`
	Level      string   `xml:"level"`
	MaxSize    int      `xml:"maxSize"`    // max size of log (MB)
	MaxBackups int      `xml:"maxBackups"` // max number of old log
	MaxAge     int      `xml:"maxAge"`     // ax days of old log retained
	Compress   bool     `xml:"compress"`   // compress or not
}

func (conf *ClientConf) newConf() (err error) {
	var (
		confContect []byte
	)
	if confContect, err = ioutil.ReadFile(CONF_NAME); err != nil {
		fmt.Println("read config file error:", err)
		return
	}

	if err = xml.Unmarshal(confContect, conf); err != nil {
		fmt.Println("conf unmarshal to struct error:", err)
		return
	}

	return nil
}
