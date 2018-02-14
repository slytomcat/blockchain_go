package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"math"
	"runtime"
	"sync/atomic"
)

const maxNonce = math.MaxInt64

const targetBits = 16

// ProofOfWork represents a proof-of-work
type ProofOfWork struct {
	name   *[]byte // pow name = block.Data
	target *[]byte // taget bytes array
	data   *[]byte // packed block data for hash calculation (nonce is the last 8 bytes)
	nonce  int64   // nonce from the source block
}

// NewProofOfWork builds and returns a ProofOfWork
func NewProofOfWork(b *Block) *ProofOfWork {

	target := make([]byte, (targetBits-1)/8+1)
	target[len(target)-1] = 128 >> ((targetBits - 1) % 8)

	// func prepareData() on-lined
	data := bytes.Join(
		[][]byte{
			b.PrevBlockHash,
			b.HashTransactions(),
			IntToHex(b.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)

	return &ProofOfWork{&b.Data, &target, &data, int64(b.Nonce)}
}

// powRes - structure to return hash and nonce from maining go-routines
type powRes struct {
	hash  []byte
	nonce int
}

// Run performs a proof-of-work
func (pow *ProofOfWork) Run() (int, []byte) {

	//fmt.Printf("Mining the block containing \"%s\"\n", *pow.name)
	res := make(chan powRes)
	cpus := int64(runtime.NumCPU())
	var p int64
	// Continue flag for mining go-routines: 0 mean that search is not finished yet.
	// The flag should treated via atomic.* methods to aviod races
	var cont int64
	dl := len(*pow.data) - 8 // Place of the nonce bytes start in the block data image
	tl := len(*pow.target)

	// Create maining go-routnes
	for p = 0; p < cpus; p++ {
		go func(result chan powRes, nonce int64) {
			// Make a local copy of block data image to avoid update races between maining go-routines
			data := make([]byte, dl+8)
			copy(data, *pow.data)

			for atomic.LoadInt64(&cont) == 0 && nonce < maxNonce {

				// Put new nonce into data image
				binary.BigEndian.PutUint64(data[dl:], uint64(nonce))

				hash := sha256.Sum256(data)

				if bytes.Compare(hash[:tl], *pow.target) == -1 {
					result <- powRes{hash[:], int(nonce)}
					atomic.AddInt64(&cont, 1)
					return
				}

				nonce += cpus // New nonce calculated with step equal to CPU count
			}
		}(res, p)
	}
	r := <-res

	return r.nonce, r.hash
}

// Validate validates block's PoW
func (pow *ProofOfWork) Validate() bool {

	binary.BigEndian.PutUint64((*pow.data)[len(*pow.data)-8:], uint64(pow.nonce))

	hash := sha256.Sum256(*pow.data)

	return bytes.Compare(hash[:], *pow.target) == -1
}
