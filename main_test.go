package main

import (
	"testing"

	"github.com/rabbitprincess/fastlz-go/fastlz"
	"github.com/rabbitprincess/fastlz-go/fastlzgo"
	"github.com/stretchr/testify/require"
)

func TestEncodeDecode(t *testing.T) {
	// encode fastlz, decode fastlzgo
	input := []byte("hello world!")
	enc, err := fastlz.Compress(input)
	require.NoError(t, err)

	dec, err := fastlzgo.Decompress(enc)
	require.NoError(t, err)
	require.Equal(t, input, dec)
}
