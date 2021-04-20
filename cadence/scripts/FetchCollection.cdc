import NonFungibleToken from 0xNONFUNGIBLETOKEN//0xf8d6e0586b0a20c7

//Get the NFT collection of an account and return all ID's
pub fun main(account: Address): [UInt64] {
    let collectionRef = getAccount(account)
        .getCapability(/public/VeoletCollection)
        .borrow<&{NonFungibleToken.CollectionPublic}>()
        ?? panic("Could not borrow capability from public collection")

    return collectionRef.getIDs()
}