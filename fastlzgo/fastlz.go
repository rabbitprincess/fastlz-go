package fastlzgo

import (
	"unsafe"

	"github.com/rabbitprincess/fastlz-go/fastlzgo/xxhash"
)

func fastlzCompress(input []byte, length int, output []byte) int {
	if length < 65536 {
		return fastlz1Compress(input, length, output)
	}
	return fastlz2Compress(input, length, output)
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

func xxHashMask(data []byte) uint {
	hash := xxhash.Sum64(data)
	return uint(hash & HASH_MASK)
}

func flzReadU32(p []byte) uint32 {
	return *(*uint32)(unsafe.Pointer(&p[0]))
}

func flzLiterals(runs int, src, dest []byte) []byte {
	for runs >= MAX_COPY {
		dest = append(dest, MAX_COPY-1)
		dest = append(dest, src[:MAX_COPY]...)
		src = src[MAX_COPY:]
		runs -= MAX_COPY
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

func fastlz1Compress(input []byte, length int, output []byte) int {
	var ip uint = 0
	var ip_bound uint = uint(length - 2)
	var ip_limit uint = uint(length - 12)
	var op uint = 0

	var htab [HASH_SIZE]uint
	var hslot uint
	var hval uint

	var copy uint

	/* sanity check */
	if length < 4 {
		if length != 0 {
			/* create literal copy only */
			output[op] = byte(length - 1)
			op++
			ip_bound++
			for ip <= ip_bound {
				output[op] = input[ip]
				op++
				ip++
			}
			return length + 1
		} else {
			return 0
		}
	}

	/* initializes hash table */
	// do nothing

	/* we start with literal copy */
	copy = 2
	output[op] = MAX_COPY - 1
	op++
	output[op] = input[ip]
	op++
	ip++
	output[op] = input[ip]
	op++
	ip++

	/* main loop */
	for ip < ip_limit {
		var ref uint
		var distance uint

		/* minimum match length */
		var len uint = 3

		/* comparison starting-point */
		anchor := ip

		/* check for a run */
		// do nothing

		/* find potential match */
		hval = xxHashMask(input[ip : ip+3])

		hslot = hval
		ref = htab[hval]

		/* calculate distance to the match */
		distance = anchor - ref

		/* update hash table */
		htab[hslot] = anchor

		/* is this a match? check the first 3 bytes */
		if distance == 0 ||
			(distance >= MAX_DISTANCE1) ||
			input[ref] != input[ip] || input[ref+1] != input[ip+1] || input[ref+2] != input[ip+2] {
			goto literal
		}

		/* last matched byte */
		ref += len
		ip = anchor + len

		/* distance is biased */
		distance--

		if distance == 0 {
			/* zero distance means a run */
			x := input[ip-1]
			for ip < ip_bound {
				if input[ref] != x {
					break
				} else {
					ip++
				}
				ref++
			}
		} else {
			for {
				/* safe because the outer check against ip limit */
				if input[ref] != input[ip] {
					break
				}
				ref++
				ip++
				if input[ref] != input[ip] {
					break
				}
				ref++
				ip++
				if input[ref] != input[ip] {
					break
				}
				ref++
				ip++
				if input[ref] != input[ip] {
					break
				}
				ref++
				ip++
				if input[ref] != input[ip] {
					break
				}
				ref++
				ip++
				if input[ref] != input[ip] {
					break
				}
				ref++
				ip++
				if input[ref] != input[ip] {
					break
				}
				ref++
				ip++
				if input[ref] != input[ip] {
					break
				}
				ref++
				ip++

				for ip < ip_bound {
					if input[ref] != input[ip] {
						break
					}
					ref++
					ip++
				}
				break
			}
			ref++
			ip++
		}
		/* if we have copied something, adjust the copy count */
		if copy != 0 {
			/* copy is biased, '0' means 1 byte copy */
			output[op-copy-1] = byte(copy - 1)
		} else {
			/* back, to overwrite the copy count */
			op--
		}
		/* reset literal counter */
		copy = 0

		/* length is biased, '1' means a match of 3 bytes */
		ip -= 3
		len = ip - anchor

		/* encode the match */
		for len > MAX_LEN-2 {
			output[op] = byte((7 << 5) + (distance >> 8))
			op++
			output[op] = MAX_LEN - 2 - 7 - 2
			op++
			output[op] = byte(distance & 255)
			op++
			len -= MAX_LEN - 2
		}

		if len < 7 {
			output[op] = byte((len << 5) + (distance >> 8))
			op++
			output[op] = byte(distance & 255)
			op++
		} else {
			output[op] = byte((7 << 5) + (distance >> 8))
			op++
			output[op] = byte(len - 7)
			op++
			output[op] = byte(distance & 255)
			op++
		}

		/* update the hash at match boundary */
		hval = xxHashMask(input[ip : ip+3])
		htab[hval] = ip
		ip++
		hval = xxHashMask(input[ip : ip+3])
		htab[hval] = ip
		ip++

		/* assuming literal copy */
		output[op] = MAX_COPY - 1
		op++

		continue

	literal:
		output[op] = input[anchor]
		op++
		anchor++
		ip = anchor
		copy++
		if copy == MAX_COPY {
			copy = 0
			output[op] = MAX_COPY - 1
			op++
		}
	}

	/* left-over as literal copy */
	ip_bound++
	for ip <= ip_bound {
		output[op] = input[ip]
		op++
		ip++
		copy++
		if copy == MAX_COPY {
			copy = 0
			output[op] = MAX_COPY - 1
			op++
		}
	}

	/* if we have copied something, adjust the copy length */
	if copy != 0 {
		output[op-copy-1] = byte(copy - 1)
	} else {
		op--
	}

	return int(op)
}

func fastlz2Compress(input []byte, length int, output []byte) int {
	var ip uint = 0
	var ip_bound uint = uint(length - 2)
	var ip_limit uint = uint(length - 12)
	var op uint = 0

	var htab [HASH_SIZE]uint
	var hslot uint
	var hval uint

	var copy uint

	/* sanity check */
	if length < 4 {
		if length != 0 {
			/* create literal copy only */
			output[op] = byte(length - 1)
			op++
			ip_bound++
			for ip <= ip_bound {
				output[op] = input[ip]
				op++
				ip++
			}
			return length + 1
		} else {
			return 0
		}
	}

	/* initializes hash table */
	// do nothing

	/* we start with literal copy */
	copy = 2
	output[op] = MAX_COPY - 1
	op++
	output[op] = input[ip]
	op++
	ip++
	output[op] = input[ip]
	op++
	ip++

	/* main loop */
	for ip < ip_limit {
		var ref uint
		var distance uint

		/* minimum match length */
		var len uint = 3

		/* comparison starting-point */
		anchor := ip

		/* check for a run */
		if input[ip] == input[ip-1] && (uint(input[ip-1])|uint(input[ip])<<8) == (uint(input[ip+1])|uint(input[ip+2])<<8) {
			distance = 1
			ip += 3
			ref = anchor - 1 + 3
			goto match
		}

		/* find potential match */
		hval = (uint(input[ip]) | uint(input[ip+1])<<8)
		hval ^= (uint(input[ip+1]) | uint(input[ip+2])<<8) ^ (hval >> (16 - HASH_LOG))
		hval &= HASH_MASK

		hslot = hval
		ref = htab[hval]

		/* calculate distance to the match */
		distance = anchor - ref

		/* update hash table */
		htab[hslot] = anchor

		/* is this a match? check the first 3 bytes */
		if distance == 0 ||
			(distance >= MAX_FARDISTANCE) ||
			input[ref] != input[ip] || input[ref+1] != input[ip+1] || input[ref+2] != input[ip+2] {
			goto literal
		}

		/* far, needs at least 5-byte match */
		if distance >= MAX_DISTANCE2 {
			if input[ref+3] != input[ip+3] || input[ref+4] != input[ip+4] {
				goto literal
			}
			len += 2
		}

		ref += len

	match:

		/* last matched byte */
		ip = anchor + len

		/* distance is biased */
		distance--

		if distance == 0 {
			/* zero distance means a run */
			x := input[ip-1]
			for ip < ip_bound {
				if input[ref] != x {
					break
				} else {
					ip++
				}
				ref++
			}
		} else {
			for {
				/* safe because the outer check against ip limit */
				if input[ref] != input[ip] {
					break
				}
				ref++
				ip++
				if input[ref] != input[ip] {
					break
				}
				ref++
				ip++
				if input[ref] != input[ip] {
					break
				}
				ref++
				ip++
				if input[ref] != input[ip] {
					break
				}
				ref++
				ip++
				if input[ref] != input[ip] {
					break
				}
				ref++
				ip++
				if input[ref] != input[ip] {
					break
				}
				ref++
				ip++
				if input[ref] != input[ip] {
					break
				}
				ref++
				ip++
				if input[ref] != input[ip] {
					break
				}
				ref++
				ip++

				for ip < ip_bound {
					if input[ref] != input[ip] {
						break
					}
					ref++
					ip++
				}
				break
			}
			ref++
			ip++
		}
		/* if we have copied something, adjust the copy count */
		if copy != 0 {
			/* copy is biased, '0' means 1 byte copy */
			output[op-copy-1] = byte(copy - 1)
		} else {
			/* back, to overwrite the copy count */
			op--
		}
		/* reset literal counter */
		copy = 0

		/* length is biased, '1' means a match of 3 bytes */
		ip -= 3
		len = ip - anchor

		/* encode the match */
		if distance < MAX_DISTANCE2 {
			if len < 7 {
				output[op] = byte((len << 5) + (distance >> 8))
				op++
				output[op] = byte(distance & 255)
				op++
			} else {
				output[op] = byte((7 << 5) + (distance >> 8))
				op++
				for len -= 7; len >= 255; len -= 255 {
					output[op] = 255
					op++
				}
				output[op] = byte(len)
				op++
				output[op] = byte(distance & 255)
				op++
			}
		} else {
			/* far away, but not yet in the another galaxy... */
			if len < 7 {
				distance -= MAX_DISTANCE2
				output[op] = byte((len << 5) + 31)
				op++
				output[op] = 255
				op++
				output[op] = byte(distance >> 8)
				op++
				output[op] = byte(distance & 255)
				op++
			} else {
				distance -= MAX_DISTANCE2
				output[op] = (7 << 5) + 31
				op++
				for len -= 7; len >= 255; len -= 255 {
					output[op] = 255
					op++
				}
				output[op] = byte(len)
				op++
				output[op] = 255
				op++
				output[op] = byte(distance >> 8)
				op++
				output[op] = byte(distance & 255)
				op++
			}
		}

		/* update the hash at match boundary */
		hval = xxHashMask(input[ip : ip+3])
		htab[hval] = ip
		ip++
		hval = xxHashMask(input[ip : ip+3])
		htab[hval] = ip
		ip++

		/* assuming literal copy */
		output[op] = MAX_COPY - 1
		op++

		continue

	literal:
		output[op] = input[anchor]
		op++
		anchor++
		ip = anchor
		copy++
		if copy == MAX_COPY {
			copy = 0
			output[op] = MAX_COPY - 1
			op++
		}
	}

	/* left-over as literal copy */
	ip_bound++
	for ip <= ip_bound {
		output[op] = input[ip]
		op++
		ip++
		copy++
		if copy == MAX_COPY {
			copy = 0
			output[op] = MAX_COPY - 1
			op++
		}
	}

	/* if we have copied something, adjust the copy length */
	if copy != 0 {
		output[op-copy-1] = byte(copy - 1)
	} else {
		op--
	}

	/* marker for fastlz2 */
	output[0] |= (1 << 5)

	return int(op)
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
