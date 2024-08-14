package fastlzgo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodeDeoode(t *testing.T) {
	bt := []byte("hello world")
	enc, err := Compress(bt)
	require.NoError(t, err)
	dec, err := Decompress(enc)
	require.NoError(t, err)
	require.Equal(t, bt, dec)
}

func BenchmarkCompress(b *testing.B) {
	b.Run("Length 2<<8", func(b *testing.B) {
		bt := make([]byte, 2<<8)
		for i := 0; i < b.N; i++ {
			_, err := Compress(bt)
			require.NoError(b, err)
		}
		b.SetBytes(int64(len(bt)))
	})

	b.Run("Length 2<<16", func(b *testing.B) {
		bt := make([]byte, 2<<16)
		for i := 0; i < b.N; i++ {
			_, err := Compress(bt)
			require.NoError(b, err)
		}
		b.SetBytes(int64(len(bt)))
	})

	b.Run("Length 2<<24", func(b *testing.B) {
		bt := make([]byte, 2<<24)
		for i := 0; i < b.N; i++ {
			_, err := Compress(bt)
			require.NoError(b, err)
		}
		b.SetBytes(int64(len(bt)))
	})
}
