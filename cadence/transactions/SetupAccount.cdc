
import Veolet from 0xf8d6e0586b0a20c7
import NonFungibleToken from 0xf8d6e0586b0a20c7
// This transaction configures a user's account

// to use the NFT contract by creating a new empty collection,
// storing it in their account storage, and publishing a capability
transaction {
    prepare(acct: AuthAccount) {

        // Return early if the account already has a collection
        if acct.borrow<&Veolet.Collection>(from: /storage/NFTCollection) != nil {
            return
        }

        // Create a new empty collection
        let collection <- Veolet.createEmptyCollection()

        // save it to the account
        acct.save(<-collection, to: /storage/NFTCollection)

        // create a public capability for the collection
        acct.link<&{NonFungibleToken.CollectionPublic}>(
            /public/NFTCollection,
            target: /storage/NFTCollection
        )
    }
}