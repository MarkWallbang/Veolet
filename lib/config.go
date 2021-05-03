package lib

import (
	"encoding/json"

	"github.com/gobuffalo/packr"
)

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

func GetConfig(runtimeenv string) *Configuration {
	// Read Flow configurations
	box := packr.NewBox("..")
	byteFile, err := box.Find("flow.json")
	if err != nil {
		panic("Could not read config file")
	}
	//file, _ := os.Open("flow.json")
	//defer file.Close()
	//byteFile, _ := ioutil.ReadAll(file)
	var configuration FlowConfiguration
	json.Unmarshal(byteFile, &configuration)

	config := new(Configuration)
	if runtimeenv == "emulator" {
		config.Contractaddresses.NonFungibleToken = configuration.Contractaddresses.Emulator.NonFungibleToken
		config.Contractaddresses.Veolet = configuration.Contractaddresses.Emulator.Veolet
		config.Contractaddresses.FungibleToken = configuration.Contractaddresses.Emulator.FungibleToken
		config.Contractaddresses.FlowToken = configuration.Contractaddresses.Emulator.FlowToken

		config.Account.Address = configuration.Accounts.Emulator_account.Address
		config.Account.Keys = configuration.Accounts.Emulator_account.Keys
		config.Account.Chain = configuration.Accounts.Emulator_account.Chain

		config.Network.Host = configuration.Networks.Emulator.Host
		config.Network.Chain = configuration.Networks.Emulator.Chain

		config.Deployments.Account = configuration.Deployments.Emulator.Emulator_account

	} else if runtimeenv == "testnet" {
		config.Contractaddresses.NonFungibleToken = configuration.Contractaddresses.Testnet.NonFungibleToken
		config.Contractaddresses.Veolet = configuration.Contractaddresses.Testnet.Veolet
		config.Contractaddresses.FungibleToken = configuration.Contractaddresses.Testnet.FungibleToken
		config.Contractaddresses.FlowToken = configuration.Contractaddresses.Testnet.FlowToken

		config.Account.Address = configuration.Accounts.Testnet_account.Address
		config.Account.Keys = configuration.Accounts.Testnet_account.Keys
		config.Account.Chain = configuration.Accounts.Testnet_account.Chain

		config.Network.Host = configuration.Networks.Testnet.Host
		config.Network.Chain = configuration.Networks.Testnet.Chain

		config.Deployments.Account = configuration.Deployments.Testnet.Testnet_account
	} else if runtimeenv == "mainnet" {
		//Todo
	} else {
		panic("Invalid runtimeenv")
	}

	return config
}
