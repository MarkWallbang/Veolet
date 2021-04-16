package lib

// TODO change this to read flow.json -> make all imports in cdc scripts/transactions dynamic
type Configuration struct {
	Node              string `json:"node"`
	ServiceAddressHex string `json:"serviceAddressHex"`
	ServicePrivKeyHex string `json:"servicePrivKeyHex"`
	ServiceSigAlgoHex string `json:"serviceSigAlgoHex"`
}
