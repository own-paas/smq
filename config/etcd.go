package config

type ETCDConfig struct {
	Endpoints []string `json:"endpoints"`
	User      string   `json:"user"`
	Password  string   `json:"password"`
}
