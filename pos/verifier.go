package pos

import (
	"encoding/binary"
	//"fmt"
	"github.com/kwonalbert/spacecoin/util"
	"golang.org/x/crypto/sha3"
)

type Verifier struct {
	pk   []byte // public key to verify the proof
	beta int    // number of challenges needed
	root []byte // root hash

	index int64 // index of the graphy in the family
	size  int64
	pow2  int64
}

func NewVerifier(pk []byte, index int64, beta int, root []byte) *Verifier {
	size := numXi(index)
	log2 := util.Log2(size) + 1
	pow2 := int64(1 << uint64(log2))
	if (1 << uint64(log2-1)) == size {
		pow2 = 1 << uint64(log2-1)
	}

	v := Verifier{
		pk:   pk,
		beta: beta,
		root: root,

		index: index,
		size:  size,
		pow2:  pow2,
	}
	return &v
}

//TODO: need to select based on some pseudorandomness/gamma function?
//      Note that these challenges are different from those of cryptocurrency
func (v *Verifier) SelectChallenges(seed []byte) []int64 {
	challenges := make([]int64, v.beta)
	rands := make([]byte, v.beta*8)
	sha3.ShakeSum256(rands, seed) //PRNG
	for i := range challenges {
		val, num := binary.Uvarint(rands[i*8 : (i+1)*8])
		if num < 0 {
			panic("Couldn't read PRNG")
		}
		challenges[i] = int64(val % uint64(v.size))
	}
	return challenges
}

func (v *Verifier) VerifySpace(challenges []int64, hashes [][]byte, proofs [][][]byte) bool {
	for i := range challenges {
		if !v.Verify(challenges[i], hashes[i], proofs[i]) {
			return false
		}
	}
	return true
}

func (v *Verifier) Verify(node int64, hash []byte, proof [][]byte) bool {
	curHash := hash
	counter := 0
	for i := node + v.pow2; i > 1; i /= 2 {
		var val []byte
		if i%2 == 0 {
			val = append(curHash, proof[counter]...)
		} else {
			val = append(proof[counter], curHash...)
		}
		hash := sha3.Sum256(val)
		curHash = hash[:]
		counter++
	}

	for i := range v.root {
		if v.root[i] != curHash[i] {
			return false
		}
	}

	return len(v.root) == len(curHash)
}
