package play

import "sort"

const (
	farmItemCountHistoryMax = 15
	farmItemCountDevPct     = 0.20
)

type farmItemCountFilter struct {
	history []int
	last    int
	hasLast bool
}

var farmItemCountFilters [4]farmItemCountFilter

func resetFarmItemCountFilters() {
	farmItemCountFilters = [4]farmItemCountFilter{}
}

func pushFarmItemCountFilter(idx, raw int, ok bool) (filtered int, filteredOK bool) {
	if idx < 0 || idx >= len(farmItemCountFilters) {
		return raw, ok
	}
	return farmItemCountFilters[idx].push(raw, ok)
}

func (f *farmItemCountFilter) push(raw int, ok bool) (filtered int, filteredOK bool) {
	if !ok {
		if f.hasLast {
			return f.last, true
		}
		return 0, false
	}
	f.history = append(f.history, raw)
	if len(f.history) > farmItemCountHistoryMax {
		f.history = f.history[len(f.history)-farmItemCountHistoryMax:]
	}
	filtered = filterCountsMedianWithinPct(f.history, farmItemCountDevPct)
	f.last = filtered
	f.hasLast = true
	return filtered, true
}

// filterCountsMedianWithinPct 取最近样本中位数，剔除偏差超过 pct 的离群值后再取中位数。
func filterCountsMedianWithinPct(values []int, pct float64) int {
	if len(values) == 0 {
		return 0
	}
	if len(values) == 1 {
		return values[0]
	}
	median := medianInt(values)
	kept := filterCountsWithinPct(values, median, pct)
	if len(kept) == 0 {
		return median
	}
	return medianInt(kept)
}

func filterCountsWithinPct(values []int, center int, pct float64) []int {
	if center == 0 {
		out := make([]int, 0, len(values))
		for _, v := range values {
			if v == 0 {
				out = append(out, v)
			}
		}
		return out
	}
	out := make([]int, 0, len(values))
	for _, v := range values {
		diff := v - center
		if diff < 0 {
			diff = -diff
		}
		if float64(diff)/float64(center) <= pct {
			out = append(out, v)
		}
	}
	return out
}

func medianInt(values []int) int {
	if len(values) == 0 {
		return 0
	}
	cp := append([]int(nil), values...)
	sort.Ints(cp)
	mid := len(cp) / 2
	if len(cp)%2 == 1 {
		return cp[mid]
	}
	return (cp[mid-1] + cp[mid]) / 2
}
