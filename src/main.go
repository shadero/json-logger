package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

func recordJsonValues(w api.WriteAPI, measurement string, tags map[string]string, fields map[string]interface{}) {
	point := influxdb2.NewPoint(
		measurement,
		tags,
		fields,
		time.Now(),
	)
	w.WritePoint(point)
}

type Config struct {
	Record   RecordConfig
	InfluxDb InfluxDbConfig
}

type RecordConfig struct {
	JsonURL            string            `toml:"jsonURL"`
	RecordIntervalMsec int               `toml:"recordIntervalMsec"`
	Tags               map[string]string `toml:"tagsToAttachWhenRecord"`
}

type InfluxDbConfig struct {
	DbURL            string `toml:"dbUrl"`
	Token            string `toml:"token"`
	BucketName       string `toml:"bucketName"`
	OrganizationName string `toml:"organizationName"`
	MeasurementName  string `toml:"measurementName"`
}

func readConfig(configFileName string) (Config, error) {
	var conf Config
	_, err := toml.DecodeFile(configFileName, &conf)
	return conf, err
}

func createConfigFile(configFilePath string) error {
	recordConfig := RecordConfig{
		JsonURL:            "http://192.168.0.XXX",
		RecordIntervalMsec: 5000,
		Tags:               map[string]string{"Host": "Home00"},
	}

	dbConfig := InfluxDbConfig{
		DbURL:            "http://192.168.0.XXX",
		Token:            "XXXXXXXXXXXXXXXXXXXXX",
		BucketName:       "MyBucket",
		OrganizationName: "MyOrganization",
		MeasurementName:  "MyMeasurement",
	}

	config := Config{Record: recordConfig, InfluxDb: dbConfig}

	var buffer bytes.Buffer
	encoder := toml.NewEncoder(&buffer)
	err := encoder.Encode(config)
	if err != nil {
		return err
	}
	fp, err := os.Create(configFilePath)
	if err != nil {
		return err
	}
	defer fp.Close()

	_, err = fp.WriteString(buffer.String())
	return err
}

func parseJsonAPI(apiUrl string) (result map[string]interface{}, err error) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get(apiUrl)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	byte1, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	json.Unmarshal(byte1, &result)
	return
}

func main() {

	conf, err := readConfig("config.toml")
	if os.IsNotExist(err) {
		log.Println("コンフィグファイルが存在しないので作成します。")
		createConfigFile("config.toml")
		return
	}
	if err != nil {
		panic(err)
	}

	c := influxdb2.NewClient(conf.InfluxDb.DbURL, conf.InfluxDb.Token)
	defer c.Close()
	w := c.WriteAPI(conf.InfluxDb.OrganizationName, conf.InfluxDb.BucketName)

	ticker := time.NewTicker(time.Duration(conf.Record.RecordIntervalMsec) * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			fields, err := parseJsonAPI(conf.Record.JsonURL)
			if err != nil {
				log.Println(err)
				continue
			}

			recordJsonValues(w, conf.InfluxDb.MeasurementName, conf.Record.Tags, fields)
		}
	}
}
