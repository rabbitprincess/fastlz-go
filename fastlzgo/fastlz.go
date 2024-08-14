package fastlzgo

import (
	"math"
	"unsafe"
)

func fastlzCompress(input []byte, output []byte) int {
	if len(input) < 65536 {
		return fastlz1Compress(input, output)
	}
	return fastlz2Compress(input, output)
}

func fastlzDecompress(input []byte, length int, output []byte, maxout int) int {
	/* magic identifier for compression level */
	level := ((*(*uint8)(unsafe.Pointer(&input[0]))) >> 5) + 1

	if level == 1 {
		return fastlz1Decompress(input, length, output, maxout)
	}
	if level == 2 {
		return fastlz2Decompress(input, length, output, maxout)
	}
	/* unknown level, trigger error */
	return 0
}

const (
	MAX_COPY        = 32
	MAX_LEN         = 256 + 8
	MAX_DISTANCE1   = 8192
	MAX_DISTANCE2   = 8191
	MAX_FARDISTANCE = 65535 + MAX_DISTANCE2 - 1
	HASH_LOG        = 13
	HASH_SIZE       = 1 << HASH_LOG
	HASH_MASK       = HASH_SIZE - 1
)

func flzHash(v uint32) uint16 {
	h := (v * 2654435769) >> (32 - HASH_LOG)
	return uint16(h & HASH_MASK)
}

func flzLiterals(length uint, anchor, op []byte) []byte {
	for length >= MAX_COPY {
		op[0] = MAX_COPY - 1
		copy(op[1:], anchor[:MAX_COPY])
		op = op[MAX_COPY+1:]
		anchor = anchor[MAX_COPY:]
		length -= MAX_COPY
	}
	if length > 0 {
		op[0] = byte(length - 1)
		copy(op[1:], anchor[:length])
		op = op[length+1:]
	}
	return op
}

func flz1Match(len uint32, distance uint32, op []byte) []byte {
	distance-- // Decrement distance

	if len > MAX_LEN-2 {
		for len > MAX_LEN-2 {
			op[0] = byte((7 << 5) + (distance >> 8))
			op[1] = MAX_LEN - 2 - 7 - 2
			op[2] = byte(distance & 255)
			op = op[3:] // Move op slice forward by 3 bytes
			len -= MAX_LEN - 2
		}
	}

	if len < 7 {
		op[0] = byte((len << 5) + (distance >> 8))
		op[1] = byte(distance & 255)
		op = op[2:] // Move op slice forward by 2 bytes
	} else {
		op[0] = byte((7 << 5) + (distance >> 8))
		op[1] = byte(len - 7)
		op[2] = byte(distance & 255)
		op = op[3:] // Move op slice forward by 3 bytes
	}

	return op
}

// flzReadU32 reads a 32-bit unsigned integer from the given byte slice.
func flzReadU32(ptr []byte) uint32 {
	return *(*uint32)(unsafe.Pointer(&ptr[0]))
}

// flzCmp compares the byte slices p and q up to r and returns the length of the matching portion.
func flzCmp(p, q, r []byte) int {
	if *(*uint32)(unsafe.Pointer(&p[0])) == *(*uint32)(unsafe.Pointer(&q[0])) {
		p = p[4:]
		q = q[4:]
	}

	for len(q) > 0 && len(p) > 0 && len(q) <= len(r) {
		if p[0] != q[0] {
			break
		}
		p = p[1:]
		q = q[1:]
	}

	// Return the length of the match.
	return len(p) - len(r)
}

func fastlz1Compress(input []byte, output []byte) int {
	ip := 0
	ipStart := 0
	ipBound := len(input) - 4
	ipLimit := len(input) - 12 - 1
	op := 0

	var htab [HASH_SIZE]uint32
	var seq, hash uint32

	// Initialize hash table
	for i := range htab {
		htab[i] = 0
	}

	// Start with literal copy
	anchor := ip
	ip += 2

	// Main loop
	for ip < ipLimit {
		var ref int
		var distance uint32
		var cmp uint32

		// Find potential match
		for {
			if ip+3 >= len(input) {
				break
			}
			seq = flzReadU32(input[ip:]) & 0xffffff
			hash = uint32(flzHash(seq))
			ref = ipStart + int(htab[hash])
			htab[hash] = uint32(ip - ipStart)
			distance = uint32(ip - ref)
			cmp = uint32(math.MaxUint32)
			if distance < MAX_DISTANCE1 {
				cmp = flzReadU32(input[ref:]) & 0xffffff
			}
			if ip >= ipLimit {
				break
			}
			ip++
			if seq == cmp {
				break
			}
		}

		if ip >= ipLimit {
			break
		}
		ip--

		if ip > anchor {
			output = flzLiterals(uint(ip-anchor), input[anchor:ip], output[op:])
			op = len(output)
		}

		matchLen := uint32(flzCmp(input[ref+3:], input[ip+3:], input[ipBound:]))
		output = flz1Match(matchLen, distance, output)
		op = len(output)

		// Update the hash at match boundary
		ip += int(matchLen)
		seq = flzReadU32(input[ip:])
		hash = uint32(flzHash(seq & 0xffffff))
		htab[hash] = uint32(ip - ipStart)
		seq >>= 8
		hash = uint32(flzHash(seq))
		htab[hash] = uint32(ip - ipStart + 1)

		anchor = ip
	}

	copyLen := len(input) - anchor
	output = flzLiterals(uint(copyLen), input[anchor:], output)
	op = len(output)

	return op
}

// Define flz2Match function
func flz2Match(len uint32, distance uint32, op []byte) []byte {
	distance-- // Decrement distance

	if distance < MAX_DISTANCE2 {
		// Near match
		if len > MAX_LEN-2 {
			for len > MAX_LEN-2 {
				op[0] = byte((7 << 5) + (distance >> 8))
				op[1] = MAX_LEN - 2 - 7 - 2
				op[2] = byte(distance & 255)
				op = op[3:] // Move op slice forward by 3 bytes
				len -= MAX_LEN - 2
			}
		}

		if len < 7 {
			op[0] = byte((len << 5) + (distance >> 8))
			op[1] = byte(distance & 255)
			op = op[2:] // Move op slice forward by 2 bytes
		} else {
			op[0] = byte((7 << 5) + (distance >> 8))
			op[1] = byte(len - 7)
			op[2] = byte(distance & 255)
			op = op[3:] // Move op slice forward by 3 bytes
		}
	} else {
		// Far match
		len -= 3
		op[0] = byte((7 << 5) + 31)
		op[1] = byte(len >> 8)
		op[2] = byte(len & 255)
		op[3] = byte(distance >> 8)
		op[4] = byte(distance & 255)
		op = op[5:] // Move op slice forward by 5 bytes
	}

	return op
}

// Define fastlz2Compress function
func fastlz2Compress(input []byte, output []byte) int {
	ip := 0
	ipStart := 0
	ipBound := len(input) - 4
	ipLimit := len(input) - 12 - 1
	op := 0

	var htab [HASH_SIZE]uint32
	var seq, hash uint32

	// Initialize hash table
	for i := range htab {
		htab[i] = 0
	}

	// Start with literal copy
	anchor := ip
	ip += 2

	// Main loop
	for ip < ipLimit {
		var ref int
		var distance uint32
		var cmp uint32

		// Find potential match
		for {
			if ip+3 >= len(input) {
				break
			}
			seq = flzReadU32(input[ip:]) & 0xffffff
			hash = uint32(flzHash(seq))
			ref = ipStart + int(htab[hash])
			htab[hash] = uint32(ip - ipStart)
			distance = uint32(ip - ref)
			cmp = uint32(math.MaxUint32)
			if distance < MAX_FARDISTANCE {
				cmp = flzReadU32(input[ref:]) & 0xffffff
			}
			if ip >= ipLimit {
				break
			}
			ip++
			if seq == cmp {
				break
			}
		}

		if ip >= ipLimit {
			break
		}
		ip--

		// Far match needs at least 5-byte match
		if distance >= MAX_DISTANCE2 {
			if input[ref+3] != input[ip+3] || input[ref+4] != input[ip+4] {
				ip++
				continue
			}
		}

		if ip > anchor {
			output = flzLiterals(uint(ip-anchor), input[anchor:ip], output)
			op = len(output)
		}

		matchLen := uint32(flzCmp(input[ref+3:], input[ip+3:], input[ipBound:]))
		output = flz2Match(matchLen, distance, output)
		op = len(output)

		// Update the hash at match boundary
		ip += int(matchLen)
		seq = flzReadU32(input[ip:])
		hash = uint32(flzHash(seq & 0xffffff))
		htab[hash] = uint32(ip - ipStart)
		seq >>= 8
		hash = uint32(flzHash(seq))
		htab[hash] = uint32(ip - ipStart + 1)

		anchor = ip
	}

	copyLen := len(input) - anchor
	output = flzLiterals(uint(copyLen), input[anchor:], output)
	op = len(output)

	// Marker for fastlz2
	output[0] |= (1 << 5)

	return op
}

func fastlz1Decompress(input []byte, length int, output []byte, maxout int) int {
	var ip uint = 0
	var ip_limit uint = uint(length)
	var op uint = 0
	var op_limit uint = uint(maxout)
	var ctrl uint = uint(input[ip] & 31)
	ip++
	loop := true

	for loop {
		ref := op
		len := ctrl >> 5
		ofs := (ctrl & 31) << 8

		if ctrl >= 32 {
			len--
			ref -= ofs
			if len == 7-1 {
				len += uint(input[ip])
				ip++
			}
			ref -= uint(input[ip])
			ip++

			if op+len+3 > op_limit {
				return 0
			}
			if int(ref-1) < 0 {
				return 0
			}
			if ip < ip_limit {
				ctrl = uint(input[ip])
				ip++
			} else {
				loop = false
			}
			if ref == op {
				/* optimize copy for a run */
				b := output[ref-1]
				output[op] = b
				op++
				output[op] = b
				op++
				output[op] = b
				op++
				for ; len != 0; len-- {
					output[op] = b
					op++
				}
			} else {
				/* copy from reference */
				ref--

				output[op] = output[ref]
				op++
				ref++
				output[op] = output[ref]
				op++
				ref++
				output[op] = output[ref]
				op++
				ref++

				for ; len != 0; len-- {
					output[op] = output[ref]
					op++
					ref++
				}
			}
		} else {
			ctrl++
			if op+ctrl > op_limit {
				return 0
			}
			if ip+ctrl > ip_limit {
				return 0
			}
			output[op] = input[ip]
			op++
			ip++
			for ctrl--; ctrl != 0; ctrl-- {
				output[op] = input[ip]
				op++
				ip++
			}
			loop = (ip < ip_limit)
			if loop {
				ctrl = uint(input[ip])
				ip++
			}
		}
	}

	return int(op)
}

func fastlz2Decompress(input []byte, length int, output []byte, maxout int) int {
	var ip uint = 0
	var ip_limit uint = uint(length)
	var op uint = 0
	var op_limit uint = uint(maxout)
	var ctrl uint = uint(input[ip] & 31)
	ip++
	loop := true

	for loop {
		ref := op
		len := ctrl >> 5
		ofs := (ctrl & 31) << 8

		if ctrl >= 32 {
			var code byte
			len--
			ref -= ofs
			if len == 7-1 {
				code = 255
				for code == 255 {
					code = input[ip]
					ip++
					len += uint(code)
				}
			}
			code = input[ip]
			ip++
			ref -= uint(code)

			/* match from 16-bit distance */
			if code == 255 {
				if ofs == (31 << 8) {
					ofs = uint(input[ip]) << 8
					ip++
					ofs += uint(input[ip])
					ip++
					ref = op - ofs - MAX_DISTANCE2
				}
			}
			if op+len+3 > op_limit {
				return 0
			}
			if int(ref-1) < 0 {
				return 0
			}
			if ip < ip_limit {
				ctrl = uint(input[ip])
				ip++
			} else {
				loop = false
			}
			if ref == op {
				/* optimize copy for a run */
				b := output[ref-1]
				output[op] = b
				op++
				output[op] = b
				op++
				output[op] = b
				op++
				for ; len != 0; len-- {
					output[op] = b
					op++
				}
			} else {
				/* copy from reference */
				ref--

				output[op] = output[ref]
				op++
				ref++
				output[op] = output[ref]
				op++
				ref++
				output[op] = output[ref]
				op++
				ref++

				for ; len != 0; len-- {
					output[op] = output[ref]
					op++
					ref++
				}
			}
		} else {
			ctrl++
			if op+ctrl > op_limit {
				return 0
			}
			if ip+ctrl > ip_limit {
				return 0
			}
			output[op] = input[ip]
			op++
			ip++
			for ctrl--; ctrl != 0; ctrl-- {
				output[op] = input[ip]
				op++
				ip++
			}
			loop = (ip < ip_limit)
			if loop {
				ctrl = uint(input[ip])
				ip++
			}
		}
	}

	return int(op)
}
