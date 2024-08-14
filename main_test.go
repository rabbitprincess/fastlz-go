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

func BenchmarkCompress(b *testing.B) {
	b.Run("fastlz cgo [Length 2<<8]", func(b *testing.B) {
		bt := make([]byte, 2<<8)
		for i := 0; i < b.N; i++ {
			_, err := fastlz.Compress(bt)
			require.NoError(b, err)
		}
		b.SetBytes(int64(len(bt)))
	})

	b.Run("fastlz  go [Length 2<<8]", func(b *testing.B) {
		bt := make([]byte, 2<<8)
		for i := 0; i < b.N; i++ {
			_, err := fastlzgo.Compress(bt)
			require.NoError(b, err)
		}
		b.SetBytes(int64(len(bt)))
	})

	b.Run("fastlz cgo [Length 2<<16]", func(b *testing.B) {
		bt := make([]byte, 2<<16)
		for i := 0; i < b.N; i++ {
			_, err := fastlz.Compress(bt)
			require.NoError(b, err)
		}
		b.SetBytes(int64(len(bt)))
	})

	b.Run("fastlz  go [Length 2<<16]", func(b *testing.B) {
		bt := make([]byte, 2<<16)
		for i := 0; i < b.N; i++ {
			_, err := fastlzgo.Compress(bt)
			require.NoError(b, err)
		}
		b.SetBytes(int64(len(bt)))
	})

	b.Run("fastlz cgo [Length 2<<24]", func(b *testing.B) {
		bt := make([]byte, 2<<24)
		for i := 0; i < b.N; i++ {
			_, err := fastlz.Compress(bt)
			require.NoError(b, err)
		}
		b.SetBytes(int64(len(bt)))
	})

	b.Run("fastlz  go [Length 2<<24]", func(b *testing.B) {
		bt := make([]byte, 2<<24)
		for i := 0; i < b.N; i++ {
			_, err := fastlzgo.Compress(bt)
			require.NoError(b, err)
		}
		b.SetBytes(int64(len(bt)))
	})
}
