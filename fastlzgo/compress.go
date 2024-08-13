package fastlzgo

import "errors"

func Compress(input []byte) ([]byte, error) {
	length := len(input)
	if length == 0 {
		return nil, errors.New("no input provided")
	}

	output := make([]byte, length*2)
	size := fastlzCompress(input, length, output)

	if size == 0 {
		return nil, errors.New("error compressing data")
	}

	return output[:size], nil
}

func Decompress(input []byte) ([]byte, error) {
	length := len(input)
	if length == 0 {
		return nil, errors.New("no input provided")
	}

	maxLength := length * 2
	output := make([]byte, maxLength)
	size := fastlzDecompress(input, length, output, maxLength)

	if size == 0 {
		return nil, errors.New("error decompressing data")
	}

	return output[:size], nil
}
