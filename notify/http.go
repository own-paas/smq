package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sestack/smq/global"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type notificationHTTP struct {
	server string
	client *http.Client
}

func InitHttp() *notificationHTTP {
	httpClient := &http.Client{
		Transport: &http.Transport{
			MaxConnsPerHost:     100,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
		},
		Timeout: time.Second * 100,
	}
	return &notificationHTTP{
		server: global.CONFIG.Notify.Server,
		client: httpClient,
	}
}

func (h *notificationHTTP) OnClientConnect(data interface{}) bool {
	url := fmt.Sprintf("%s/%s", h.server, "OnClientConnect")
	obj, err := json.Marshal(&data)
	if err != nil {
		return false
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(obj))
	if err != nil {
		return false
	}

	defer resp.Body.Close()
	io.Copy(ioutil.Discard, resp.Body)
	if resp.StatusCode == http.StatusOK {
		return true
	}

	return false
}
func (h *notificationHTTP) OnClientDisconnected(data interface{}) bool {
	url := fmt.Sprintf("%s/%s", h.server, "OnClientDisconnected")
	obj, err := json.Marshal(&data)
	if err != nil {
		return false
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(obj))
	if err != nil {
		return false
	}

	defer resp.Body.Close()
	io.Copy(ioutil.Discard, resp.Body)
	if resp.StatusCode == http.StatusOK {
		return true
	}

	return false
}
func (h *notificationHTTP) OnSubscribe(data interface{}) bool {
	url := fmt.Sprintf("%s/%s", h.server, "OnSubscribe")
	obj, err := json.Marshal(&data)
	if err != nil {
		return false
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(obj))
	if err != nil {
		return false
	}

	defer resp.Body.Close()
	io.Copy(ioutil.Discard, resp.Body)
	if resp.StatusCode == http.StatusOK {
		return true
	}

	return false
}
func (h *notificationHTTP) OnUnSubscribe(data interface{}) bool {
	url := fmt.Sprintf("%s/%s", h.server, "OnUnSubscribe")
	obj, err := json.Marshal(&data)
	if err != nil {
		return false
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(obj))
	if err != nil {
		return false
	}

	defer resp.Body.Close()
	io.Copy(ioutil.Discard, resp.Body)
	if resp.StatusCode == http.StatusOK {
		return true
	}

	return false
}
func (h *notificationHTTP) OnAgentConnect(data interface{}) bool {
	url := fmt.Sprintf("%s/%s", h.server, "OnAgentConnect")
	obj, err := json.Marshal(&data)
	if err != nil {
		return false
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(obj))
	if err != nil {
		return false
	}

	defer resp.Body.Close()
	io.Copy(ioutil.Discard, resp.Body)
	if resp.StatusCode == http.StatusOK {
		return true
	}

	return false
}
func (h *notificationHTTP) OnAgentDisconnected(data interface{}) bool {
	url := fmt.Sprintf("%s/%s", h.server, "OnAgentDisconnected")
	obj, err := json.Marshal(&data)
	if err != nil {
		return false
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(obj))
	if err != nil {
		return false
	}

	defer resp.Body.Close()
	io.Copy(ioutil.Discard, resp.Body)
	if resp.StatusCode == http.StatusOK {
		return true
	}

	return false
}
