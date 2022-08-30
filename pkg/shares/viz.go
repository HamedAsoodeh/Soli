package shares

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"

	"github.com/tendermint/tendermint/pkg/consts"
)

func VisualizeSquare(file string, pixelWidth int, data [][]byte) error {

	myimage := image.NewRGBA(image.Rect(0, 0, pixelWidth, pixelWidth))

	squareSize := int(math.Sqrt(float64(len(data))))
	fmt.Println(squareSize)

	sharePixelWidth := int(pixelWidth / squareSize)

	for i, share := range data {
		rowIndex := i / squareSize
		colIndex := i % squareSize
		color := determintColor(share)
		draw.Draw(
			myimage,
			image.Rect(
				colIndex*sharePixelWidth,
				rowIndex*sharePixelWidth,
				(colIndex+1)*sharePixelWidth,
				(rowIndex+1)*sharePixelWidth,
			),
			&image.Uniform{color}, image.Point{}, draw.Src)
	}

	myfile, err := os.Create(file)
	if err != nil {
		return err
	}
	defer myfile.Close()
	err = png.Encode(myfile, myimage) // ... save image
	if err != nil {
		return err
	}
	fmt.Println("saved", file) // view image issue : firefox  /tmp/chessboard.png
	return nil
}

var (
	txNSColor     = color.RGBA{39, 54, 183, 255}
	evdNSColor    = color.RGBA{255, 251, 51, 255}
	tailNSColor   = color.RGBA{192, 192, 192, 255}
	paddingColor  = color.RGBA{72, 189, 232, 255}
	msg1Color     = color.RGBA{50, 205, 50, 255}
	msg2Color     = color.RGBA{50, 205, 50, 255}
	paddedNSShare = bytes.Repeat([]byte{0}, consts.MsgShareSize)
)

func determintColor(share []byte) color.RGBA {
	isPadding := bytes.Equal(paddedNSShare, share[consts.NamespaceSize:])
	ns := share[:consts.NamespaceSize]
	switch {
	case bytes.Equal(ns, consts.TailPaddingNamespaceID):
		return tailNSColor
	case isPadding:
		return paddingColor
	case bytes.Equal(ns, consts.TxNamespaceID):
		return txNSColor
	case bytes.Equal(ns, consts.EvidenceNamespaceID):
		return evdNSColor
	default:
		return msg1Color
	}

}
