package levenshtien

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
