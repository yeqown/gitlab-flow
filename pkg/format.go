package pkg

import (
	"fmt"
	"strconv"
)

// FormatNum is a function to format number, output string
// e.g.
// FormatNum(123, 3) => "123", FormatNum(123, 5) => "00123"
// if num is over bitSize, return original number string.
func FormatNum(num int, bitSize int) string {
	maxNum := 1
	for i := 0; i < bitSize; i++ {
		maxNum *= 10
	}

	if num >= maxNum {
		return strconv.Itoa(num)
	}

	return fmt.Sprintf("%0"+strconv.Itoa(bitSize)+"d", num)
}
