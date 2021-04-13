import NonFungibleToken from 0xf8d6e0586b0a20c7
import Veolet from 0xf8d6e0586b0a20c7

// This script uses the NFTMinter resource to mint a new NFT
// It must be run with the account that has the minter resource
// stored in /storage/NFTMinter

transaction(initMediaURL: String, initCreatorName: String,addresstest: Address) {

    prepare(signer: AuthAccount) {
        log(signer.address)
    }

    execute {
        log("Logging test transaction")
        log(getAccount(addresstest))
    }
}