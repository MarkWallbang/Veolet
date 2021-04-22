package lib

// structure of flow.json config
type FlowConfiguration struct {
	Emulators struct {
		Default struct {
			Port           int    `json:"port"`
			ServiceAccount string `json:"serviceAccount"`
		} `json:"default"`
	} `json:"emulators"`

	Contracts struct {
		NonFungibleToken struct {
			Source  string `json:"source"`
			Aliases struct {
				Testnet string `json:"testnet"`
			} `json:"aliases"`
		} `json:"NonFungibleToken"`
		Veolet string `json:"Veolet"`
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

	Contractaddresses struct {
		Emulator struct {
			NonFungibleToken string `json:"NonFungibleToken"`
			Veolet           string `json:"Veolet"`
			FungibleToken    string `json:"FungibleToken"`
			FlowToken        string `json:"FlowToken"`
		} `json:"emulator"`
		Testnet struct {
			NonFungibleToken string `json:"NonFungibleToken"`
			Veolet           string `json:"Veolet"`
			FungibleToken    string `json:"FungibleToken"`
			FlowToken        string `json:"FlowToken"`
		} `json:"testnet"`
	} `json:"contractaddresses"`
}

// Own defined config that will adapt based on runtime env
type Configuration struct {
	Contractaddresses struct {
		NonFungibleToken string
		Veolet           string
		FungibleToken    string
		FlowToken        string
	}
	Account struct {
		Address string
		Keys    string
		Chain   string
	}
	Network struct {
		Host  string
		Chain string
	}
	Deployments struct {
		Account []string
	}
}

func GetConfig(flowconfig FlowConfiguration, runtimeenv string) *Configuration {
	config := new(Configuration)
	if runtimeenv == "emulator" {
		config.Contractaddresses.NonFungibleToken = flowconfig.Contractaddresses.Emulator.NonFungibleToken
		config.Contractaddresses.Veolet = flowconfig.Contractaddresses.Emulator.Veolet
		config.Contractaddresses.FungibleToken = flowconfig.Contractaddresses.Emulator.FungibleToken
		config.Contractaddresses.FlowToken = flowconfig.Contractaddresses.Emulator.FlowToken

		config.Account.Address = flowconfig.Accounts.Emulator_account.Address
		config.Account.Keys = flowconfig.Accounts.Emulator_account.Keys
		config.Account.Chain = flowconfig.Accounts.Emulator_account.Chain

		config.Network.Host = flowconfig.Networks.Emulator.Host
		config.Network.Chain = flowconfig.Networks.Emulator.Chain

		config.Deployments.Account = flowconfig.Deployments.Emulator.Emulator_account

	} else if runtimeenv == "testnet" {
		config.Contractaddresses.NonFungibleToken = flowconfig.Contractaddresses.Testnet.NonFungibleToken
		config.Contractaddresses.Veolet = flowconfig.Contractaddresses.Testnet.Veolet
		config.Contractaddresses.FungibleToken = flowconfig.Contractaddresses.Testnet.FungibleToken
		config.Contractaddresses.FlowToken = flowconfig.Contractaddresses.Testnet.FlowToken

		config.Account.Address = flowconfig.Accounts.Testnet_account.Address
		config.Account.Keys = flowconfig.Accounts.Testnet_account.Keys
		config.Account.Chain = flowconfig.Accounts.Testnet_account.Chain

		config.Network.Host = flowconfig.Networks.Testnet.Host
		config.Network.Chain = flowconfig.Networks.Testnet.Chain

		config.Deployments.Account = flowconfig.Deployments.Testnet.Testnet_account
	} else if runtimeenv == "mainnet" {
		//Todo
	} else {
		panic("Invalid runtimeenv")
	}

	return config
}
