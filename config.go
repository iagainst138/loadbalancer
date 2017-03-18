package main

import (
	"encoding/json"
	"io/ioutil"
)

const (
	DefaultTimeout = 2
	DefaultType    = "tcp"
)

type Config []*Entry

type Entry struct {
	ListenAddr string
	Type       string
	Timeout    int
	Backends   []*Backend
	Backend    string
	CertFile   string
	KeyFile    string
}

func LoadConfig(path string) (Config, error) {
	c := Config{}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &c)
	if err != nil {
		return nil, err
	}

	for _, e := range c {
		if e.Timeout == 0 {
			e.Timeout = DefaultTimeout
		}
		if e.Type == "" {
			e.Type = DefaultType
		}
		// TODO error on incorrect backend
		if e.Backend == "" {
			e.Backend = "RoundRobin"
		}
	}

	return c, nil
}
