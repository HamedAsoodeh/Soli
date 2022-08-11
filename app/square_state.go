package app

// squareState is used to keep track of the state of a data square as block
// producers fill a block. This is useful for simultaneously prioritizing
// inclusion of MsgPayForData with higher fees, abiding by the non-interactive
// default rules, keeping messages ordered by namespace, and atomically adding a
// tx and its corresponding message to the square.
type squareState struct {
	squareSize        int
	totalShareCount   int
	currentShareCount int
	// sortedTxs is nested slice sorted by namespace.
	sortedTxs []node
}

type node struct {
	shareIndex uint32
	remaining  int
	malleatedTx
}

// fillSquare
// func fillSquare(
// 	start, squareSize int,
// 	prioritizedTxs []malleatedTx,
// ) ([]coretypes.Tx, []uint32, []coretypes.Message) {
// 	remaining := start/squareSize + start%squareSize
// 	ss := squareState{
// 		totalShareCount:   squareSize * squareSize,
// 		currentShareCount: start,
// 		sortedTxs:         []node{{shareIndex: uint32(start), remaining: remaining}},
// 	}
// 	for _, tx := range prioritizedTxs {
// 		i := sort.Search(len(ss.sortedTxs), func(i int) bool {
// 			return bytes.Compare(tx.namespace, ss.sortedTxs[i].namespace) >= 0
// 		})
// 		switch {
// 		case ss.hasRoomInRow(i, tx):
// 			ss.insert(i, tx)
// 		case ss.hasRoomOutOfRow(i, tx):
// 		default:
// 			break
// 		}
// 	}
// }

// // check if the tx can fit
// func (ss *squareState) hasRoomOutOfRow(i int, tx malleatedTx) bool {

// 	// check if we have room to add on an existing row
// 	// if not then check if it can fit in the square
// }

// func (ss *squareState) hasRoomInRow(i int, tx malleatedTx) bool {
// 	return len(tx.msg.Data) < ss.sortedTxs[i-1].remaining
// }

// func (ss *squareSize) insertInExisting

// // insertOrdered inserts the malleated tx into the sorted slice
// func (ss *squareState) insert(i int, tx malleatedTx) {
// 	ss.sortedTxs = append(ss.sortedTxs, node{})
// 	copy(ss.sortedTxs[i+1:], ss.sortedTxs[i:])
// 	ss.sortedTxs[i] = node{malleatedTx: tx}
// 	return
// }
