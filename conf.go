package main

import(
	"encoding/json"
	"io"
	"io/ioutil"
)

type Config struct {
	DAServer string
	DABaseDN string
	Admins []string
	Users []string
}

func readConf(confFile io.ReadSeeker) (*Config, error) {
	b, err := ioutil.ReadAll(confFile)
	if err != nil {
		return nil, err
	}

	conf := &Config{}
	if err = json.Unmarshal(b, conf); err != nil {
		return nil, err
	}
	return conf, nil
}
