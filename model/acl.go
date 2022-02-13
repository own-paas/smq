package model

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
)

type Acl struct {
	ID       string `json:"id"`
	Priority uint   `json:"priority"`
	Topic    string `json:"topic"`
	Classify string `json:"classify"`
	Type     string `json:"type"`
	Value    string `json:"value"`
	Action   string `json:"action"`
	CreateAt int64  `json:"create_at"`
}

func (self *Acl) Marshal() ([]byte, error) {
	return json.Marshal(self)
}

func (self *Acl) Unmarshal(byte []byte) error {
	return json.Unmarshal(byte, self)
}

func (a *Acl) CheckWithClientID(classify, clientid, topic string) (bool, bool) {
	auth := false
	match := false
	if a.Value == "*" || a.Value == clientid {

		des := strings.Replace(a.Topic, "%c", clientid, -1)
		if classify == "pub" {
			if pubTopicMatch(topic, des) {
				match = true
				auth = a.CheckAuth("pub")
			}
		} else if classify == "sub" {
			if subTopicMatch(topic, des) {
				match = true
				auth = a.CheckAuth("sub")
			}
		}
	}
	return match, auth
}

func (a *Acl) CheckWithUsername(classify, username, topic string) (bool, bool) {
	auth := false
	match := false
	if a.Value == "*" || a.Value == username {
		des := strings.Replace(a.Topic, "%u", username, -1)
		if classify == "pub" {
			if pubTopicMatch(topic, des) {
				match = true
				auth = a.CheckAuth("pub")
			}
		} else if classify == "sub" {
			if subTopicMatch(topic, des) {
				match = true
				auth = a.CheckAuth("sub")
			}
		}
	}
	return match, auth
}

func (a *Acl) CheckWithIP(classify, ip, topic string) (bool, bool) {
	auth := false
	match := false
	if a.Value == "*" || a.Value == ip {
		des := a.Topic
		if classify == "pub" {
			if pubTopicMatch(topic, des) {
				auth = a.CheckAuth("pub")
				match = true
			}
		} else if classify == "sub" {
			if subTopicMatch(topic, des) {
				auth = a.CheckAuth("sub")
				match = true
			}
		}
	}
	return match, auth
}

func (a *Acl) CheckAuth(classify string) bool {
	auth := false
	if classify == "pub" {
		if a.Action == "allow" {
			auth = true
		}
	} else if classify == "sub" {
		if a.Action == "allow" {
			auth = true
		}
	}
	return auth
}

func pubTopicMatch(pub, des string) bool {
	dest, _ := SubscribeTopicSpilt(des)
	topic, _ := PublishTopicSpilt(pub)
	for i, t := range dest {
		if i > len(topic)-1 {
			return false
		}
		if t == "#" {
			return true
		}
		if t == "+" || t == topic[i] {
			continue
		}
		if t != topic[i] {
			return false
		}
	}
	return true
}

func subTopicMatch(pub, des string) bool {
	dest, _ := SubscribeTopicSpilt(des)
	topic, _ := SubscribeTopicSpilt(pub)
	for i, t := range dest {
		if i > len(topic)-1 {
			return false
		}
		if t == "#" {
			return true
		}
		if t == "+" || "+" == topic[i] || t == topic[i] {
			continue
		}
		if t != topic[i] {
			return false
		}
	}
	return true
}

func SubscribeTopicSpilt(topic string) ([]string, error) {
	subject := []byte(topic)
	if bytes.IndexByte(subject, '#') != -1 {
		if bytes.IndexByte(subject, '#') != len(subject)-1 {
			return nil, errors.New("Topic format error with index of #")
		}
	}
	re := strings.Split(topic, "/")
	for i, v := range re {
		if i != 0 && i != (len(re)-1) {
			if v == "" {
				return nil, errors.New("Topic format error with index of //")
			}
			if strings.Contains(v, "+") && v != "+" {
				return nil, errors.New("Topic format error with index of +")
			}
		} else {
			if v == "" {
				re[i] = "/"
			}
		}
	}
	return re, nil

}

func PublishTopicSpilt(topic string) ([]string, error) {
	subject := []byte(topic)
	if bytes.IndexByte(subject, '#') != -1 || bytes.IndexByte(subject, '+') != -1 {
		return nil, errors.New("Publish Topic format error with + and #")
	}
	re := strings.Split(topic, "/")
	for i, v := range re {
		if v == "" {
			if i != 0 && i != (len(re)-1) {
				return nil, errors.New("Topic format error with index of //")
			} else {
				re[i] = "/"
			}
		}

	}
	return re, nil
}
