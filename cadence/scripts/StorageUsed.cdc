// This script checks the storage capacity and usage of an account

pub fun main(address: Address): [UInt64] {
    let account = getAccount(address)
    return [account.storageCapacity,account.storageUsed]
}