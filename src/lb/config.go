package lb

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"syscall"
	"time"
)

const (
	DefaultTimeout = 2
	DefaultType    = "tcp"
	CheckInterval  = 10
)

type Config struct {
	Entries []*Entry
}

type Entry struct {
	ListenAddr string
	Type       string
	Timeout    int
	Backends   []*Backend
	Backend    string
	CertFile   string
	KeyFile    string
	Comment    string
}

type ReadFunc func(string) ([]byte, error)

func getHTTP(url string) ([]byte, error) {
	r, err := http.Get(url)
	if err != nil {
		return []byte{}, err
	}

	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return []byte{}, err
	}

	return data, nil
}

func getHash(b []byte) string {
	hasher := sha1.New()
	hasher.Write(b)
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

func checkConfig(url string, original []byte, rf ReadFunc, signalChan chan os.Signal) {
	hash := getHash(original)
	for true {
		time.Sleep(time.Second * CheckInterval)
		data, err := rf(url)
		if err == nil {
			if getHash(data) != hash {
				break
			}
		}
		// else backoff with sleep???
	}
	log.Printf("config change detected: [%v]", url)
	signalChan <- syscall.SIGHUP
}

func LoadConfig(path string, signalChan chan os.Signal) (*Config, error) {
	config := Config{}
	var data []byte
	var err error
	var rf ReadFunc
	r := regexp.MustCompile("^http[s]?://.*$")
	if r.MatchString(path) {
		rf = getHTTP
	} else {
		rf = ioutil.ReadFile
	}

	if data, err = rf(path); err != nil {
		return nil, err
	}

	// start a goroutine that checks for changes to the config
	go checkConfig(path, data, rf, signalChan)

	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	for _, e := range config.Entries {
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

	return &config, nil
}
