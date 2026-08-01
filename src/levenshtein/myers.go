package levenshtien

var peq = make([]uint32, 0x10000)

func myers32(a, b string) int {
	n := len(a)
	m := len(b)
	lst := uint32(1 << (n - 1))
	pv := ^uint32(0)
	var mv uint32 = 0
	sc := n

	for i := n - 1; i >= 0; i-- {
		peq[a[i]] |= 1 << i
	}

	for i := 0; i < m; i++ {
		eq := peq[b[i]]
		xv := eq | mv
		eq |= ((eq & pv) + pv) ^ pv
		mv |= ^(eq | pv)
		pv &= eq

		if (mv & lst) != 0 {
			sc++
		}
		if (pv & lst) != 0 {
			sc--
		}

		mv = (mv << 1) | 1
		pv = (pv << 1) | ^(xv | mv)
		mv &= xv
	}

	for i := n - 1; i >= 0; i-- {
		peq[a[i]] = 0
	}

	return sc
}

func myersX(b, a string) int {
	n := len(a)
	m := len(b)

	hsize := (n + 31) / 32
	vsize := (m + 31) / 32

	mhc := make([]uint32, hsize)
	phc := make([]uint32, hsize)

	for i := 0; i < hsize; i++ {
		phc[i] = ^uint32(0)
		mhc[i] = 0
	}

	j := 0
	for ; j < vsize-1; j++ {
		var mv uint32 = 0
		pv := ^uint32(0)

		start := j * 32
		vlen := start + min(32, m-start)

		for k := start; k < vlen; k++ {
			peq[b[k]] |= uint32(1) << uint(k-start)
		}

		for i := 0; i < n; i++ {
			idx := i / 32

			eq := peq[a[i]]
			pb := (phc[idx] >> uint(i&31)) & 1
			mb := (mhc[idx] >> uint(i&31)) & 1

			xv := eq | mv
			xh := ((((eq | mb) & pv) + pv) ^ pv) | eq | mb

			ph := mv | ^(xh | pv)
			mh := pv & xh

			if ((ph >> 31) ^ pb) != 0 {
				phc[idx] ^= uint32(1) << uint(i&31)
			}
			if ((mh >> 31) ^ mb) != 0 {
				mhc[idx] ^= uint32(1) << uint(i&31)
			}

			ph = (ph << 1) | pb
			mh = (mh << 1) | mb

			pv = mh | ^(xv | ph)
			mv = ph & xv
		}

		for k := start; k < vlen; k++ {
			peq[b[k]] = 0
		}
	}

	var mv uint32 = 0
	pv := ^uint32(0)

	start := j * 32
	vlen := start + min(32, m-start)

	for k := start; k < vlen; k++ {
		peq[b[k]] |= uint32(1) << uint(k-start)
	}

	score := m

	for i := 0; i < n; i++ {
		idx := i / 32

		eq := peq[a[i]]
		pb := (phc[idx] >> uint(i&31)) & 1
		mb := (mhc[idx] >> uint(i&31)) & 1

		xv := eq | mv
		xh := ((((eq | mb) & pv) + pv) ^ pv) | eq | mb

		ph := mv | ^(xh | pv)
		mh := pv & xh

		score += int((ph >> uint((m-1)&31)) & 1)
		score -= int((mh >> uint((m-1)&31)) & 1)

		if ((ph >> 31) ^ pb) != 0 {
			phc[idx] ^= uint32(1) << uint(i&31)
		}
		if ((mh >> 31) ^ mb) != 0 {
			mhc[idx] ^= uint32(1) << uint(i&31)
		}

		ph = (ph << 1) | pb
		mh = (mh << 1) | mb

		pv = mh | ^(xv | ph)
		mv = ph & xv
	}

	for k := start; k < vlen; k++ {
		peq[b[k]] = 0
	}

	return score
}
