import Veolet from 0xVEOLET//0xf8d6e0586b0a20c7

transaction(newURL: String,tokenID: UInt64) {
    prepare(acct: AuthAccount) {
        // Borrow a reference from the stored collection
        let collectionRef = acct.borrow<&Veolet.Collection>(from: /storage/VeoletCollection)
            ?? panic("Could not borrow a reference to the owner's collection")
        collectionRef.setNFTMediaURL(id: tokenID,newMediaURL: newURL)
    }   
}