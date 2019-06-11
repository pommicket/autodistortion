/*
Copyright (C) 2019 Leo Tenenbaum

This file is part of AutoDistortion.

AutoDistortion is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

AutoDistortion is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with AutoDistortion.  If not, see <https://www.gnu.org/licenses/>.
*/
package autodistortion

import (
    "image"
    "image/color"
    "github.com/pommicket/autodistortion/autoutils"
)

func mod(a int, b int) int {
    // Because a % b is the remainder, not modulus!
    return ((a % b) + b) % b
}

type WorkInfo struct {
    Img image.Image
    Xfunction autoutils.Function
    Yfunction autoutils.Function
}

func DistortRows(info *WorkInfo, out [][]color.RGBA, yFrom int, yTo int,
                 done chan struct{}) {

    minX := info.Img.Bounds().Min.X
    maxX := info.Img.Bounds().Max.X
    minY := info.Img.Bounds().Min.Y
    maxY := info.Img.Bounds().Max.Y
    width := maxX - minX
    height := maxY - minY
    fwidth := float64(width)
    fheight := float64(height)

    xy := make([]float64, 2)
    for y := yFrom; y < yTo; y++ {
        xy[1] = (float64(y) - float64(minY)) / fheight
        for x := 0; x < width; x++ {
            xy[0] = float64(x) / fwidth
            srcX := mod(int(info.Xfunction.Evaluate(xy) * fwidth),  width)  + minX
            srcY := mod(int(info.Yfunction.Evaluate(xy) * fheight), height) + minY
            r, g, b, a := info.Img.At(int(srcX), int(srcY)).RGBA()
            out[y][x] = color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8),
                                   uint8(a >> 8)}
        }
    }
    done <- struct{}{}

}

func Distort(img image.Image, functionLen int, nThreads int) image.Image {
    rgba := image.NewRGBA(img.Bounds())

    var xfunction autoutils.Function
    var yfunction autoutils.Function
    xfunction.Generate(2, functionLen)
    yfunction.Generate(2, functionLen)

    minY := img.Bounds().Min.Y
    maxY := img.Bounds().Max.Y
    minX := img.Bounds().Min.X
    maxX := img.Bounds().Max.X
    width := maxX - minX
    height := maxY - minY

    if width == 0 || height == 0 {
        // I don't even know if this is possible, but just in case
        return rgba
    }

    info := WorkInfo{img, xfunction, yfunction}

    // Make a slice of pixels (image.RGBA.Set isn't thread safe ): )
    pixels := make([][]color.RGBA, height)
    done := make(chan struct{})
    for y := 0; y < height; y++ {
        pixels[y] = make([]color.RGBA, width)
    }

    if nThreads > height {
        nThreads = height // Don't make more than one thread per row
    }

    for t := 0; t < nThreads; t++ {
        yFrom := t * (height / nThreads) + minY
        var yTo int
        if t == nThreads - 1 { // Deal with final thread
            yTo = maxY // (go to end of image)
        } else {
            yTo = (t+1) * (height / nThreads) + minY
        }
        go DistortRows(&info, pixels, yFrom, yTo, done)
    }

    for y := 0; y < nThreads; y++ {
        <-done // Wait for all goroutines to finish
    }

    for y := 0; y < height; y++ {
        for x := 0; x < width; x++ {
            rgba.Set(x, y, pixels[y][x])
        }
    }

    return rgba
}
