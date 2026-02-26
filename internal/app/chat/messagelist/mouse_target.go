package messagelist

type mouseTarget struct {
	viewX    int
	viewY    int
	blockIdx int
	blockY   int
	hit      bool
}

func projectMouseToView(msgX, msgY, originX, originY int) (viewX, viewY int) {
	viewX = msgX - originX - outerBorderWidth
	viewY = msgY - originY
	return viewX, viewY
}

func resolveMouseTarget(msgX, msgY, originX, originY int, itemAtY func(int) (int, int)) mouseTarget {
	viewX, viewY := projectMouseToView(msgX, msgY, originX, originY)
	blockIdx, blockY := itemAtY(viewY)
	return mouseTarget{
		viewX:    viewX,
		viewY:    viewY,
		blockIdx: blockIdx,
		blockY:   blockY,
		hit:      blockIdx >= 0,
	}
}
