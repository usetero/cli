package messagelist

type projectedItem struct {
	roundIndex int
	height     int
}

type projectedGap struct {
	height       int
	dividerRound int
}

type layoutProjection struct {
	heights              []int
	gaps                 []projectedGap
	trailingDividerRound int
}

func projectItems(entries []blockEntry, blockHeight func(int) int) []projectedItem {
	items := make([]projectedItem, 0, len(entries))
	for i, e := range entries {
		items = append(items, projectedItem{
			roundIndex: e.roundIndex,
			height:     blockHeight(i),
		})
	}
	return items
}

func projectLayout(items []projectedItem, roundActive func(int) bool) layoutProjection {
	n := len(items)
	p := layoutProjection{
		heights:              make([]int, n),
		gaps:                 make([]projectedGap, n),
		trailingDividerRound: -1,
	}
	for i := range p.gaps {
		p.gaps[i].dividerRound = -1
	}
	for i := range items {
		p.heights[i] = items[i].height
		if i == 0 {
			continue
		}
		prev := items[i-1]
		curr := items[i]
		if prev.roundIndex == curr.roundIndex {
			p.gaps[i].height = blockGap
			continue
		}
		g := roundGap
		if !roundActive(prev.roundIndex) {
			g += gapBeforeDivider + dividerHeight
			p.gaps[i].dividerRound = prev.roundIndex
		}
		p.gaps[i].height = g
	}
	if n == 0 {
		return p
	}
	last := items[n-1]
	if !roundActive(last.roundIndex) {
		p.trailingDividerRound = last.roundIndex
	}
	return p
}

func (p layoutProjection) gapHeights() []int {
	g := make([]int, len(p.gaps))
	for i := range p.gaps {
		g[i] = p.gaps[i].height
	}
	return g
}

func (p layoutProjection) trailingHeight() int {
	if p.trailingDividerRound < 0 {
		return 0
	}
	return gapBeforeDivider + dividerHeight
}
