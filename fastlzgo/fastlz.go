package fastlzgo

import (
	"errors"
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

func fastlz1Compress(input, output []byte) int {

	return 0
}

func fastlz2Compress(input, output []byte) int {

	return 0
}
