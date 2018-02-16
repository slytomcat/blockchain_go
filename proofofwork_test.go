package main

import (
	"bytes"
	"crypto/sha256"
	"math/big"
	"testing"
	"time"
)

// *T suffixed - original code for benchmark comparison

type ProofOfWorkT struct {
	block  *Block
	target *big.Int
}

// NewProofOfWork builds and returns a ProofOfWork
func NewProofOfWorkT(b *Block) *ProofOfWorkT {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWorkT{b, target}

	return pow
}

func (pow *ProofOfWorkT) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.HashTransactions(),
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)

	return data
}

// Run performs a proof-of-work
func (pow *ProofOfWorkT) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	//fmt.Printf("Mining a new block")
	for nonce < maxNonce {
		data := pow.prepareData(nonce)

		hash = sha256.Sum256(data)
		//if math.Remainder(float64(nonce), 100000) == 0 {
		//	fmt.Printf("\r%x", hash)
		//}
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}
	//fmt.Print("\n\n")

	return nonce, hash[:]
}

// Validate validates block's PoW
func (pow *ProofOfWorkT) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}

var testBlock = &Block{time.Now().Unix(), []*Transaction{NewCoinbaseTX(`13Uu7B1vDP4ViXqHFsWtbraM3EfQ3UkWXt`, "Test block")}, []byte{}, []byte{}, 0, 0}

func BenchmarkPowT(b *testing.B) {
	//fmt.Printf("\n%#v", block)
	pow := NewProofOfWorkT(testBlock)
	for i := 0; i < b.N; i++ {
		_, _ = pow.Run()
	}
}

func BenchmarkPow(b *testing.B) {
	pow := NewProofOfWork(testBlock)
	for i := 0; i < b.N; i++ {
		_, _ = pow.Run()
	}
}

func TestPow(t *testing.T) {
	testBlock.Nonce, testBlock.Hash = NewProofOfWork(testBlock).Run()
	if !NewProofOfWork(testBlock).Validate() {
		t.Errorf("\nPOW is not valid %d, %x\n", testBlock.Nonce, testBlock.Hash)
	}
}

func TestPowT(t *testing.T) {
	testBlock.Nonce, testBlock.Hash = NewProofOfWorkT(testBlock).Run()
	if !NewProofOfWorkT(testBlock).Validate() {
		t.Errorf("\nPOW is not valid %d, %x\n", testBlock.Nonce, testBlock.Hash)
	}
}

func TestPows(t *testing.T) {
	for i := 0; i < 10; i++ {
		var bl0, bl1 Block
		bl0 = *(testBlock)
		bl1 = *(testBlock)

		bl0.Nonce, bl0.Hash = NewProofOfWorkT(&bl0).Run()
		bl1.Nonce, bl1.Hash = NewProofOfWork(&bl1).Run()

		if bl0.Nonce != bl1.Nonce {
			t.Logf("\n%d: Nonces/hashes not much \nOrig: %d, %x \nNew:  %d, %x\n", i, bl0.Nonce, bl0.Hash, bl1.Nonce, bl1.Hash)
		}
	}
}
