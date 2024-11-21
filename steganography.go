package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
)

type PNGSignature []byte
type CHUNK []byte

const (
	ChunkTypeIHDR ChunkType = "IHDR" // Image Header
	ChunkTypePLTE ChunkType = "PLTE" // Palette Table
	ChunkTypeIDAT ChunkType = "IDAT" // Image Data
	ChunkTypeIEND ChunkType = "IEND" // Image End
	ChunkTypetRNS ChunkType = "tRNS" // Transparency
	ChunkTypecHRM ChunkType = "cHRM" // Chromaticity
	ChunkTypegAMA ChunkType = "gAMA" // Gamma
	ChunkTypeiCCP ChunkType = "iCCP" // ICC Profile
	ChunkTypetEXt ChunkType = "tEXt" // Textual Data
	ChunkTypezTXt ChunkType = "zTXt" // Compressed Textual Data
	ChunkTypeiTXt ChunkType = "iTXt" // International Textual Data
	ChunkTypebKGD ChunkType = "bKGD" // Background Color
	ChunkTypepHYs ChunkType = "pHYs" // Physical Pixel Dimensions
	ChunkTypehIST ChunkType = "hIST" // Image Histogram
	ChunkTypesPLT ChunkType = "sPLT" // Suggested Palette
	ChunkTypetIME ChunkType = "tIME" // Image Last-Modification Time
)

type IHDR struct {
	Length						uint32//4 bytes
	Width             uint32
	Height            uint32
	BitDepth          byte
	ColorType         byte
	CompressionMethod byte
	FilterMethod      byte
	InterlaceMethod   byte
}

func main() {
	//open cover image
	var cover os.File
	err := OpenImage("goodgopher.png", cover)
	if err != nil {
		log.Fatal(err)
		return
	}

	//iterate through data chunks of cover image
	var ihdr IHDR
	err = ReadIHDR(cover, &ihdr)
	if err != nil {
		log.Fatal(err)
		return
	}

	// getIHDR chunk
	fmt.Printf("%+v",ihdr)
	
}

func (*os.File) ReadChunkType () string {
	var typ string


	var typebuf [4]byte
	_, err = file.Read(typebuf[:])
	if err != nil {
		return err
	}

	typ = string(typebuf[:])

	return typ
}

func OpenImage(path string, file *os.File) error{
	// ready first 8 bytes to verify its a valid png file
	pngSig := PNGSignature{137, 80, 78, 71, 13, 10, 26, 10}
	sig := make(PNGSignature, 8)
	good, err := os.Open(path) 
	if err != nil {
		return err
	}
	_, err = good.Read(sig)
	if err != nil {
		return err
	}

	if !bytes.Equal(sig, pngSig){
		return fmt.Errorf("Not an image")
	}
	file = good
}

func ReadIHDR(file *os.File, ihdr *IHDR) error  {
	//length of ihdr always before ihdr and 4 bytes long
	var lenbuf [4]byte
	_, err := file.Read(lenbuf[:])
	if err != nil {
		return err
	}

	chunkLength := uint32(lenbuf[0])<<24 | uint32(lenbuf[1])<<16 | uint32(lenbuf[2])<<8 | uint32(lenbuf[3])

	if chunkLength != 13 {
		return fmt.Errorf("invalid IHDR chunk length: expected 13, got %d", chunkLength)
	}

	// chunk type
	var typebuf [4]byte
	_, err = file.Read(typebuf[:])
	if err != nil {
		return err
	}

	if string(typebuf[:]) != "IHDR" {
		return fmt.Errorf("invalid chunk type: expected IHDR, got %s", string(typebuf[:]))
	}


	//ihdr always 13 bytes long
	var chunkBuf [13]byte
	_, err = file.Read(chunkBuf[:])
	if err != nil {
		return err
	}

	ihdr.Length = uint32(lenbuf[0] << 24) | uint32(lenbuf[1])<<16 | uint32(lenbuf[2])<<8 | uint32(lenbuf[3])
	ihdr.Width = uint32(chunkBuf[0])<<24 | uint32(chunkBuf[1])<<16 | uint32(chunkBuf[2])<<8 | uint32(chunkBuf[3])
	ihdr.Height = uint32(chunkBuf[4])<<24 | uint32(chunkBuf[5])<<16 | uint32(chunkBuf[6])<<8 | uint32(chunkBuf[7])

	ihdr.BitDepth = chunkBuf[8]
	ihdr.ColorType = chunkBuf[9]
	ihdr.CompressionMethod = chunkBuf[10]
	ihdr.FilterMethod = chunkBuf[11]
	ihdr.InterlaceMethod = chunkBuf[12]



	return nil
}