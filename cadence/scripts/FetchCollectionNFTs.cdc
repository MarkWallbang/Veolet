import NonFungibleToken from 0xNONFUNGIBLETOKEN//0xf8d6e0586b0a20c7
import Veolet from 0xVEOLET//0xf8d6e0586b0a20c7

//Get the NFT collection of an account and return all NFTs
pub fun main(account: Address): [&Veolet.NFT?] {
    let collectionRef = getAccount(account)
        .getCapability(/public/VeoletCollection)
        .borrow<&{Veolet.VeoletGetter}>()
        ?? panic("Could not borrow capability from public collection")

    let ids = collectionRef.getIDs()
    let nfts : [&Veolet.NFT?] = []

    for id in ids {
        nfts.append(collectionRef.borrowVeoletRef(id: id))
}

    return nfts
}