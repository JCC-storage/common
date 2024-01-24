package imsdk

type Config struct {
	URL string `json:"url"`
}

type ProxyConfig struct {
	IP         string `json:"ip"`
	ClientPort string `json:"clientPort"`
	NodePort   string `json:"nodePort"`
}
