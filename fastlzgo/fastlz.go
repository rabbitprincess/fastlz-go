package fastlzgo

import (
	"errors"
	"unsafe"

	"github.com/rabbitprincess/fastlz-go/fastlzgo/xxhash"
)

func Compress(input []byte) ([]byte, error) {
	length := len(input)
	if length == 0 {
		return nil, errors.New("no input provided")
	}

	result := make([]byte, length*2)
	size := fastlzCompress(input, result)

	if size == 0 {
		return nil, errors.New("error compressing data")
	}

	return result[:size], nil
}

func fastlzCompress(input, output []byte) int {
	if len(input) < 65536 {
		return fastlz1Compress(input, output)
	}
	return fastlz2Compress(input, output)
}

const (
	MaxCopy        = 32
	MaxLen         = 256 + 8
	MaxL1Distance  = 8192
	MaxL2Distance  = 8191
	MaxFarDistance = 65535 + MaxL2Distance - 1
	HashLog        = 13
	HashSize       = 1 << HashLog
	HashMask       = HashSize - 1
)

func xxHashMask(data []byte) uint32 {
	hash := xxhash.Sum64(data)
	return uint32(hash & HashMask)
}

func flzReadU32(p []byte) uint32 {
	return *(*uint32)(unsafe.Pointer(&p[0]))
}

func flzLiterals(runs int, src, dest []byte) []byte {
	for runs >= MaxCopy {
		dest = append(dest, MaxCopy-1)
		dest = append(dest, src[:MaxCopy]...)
		src = src[MaxCopy:]
		runs -= MaxCopy
	}
	if runs > 0 {
		dest = append(dest, byte(runs-1))
		dest = append(dest, src[:runs]...)
	}
	return dest
}

func flzCmp(p, q, r []byte) int {
	n := min(len(p), len(q))
	n = min(n, int(uintptr(unsafe.Pointer(&r[0]))-uintptr(unsafe.Pointer(&p[0]))))

	for i := 0; i < n; i++ {
		if p[i] != q[i] {
			return i
		}
	}
	return n
}
