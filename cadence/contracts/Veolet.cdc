import NonFungibleToken from "./NonFungibleToken.cdc"

pub contract Veolet: NonFungibleToken {
    pub var totalSupply: UInt64

    pub event ContractInitialized()
    pub event Withdraw(id: UInt64, from: Address?)
    pub event Deposit(id: UInt64, to: Address?)

    pub resource NFT: NonFungibleToken.INFT {
        pub let id: UInt64
        pub let originalMediaURL: String
        pub let creatorAddress: Address
        pub let creatorName: String
        pub let createdDate: UInt64

        pub(set) var currentMediaURL: String

        init(initID: UInt64,initMediaURL: String, initCreatorName: String, initCreatorAddress: Address, initCreatedDate: UInt64) {
            self.id = initID
            self.originalMediaURL = initMediaURL
            self.creatorAddress = initCreatorAddress
            self.creatorName = initCreatorName
            self.createdDate = initCreatedDate

            self.currentMediaURL = initMediaURL
        }
    }

    pub resource Collection: NonFungibleToken.Provider, NonFungibleToken.Receiver, NonFungibleToken.CollectionPublic {
        // dictionary of NFT conforming tokens
        // NFT is a resource type with an `UInt64` ID field
        pub var ownedNFTs: @{UInt64: NonFungibleToken.NFT}

        init () {
            self.ownedNFTs <- {}
        }

        pub fun setNFTMediaURL(id: UInt64, newMediaURL: String){
            let changetoken <- self.ownedNFTs.remove(key: id)! as! @Veolet.NFT
            changetoken.currentMediaURL = newMediaURL
            self.deposit(token: <-changetoken)

        }


        // withdraw removes an NFT from the collection and moves it to the caller
        pub fun withdraw(withdrawID: UInt64): @NonFungibleToken.NFT {
            let token <- self.ownedNFTs.remove(key: withdrawID) ?? panic("missing NFT")

            emit Withdraw(id: token.id, from: self.owner?.address)

            return <-token
        }

        // deposit takes a NFT and adds it to the collections dictionary
        // and adds the ID to the id array
        pub fun deposit(token: @NonFungibleToken.NFT) {
            let token <- token as! @Veolet.NFT

            let id: UInt64 = token.id

            // add the new token to the dictionary which removes the old one
            let oldToken <- self.ownedNFTs[id] <- token

            emit Deposit(id: id, to: self.owner?.address)

            destroy oldToken
        }

        // getIDs returns an array of the IDs that are in the collection
        pub fun getIDs(): [UInt64] {
            return self.ownedNFTs.keys
        }

        // borrowNFT gets a reference to an NFT in the collection
        // so that the caller can read its metadata and call its methods
        pub fun borrowNFT(id: UInt64): &NonFungibleToken.NFT {
            return &self.ownedNFTs[id] as &NonFungibleToken.NFT
        }

        destroy() {
            destroy self.ownedNFTs
        }
    }

    // public function that anyone can call to create a new empty collection
    pub fun createEmptyCollection(): @NonFungibleToken.Collection {
        return <- create Collection()
    }

    // Resource that an admin or something similar would own to be
    // able to mint new NFTs
    //
	pub resource NFTMinter {

		// mintNFT mints a new NFT with a new ID
		// and deposit it in the recipients collection using their collection reference
		pub fun mintNFT(recipient: &{NonFungibleToken.CollectionPublic},initMediaURL: String, initCreatorName: String, initCreatorAddress: Address, initCreatedDate: UInt64 ) {

			// create a new NFT
			var newNFT <- create NFT(initID: Veolet.totalSupply,initMediaURL: initMediaURL, initCreatorName: initCreatorName, initCreatorAddress: initCreatorAddress, initCreatedDate: initCreatedDate )

			// deposit it in the recipient's account using their reference
			recipient.deposit(token: <-newNFT)

            Veolet.totalSupply = Veolet.totalSupply + (1 as UInt64)
		}
	}

	init() {
        // Initialize the total supply
        self.totalSupply = 0

        // Create a Collection resource and save it to storage
        let collection <- create Collection()
        self.account.save(<-collection, to: /storage/VeoletCollection)

        // create a public capability for the collection
        self.account.link<&{NonFungibleToken.CollectionPublic}>(
            /public/VeoletCollection,
            target: /storage/VeoletCollection
        )

        // Create a Minter resource and save it to storage
        let minter <- create NFTMinter()
        self.account.save(<-minter, to: /storage/NFTMinter)

        emit ContractInitialized()
	}

}