package fastlzgo

func flz1Match(length, distance uint32, op []byte) []byte {
	distance--
	if length > MaxLen-2 {
		for length > MaxLen-2 {
			op = append(op, byte((7<<5)+(distance>>8)))
			op = append(op, byte(MaxLen-2-7-2))
			op = append(op, byte(distance&255))
			length -= MaxLen - 2
		}
	}
	if length < 7 {
		op = append(op, byte((length<<5)+(distance>>8)))
		op = append(op, byte(distance&255))
	} else {
		op = append(op, byte((7<<5)+(distance>>8)))
		op = append(op, byte(length-7))
		op = append(op, byte(distance&255))
	}
	return op
}

func fastlz1Compress(input, output []byte) int {
	ip := input
	ipStart := ip
	ipBound := ip[len(ip)-4:]
	ipLimit := ip[len(ip)-12-1:]
	op := output

	var htab [HashSize]uint32

	anchor := ip
	ip = ip[2:]

	for len(ip) >= len(ipLimit) {
		var ref []byte
		var distance, cmp uint32

		// Find potential match
		for {
			seq := flzReadU32(ip) & 0xffffff
			hash := xxHashMask(ip[:4])
			ref = ipStart[htab[hash]:]
			htab[hash] = uint32(len(ipStart) - len(ip))
			distance = uint32(len(ip) - len(ref))
			if distance < MaxL1Distance {
				cmp = flzReadU32(ref) & 0xffffff
			} else {
				cmp = 0x1000000
			}
			if len(ip) >= len(ipLimit) {
				break
			}
			ip = ip[1:]
			if seq == cmp {
				break
			}
		}

		if len(ip) >= len(ipLimit) {
			break
		}
		ip = ip[:len(ip)-1]

		if len(ip) > len(anchor) {
			op = flzLiterals(len(ip)-len(anchor), anchor, op)
		}

		length := flzCmp(ref[3:], ip[3:], ipBound)
		op = flz1Match(uint32(length), distance, op)

		// Update the hash at match boundary
		ip = ip[length:]
		seq := flzReadU32(ip)
		hash := xxHashMask(ip[:4])
		htab[hash] = uint32(len(ipStart) - len(ip))
		seq >>= 8
		hash = xxHashMask(ip[:4])
		htab[hash] = uint32(len(ipStart) - len(ip))

		anchor = ip
	}

	copySize := len(input) - len(anchor)
	op = flzLiterals(copySize, anchor, op)
	return len(op)
}
