package test

import (
	"fmt"
	"os"
	"testing"
)

const (
	runtimeenv = "emulator"
)

/*
	This script tests all core functions of the Veolet contract and its transactions/scripts. The test functions will use the package "lib", which
	defines functions to send transactions, create and setup accounts, mint tokens etc.
*/

func TestDeployVeoletContract(t *testing.T) {
	// Should be able to deploy Veolet contract (and the NFT core contract if runtimeenv = emulator)
	runtimeenv := os.Getenv("veolettestenv")
	fmt.Print(runtimeenv)
}

func TestSetupAccount(t *testing.T) {
	// Should be able to create new Flow account and Veolet vault for the new account
}

func TestMintToken(t *testing.T) {
	// Should be able to mint new token and send to different account adresses
	// It will be tested to send tokens to adresses with and without a valid Veolet vault
}

func TestSendToken(t *testing.T) {
	// Should be able to send tokens from one adress to another
}
