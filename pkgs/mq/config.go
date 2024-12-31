package mq

type Config struct {
	Address  string        `json:"address"`
	Account  string        `json:"account"`
	Password string        `json:"password"`
	VHost    string        `json:"vhost"`
	Param    RabbitMQParam `json:"param"`
}
