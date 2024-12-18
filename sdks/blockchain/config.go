package blockchain

type Config struct {
	URL             string `json:"url"`
	ContractAddress string `json:"contractAddress"`
	FunctionName    string `json:"functionName"`
	MemberName      string `json:"memberName"`
	Type            string `json:"type"`
}
