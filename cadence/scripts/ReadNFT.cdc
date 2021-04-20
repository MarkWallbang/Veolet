import NonFungibleToken from 0xNONFUNGIBLETOKEN//0xf8d6e0586b0a20c7
import Veolet from 0xVEOLET//0xf8d6e0586b0a20c7

// This script reads metadata about an NFT in a user's collection
pub fun main(account: Address, id: UInt64):  &Veolet.NFT? {

    // Get the public collection of the owner of the token
    let collectionRef = getAccount(account)
        .getCapability(/public/VeoletCollection)
        .borrow<&{Veolet.VeoletGetter}>()
        ?? panic("Could not borrow capability from public collection")

    // Borrow a reference to a specific NFT in the collection
    let nft = collectionRef.borrowVeoletRef(id: id)
    return nft
}