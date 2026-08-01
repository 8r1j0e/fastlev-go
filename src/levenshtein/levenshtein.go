package levenshtien

func Distance(a, b string) int {
	if len(a) < len(b) {
		a, b = b, a // swap
	}

	if len(b) == 0 {
		return len(a)
	}

	if len(a) <= 32 {
		return myers32(a, b)
	}

	return myersX(a, b)
}

func Closest(str string, arr []string) string {
	if len(arr) == 0 {
		return ""
	}

	minIndex := 0
	minDistance := Distance(str, arr[0])

	for i := 1; i < len(arr); i++ {
		dist := Distance(str, arr[i])
		if dist < minDistance {
			minDistance = dist
			minIndex = i
		}
	}

	return arr[minIndex]
}
