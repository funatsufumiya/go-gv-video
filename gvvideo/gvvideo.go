package gvvideo

import (
	"bytes"
	"encoding/binary"
	"errors"
	"image"
	"image/draw"
	"io"
	"os"

	"github.com/pierrec/lz4/v4"
	"github.com/robroyd/dds/decoder/dxt"
)

// GVFormat represents supported formats
const (
	GVFormatDXT1 = 1
	GVFormatDXT3 = 3
	GVFormatDXT5 = 5
)

type GVHeader struct {
	Width      uint32
	Height     uint32
	FrameCount uint32
	FPS        float32
	Format     uint32
	FrameBytes uint32
}

type GVAddressSizeBlock struct {
	Address uint64
	Size    uint64
}

type GVVideo struct {
	Header            GVHeader
	AddressSizeBlocks []GVAddressSizeBlock
	Reader            io.ReadSeeker
}

func ReadHeader(r io.Reader) (GVHeader, error) {
	header := GVHeader{}
	if err := binary.Read(r, binary.LittleEndian, &header.Width); err != nil {
		return header, err
	}
	if err := binary.Read(r, binary.LittleEndian, &header.Height); err != nil {
		return header, err
	}
	if err := binary.Read(r, binary.LittleEndian, &header.FrameCount); err != nil {
		return header, err
	}
	if err := binary.Read(r, binary.LittleEndian, &header.FPS); err != nil {
		return header, err
	}
	if err := binary.Read(r, binary.LittleEndian, &header.Format); err != nil {
		return header, err
	}
	if err := binary.Read(r, binary.LittleEndian, &header.FrameBytes); err != nil {
		return header, err
	}
	return header, nil
}

func ReadAddressSizeBlocks(r io.ReadSeeker, frameCount uint32) ([]GVAddressSizeBlock, error) {
	blocks := make([]GVAddressSizeBlock, frameCount)
	// Seek to address blocks
	if _, err := r.Seek(-int64(frameCount*16), io.SeekEnd); err != nil {
		return nil, err
	}
	for i := range blocks {
		if err := binary.Read(r, binary.LittleEndian, &blocks[i].Address); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.LittleEndian, &blocks[i].Size); err != nil {
			return nil, err
		}
	}
	return blocks, nil
}

func LoadGVVideo(path string) (*GVVideo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	header, err := ReadHeader(f)
	if err != nil {
		return nil, err
	}
	blocks, err := ReadAddressSizeBlocks(f, header.FrameCount)
	if err != nil {
		return nil, err
	}
	return &GVVideo{
		Header:            header,
		AddressSizeBlocks: blocks,
		Reader:            f,
	}, nil
}

// ReadFrame returns RGBA []uint8 for the specified frame
func (v *GVVideo) ReadFrame(frameID uint32) ([]uint8, error) {
	if frameID >= v.Header.FrameCount {
		return nil, errors.New("end of video")
	}
	block := v.AddressSizeBlocks[frameID]
	if _, err := v.Reader.Seek(int64(block.Address), io.SeekStart); err != nil {
		return nil, err
	}
	compressed := make([]byte, block.Size)
	if _, err := io.ReadFull(v.Reader, compressed); err != nil {
		return nil, err
	}
	width := int(v.Header.Width)
	height := int(v.Header.Height)
	uncompressedSize := width * height * 4
	decompressed := make([]byte, uncompressedSize)
	if _, err := lz4.UncompressBlock(compressed, decompressed); err != nil {
		return nil, err
	}
	var fourCC string
	switch v.Header.Format {
	case GVFormatDXT1:
		fourCC = "DXT1"
	case GVFormatDXT3:
		fourCC = "DXT3"
	case GVFormatDXT5:
		fourCC = "DXT5"
	default:
		return nil, errors.New("unsupported format")
	}
	// DXT decode
	decoder, err := dxt.New(fourCC, width, height)
	if err != nil {
		return nil, err
	}
	img, err := decoder.Decode(bytes.NewReader(decompressed))
	if err != nil {
		return nil, err
	}
	if rgbaImg, ok := img.(*image.RGBA); ok {
		return rgbaImg.Pix, nil
	}
	if nrgbaImg, ok := img.(*image.NRGBA); ok {
		// Convert NRGBA to RGBA
		rgba := image.NewRGBA(nrgbaImg.Bounds())
		draw.Draw(rgba, nrgbaImg.Bounds(), nrgbaImg, image.Point{}, draw.Src)
		return rgba.Pix, nil
	}
	return nil, errors.New("not RGBA image")
}
