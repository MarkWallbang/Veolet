package lib

// structure of flow.json config
type Configuration struct {
	Emulators struct {
		Default struct {
			Port           int    `json:"port"`
			ServiceAccount string `json:"serviceAccount"`
		} `json:"default"`
	} `json:"emulators"`

	Contracts struct {
		NonFungibleToken string `json:"NonFungibleToken"`
		Veolet           string `json:"Veolet"`
	} `json:"contracts"`

	Networks struct {
		Emulator struct {
			Host  string `json:"host"`
			Chain string `json:"chain"`
		} `json:"emulator"`

		Testnet struct {
			Host  string `json:"host"`
			Chain string `json:"chain"`
		} `json:"testnet"`
	} `json:"networks"`

	Accounts struct {
		Emulator_account struct {
			Address string `json:"address"`
			Keys    string `json:"keys"`
			Chain   string `json:"chain"`
		} `json:"emulator-account"`

		Testnet_account struct {
			Address string `json:"address"`
			Keys    string `json:"keys"`
			Chain   string `json:"chain"`
		} `json:"testnet-account"`
	} `json:"accounts"`

	Deployments struct {
		Emulator struct {
			Emulator_account []string `json:"emulator-account"`
		} `json:"emulator"`

		Testnet struct {
			Testnet_account []string `json:"testnet-account"`
		} `json:"testnet"`
	} `json:"deployments"`
}
