package fastlz

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
