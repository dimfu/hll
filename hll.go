package hll

import (
	"hash/fnv"
	"math"
	"math/bits"
)

const bitUsed = 64

type HLL struct {
	// registers to remove the effect of outliers and to lower the variance
	// by splitting the input into several subsets.
	//
	// then later on the estimates or data from registers is combined using harmonic mean to deduce
	// an estimate of the cardinality for the entire set.
	registers []uint

	// size of the bucket
	m uint64

	// number of bits used to determine the register index length (4 to 16)
	// using higher bit will lower the error rate of the cardinality estimation.
	p uint8
}

func New(precision uint32) *HLL {
	hll := &HLL{
		// size increase proportionally to precision
		registers: make([]uint, 1<<precision),
		m:         1 << precision,
		p:         uint8(precision),
	}
	return hll
}

func (h *HLL) Add(data []byte) {
	hashval := createHash(data)

	// remaining bits used to calculate rank
	numHashBits := bitUsed - h.p
	registerIndex := hashval >> uint64(numHashBits) // take the hased bit prefix as the idx for the register

	// take the bit suffix to be used for counting the leading zeros
	remainingBits := hashval << uint64(h.p)

	// rank is to meassure how rare the hash is. by how rare, we are checking for
	// how much zero in the hash, if the hash has many zeros the higher the rank value is
	rank := leftmostActiveBit(uint64(remainingBits)) // count the leading zeros in the prefix

	// should update the rank if the previous hash rank is lower than the current hash
	if rank > h.registers[registerIndex] {
		h.registers[registerIndex] = rank
	}
}

func (h *HLL) Count() uint64 {
	sum := 0.
	m := float64(h.m)
	for _, v := range h.registers {
		// if a rare event like N leading zeros happens, it suggests a lot of distinct input must have tried to see it.
		//
		// the probability of seeing N leading zeros in a random bitstring is 1 / 2^N,
		// so to estimate the cardinality from this register, we use 1 / 2^N.
		sum += math.Pow(math.Pow(2, float64(v)), -1)
	}
	// estimate the cardinality using the HyperLogLog formula:
	// - `sum` is the harmonic sum of 2^(-register value) across all registers.
	// - `m` is the number of registers.
	// - The raw estimate is (m * m) / sum, based on how uniformly hash values are distributed.
	// - The constant 0.79402 (α ≈ 0.79402 for m=16) is a bias correction factor from the Durand-Flajolet paper.
	estimate := .79402 * m * m / sum

	if estimate <= float64(h.m)*2.5 {
		var c uint64
		for _, v := range h.registers {
			if v == 0 {
				c++
			}
		}
		// if there is nothing in the registers, use the standard estimator
		if c == 0 {
			return uint64(estimate)
		}

		// otherwise, use linear counting: est = m log(bucket size/register counts)
		fm := float64(h.m)
		return uint64(fm * math.Log(fm/float64(c)))
	}

	return uint64(estimate)
}

func leftmostActiveBit(x uint64) uint {
	return uint(1 + bits.LeadingZeros64(x))
}

// create a 64-bit hash
func createHash(stream []byte) uint64 {
	h := fnv.New64()
	h.Write(stream)
	sum := h.Sum64()
	h.Reset()
	return sum
}
