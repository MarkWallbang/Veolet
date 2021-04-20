import NonFungibleToken from 0xNONFUNGIBLETOKEN//0xf8d6e0586b0a20c7
import Veolet from 0xVEOLET//0xf8d6e0586b0a20c7

// This script uses the NFTMinter resource to mint a new NFT
// It must be run with the account that has the minter resource
// stored in /storage/NFTMinter

transaction(recipient: Address, initMediaURL: String, initCreatorName: String, initCreatorAddress: Address, initCreatedDate: UInt64, initCaption:String, initHash: String, initEdition: UInt16) {

    // local variable for storing the minter reference
    let minter: &Veolet.NFTMinter

    prepare(signer: AuthAccount) {

        // borrow a reference to the NFTMinter resource in storage
        self.minter = signer.borrow<&Veolet.NFTMinter>(from: /storage/NFTMinter)
            ?? panic("Could not borrow a reference to the NFT minter")
    }

    execute {
        // Borrow the recipient's public NFT collection reference
        let receiver = getAccount(recipient)
            .getCapability(/public/VeoletCollection)
            .borrow<&{NonFungibleToken.CollectionPublic}>()
            ?? panic("Could not get receiver reference to the NFT Collection")

        // Mint the NFT and deposit it to the recipient's collection
        self.minter.mintNFT(recipient: receiver,initMediaURL: initMediaURL, initCreatorName: initCreatorName, initCreatorAddress: initCreatorAddress, initCreatedDate: initCreatedDate, initCaption:initCaption,initHash:initHash,initEdition:initEdition )

        log("Minted new NFT")
    }
}