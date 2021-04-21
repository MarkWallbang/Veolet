package lib

import (
	"bytes"
)

// replace all placeholder addresses with the correct ones
func ReplaceAddressPlaceholders(code []byte, nftAddress string, veoletAddress string, ftAddress string, flowAddress string) []byte {
	code = bytes.ReplaceAll(code, []byte("0xNONFUNGIBLETOKEN"), []byte("0x"+nftAddress))
	code = bytes.ReplaceAll(code, []byte("0xVEOLET"), []byte("0x"+veoletAddress))
	code = bytes.ReplaceAll(code, []byte("0xFUNGIBLETOKEN"), []byte("0x"+ftAddress))
	code = bytes.ReplaceAll(code, []byte("0xFLOW"), []byte("0x"+flowAddress))
	return code
}
