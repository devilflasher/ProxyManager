package config

// ProxyConfig 代表一个代理配置
type ProxyConfig struct {
	ID        string        `json:"id" yaml:"id"`
	Name      string        `json:"name" yaml:"name"`
	Upstream  UpstreamProxy `json:"upstream" yaml:"upstream"`
	Local     LocalProxy    `json:"local" yaml:"local"`
	Enabled   bool          `json:"enabled" yaml:"enabled"`
	AutoStart bool          `json:"auto_start" yaml:"auto_start"`
}

type UpstreamProxy struct {
	Protocol   string `json:"protocol" yaml:"protocol"`
	Address    string `json:"address" yaml:"address"`
	Username   string `json:"username" yaml:"username"`
	Password   string `json:"password" yaml:"password"`
	AuthMethod string `json:"auth_method,omitempty" yaml:"auth_method,omitempty"`
}

type LocalProxy struct {
	Protocol   string `json:"protocol" yaml:"protocol"`
	ListenIP   string `json:"listen_ip" yaml:"listen_ip"`
	ListenPort int    `json:"listen_port" yaml:"listen_port"`
}
