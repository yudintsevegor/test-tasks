package uniq

func Find(original, withUniq []string) ([]string, bool) {
	uniq := make(map[string]struct{})
	for _, word := range original {
		uniq[word] = struct{}{}
	}

	var (
		out     []string
		notUniq int
	)

	for _, v := range withUniq {
		if _, ok := uniq[v]; !ok {
			out = append(out, v)
			continue
		}
		notUniq++
	}

	if notUniq == len(withUniq) {
		return nil, false
	}

	return out, true
}
