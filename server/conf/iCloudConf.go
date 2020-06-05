package conf

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
)

const CONF_NAME = "./iCloud.xml"

var (
	Iconf = new(iCloudConf)
)

type iCloudConf struct {
	Ip          string        `xml:"ip"`          // service listen on
	Port        int           `xml:"port"`        // service listen on
	Etcd        []string      `xml:"etcd"`        // etcd ip:port
	Mongo       string        `xml:"mongo"`       // mondoDB
	Log         iCloudLogConf `xml:"log"`
}

type iCloudLogConf struct {
	WebLogName string `xml:"webLogName"`
	Name       string `xml:"name"`       // log name of project
	MaxSize    int    `xml:"maxSize"`    // max size of log (MB)
	MaxBackups int    `xml:"maxBackups"` // max number of old log
	MaxAge     int    `xml:"maxAge"`     // ax days of old log retained
	Compress   bool   `xml:"compress"`   // compress or not
	Level      string `xml:"level"`      // log level
}

func (conf *iCloudConf) newConf() (err error) {
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

func ICloudConfInit() (err error) {
	if err = Iconf.newConf(); err != nil {
		return
	}

	return nil
}
