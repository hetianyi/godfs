package util

import (
	"crypto/md5"
	"encoding/hex"
	"hash"
	"hash/crc32"
)

var tablePolynomial *crc32.Table

func init() {
	//Create the table with the given polynomial
	tablePolynomial = crc32.MakeTable(crc32.IEEE)
}

func CreateCrc32Hash() hash.Hash32 {
	//Open a new hash interface to write the file to
	return crc32.New(tablePolynomial)
}

func GetCrc32HashString(h hash.Hash32) string {
	//Generate the hash
	hashInBytes := h.Sum(nil)[:]
	//Encode the hash to a string
	return hex.EncodeToString(hashInBytes)
}

func CreateMd5Hash() hash.Hash {
	return md5.New()
}

func GetMd5HashString(h hash.Hash) string {
	hashInBytes := h.Sum(nil)
	return hex.EncodeToString(hashInBytes)
}
