package hll

import (
	"fmt"
	"testing"
)

func TestHLL(t *testing.T) {
	precision := uint32(16)
	hll := New(precision)
	hll.Add([]byte("Sigma Balls"))
	hll.Add([]byte("Ridho"))
	hll.Add([]byte("Rizki"))
	hll.Add([]byte("Juli"))
	hll.Add([]byte("Juli"))
	hll.Add([]byte("Siti"))
	fmt.Println(hll.Count())
}
