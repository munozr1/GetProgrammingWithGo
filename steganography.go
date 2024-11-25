package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"log"
	"os"
)

type PNGSignature []byte

const (
	ChunkTypeIHDR string = "IHDR" // Image Header
	ChunkTypePLTE string = "PLTE" // Palette Table
	ChunkTypeIDAT string = "IDAT" // Image Data
	ChunkTypeIEND string = "IEND" // Image End
	ChunkTypetRNS string = "tRNS" // Transparency
	ChunkTypecHRM string = "cHRM" // Chromaticity
	ChunkTypegAMA string = "gAMA" // Gamma
	ChunkTypeiCCP string = "iCCP" // ICC Profile
	ChunkTypetEXt string = "tEXt" // Textual Data
	ChunkTypezTXt string = "zTXt" // Compressed Textual Data
	ChunkTypeiTXt string = "iTXt" // International Textual Data
	ChunkTypebKGD string = "bKGD" // Background Color
	ChunkTypepHYs string = "pHYs" // Physical Pixel Dimensions
	ChunkTypehIST string = "hIST" // Image Histogram
	ChunkTypesPLT string = "sPLT" // Suggested Palette
	ChunkTypetIME string = "tIME" // Image Last-Modification Time
)

type ImageData struct{
	ihdr		IHDR
	data		[]byte
}

type IHDR struct {
	Length						uint32//4 bytes
	Width             uint32
	Height            uint32
	CRC								uint32
	BitDepth          byte
	ColorType         byte
	CompressionMethod byte
	FilterMethod      byte
	InterlaceMethod   byte
}
type pHYs struct {
	X								uint32
	Y								uint32
	CRC							uint32
	UnitSpecifier		byte
}

type IDAT struct {
	Data					bytes.Buffer
	Length				uint32
}


func main() {
	//coverImagePixels, err := ImagePixels("badgopher.png")
	cover, err := ImagePixels("good_gopher.png")
	if err != nil {
		log.Fatalf("Failed to process image: %v", err)
	}

	//secretImagePixels, err := ImagePixels("badgopher.png")
	secret, err := ImagePixels("evil_gopher.png")
	if err != nil {
		log.Fatalf("Failed to process image: %v", err)
	}
	log.Println("Successfully extracted pixel data:")

	compatible := compatibility(&cover, &secret)

	if !compatible{
		log.Fatal("Images not compatible")
	}
	fmt.Println("Images are compatible")
}


func compatibility(cover, secret *ImageData) bool {
    widthMatch := cover.ihdr.Width == secret.ihdr.Width
    heightMatch := cover.ihdr.Height == secret.ihdr.Height
    filterMatch := cover.ihdr.FilterMethod == secret.ihdr.FilterMethod
    colorMatch := cover.ihdr.ColorType == secret.ihdr.ColorType

    // All properties must match for compatibility
    return widthMatch && heightMatch && filterMatch && colorMatch
}


func ImagePixels(filePath string) (ImageData, error) {
	var ihdr IHDR
	var idat IDAT
	var imgData ImageData

	cover, err := OpenImage(filePath)
	if err != nil {
		return imgData, fmt.Errorf("failed to open image: %v", err)
	}
	defer cover.Close()

	err = ReadIHDR(cover, &ihdr)
	if err != nil {
		return imgData, fmt.Errorf("failed to read IHDR: %v", err)
	}
	fmt.Printf("IHDR: %+v\n", ihdr)

	//process the PNG chunks
	for {
		length, err := ReadChunkLength(cover)
		if err != nil {
			return imgData, fmt.Errorf("error reading chunk length: %v", err)
		}

		typ, err := ReadChunkType(cover)
		if err != nil {
			return imgData, fmt.Errorf("error reading chunk type: %v", err)
		}

		switch typ {
		case ChunkTypeIDAT:
			err = ReadIDAT(cover, &idat, length)
			if err != nil {
				return imgData, fmt.Errorf("error processing IDAT chunk: %v", err)
			}
		case ChunkTypeIEND:
			fmt.Println("Reached IEND chunk, finishing processing.")
			break
		default:
			//skip unknown chunks
			if _, err := cover.Seek(int64(length+4), io.SeekCurrent); err != nil {
				return imgData, fmt.Errorf("error seeking past unknown chunk: %v", err)
			}
		}

		if typ == ChunkTypeIEND {
			break
		}
	}

	//decompress IDAT data
	rawPNGData, err := ZDecompress(&idat)
	if err != nil {
		return imgData, fmt.Errorf("failed to decompress IDAT data: %v", err)
	}

	//bytes per pixel
	bpp := (int(ihdr.BitDepth) * int(ihdr.ColorType)) / 8

	rawPixelData, err := ApplyFilterMethod(rawPNGData, int(ihdr.Width), int(ihdr.Height), bpp)
	if err != nil {
		return imgData, fmt.Errorf("error processing scanlines: %v", err)
	}
	imgData.ihdr = ihdr
	imgData.data = rawPixelData
	return imgData, nil
}


func ApplyFilterMethod(data []byte, width int, height int, bytesPerPixel int) ([]byte, error) {
	//TODO: may have different filter methods other than 0
	var pixels []byte
	var bytesPerRow = width * bytesPerPixel
  var scanlineLength = bytesPerRow + 1 

	//iterate through the rows of the image
	for i := 0; i < height; i++ {
        start := i * scanlineLength
				if start+scanlineLength > len(data) {
					// Ensure the data is large enough to contain this scanline
					fmt.Printf("Skipping row %d: insufficient data for scanline\n", i)
					continue
				}
        filterByte := data[start]

        if filterByte > 4 {
            fmt.Errorf("unsupported filter byte at row %d: expected 0, got %d", i, filterByte)
						continue
        }

        //extract raw pixel data from the scanline
				//Assume filter method 0
        rawScanline := data[start+1 : start+1+bytesPerRow]
        pixels = append(pixels, rawScanline...)
    }

    return pixels, nil


}


func ZDecompress(dataBuf *IDAT) ([]byte, error) {
    var decompressedData bytes.Buffer

    r, err := zlib.NewReader(&dataBuf.Data)
    if err != nil {
        return nil, fmt.Errorf("failed to create zlib reader: %v", err)
    }
    defer r.Close()

    _, err = io.Copy(&decompressedData, r)
    if err != nil {
        return nil, fmt.Errorf("failed to decompress data: %v", err)
    }

    return decompressedData.Bytes(), nil
}



func OpenImage(path string ) (*os.File, error){
	// ready first 8 bytes to verify its a valid png file
	pngSig := PNGSignature{137, 80, 78, 71, 13, 10, 26, 10}
	sig := make(PNGSignature, 8)
	good, err := os.Open(path) 
	if err != nil {
		return nil, err
	}
	_, err = good.Read(sig)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(sig, pngSig){
		return nil, fmt.Errorf("Not an image")
	}

	return good, nil
}

func ReadIDAT(file *os.File, idat *IDAT, length uint32) error {
    newDataBuf := make([]byte, length)
    count, err := file.Read(newDataBuf)
    if err != nil {
        return err
    }
    if uint32(count) != length {
        return fmt.Errorf("bytes read not equal to IDAT length: (%v,%v)", count, idat.Length)
    }

    //idat.Data = append(idat.Data, newDataBuf...)
		_, err = idat.Data.Write(newDataBuf)
    if err != nil {
        return fmt.Errorf("error writing to IDAT buffer: %v", err)
    }
		idat.Length += uint32(count)
		// TODO: verify crc
		_, err = ReadChunkCRC(file)
		if err != nil{
			return err
		}


    return nil
}

func ReadpHYs(file *os.File,p *pHYs) error  {
	//length of pHYs always 9 
	var chunkBuf [9]byte
	_, err := file.Read(chunkBuf[:])
	if err != nil {
		return err
	}

	
	p.X = uint32(chunkBuf[0])<<24 | uint32(chunkBuf[1])<<16 | uint32(chunkBuf[2])<<8 | uint32(chunkBuf[3])
	p.Y = uint32(chunkBuf[4])<<24 | uint32(chunkBuf[5])<<16 | uint32(chunkBuf[6])<<8 | uint32(chunkBuf[7])
	p.UnitSpecifier = chunkBuf[8]
	crc, err:= ReadChunkCRC(file)
	if err != nil{
		return err
	}

	p.CRC = crc


	return nil
}



func ReadIHDR(file *os.File, ihdr *IHDR) error  {
	//length of ihdr always before ihdr and 4 bytes long
	chunkLength, err := ReadChunkLength(file)
	if err != nil {
		return fmt.Errorf("ReadIHDR() => %v")
	}

	if chunkLength != 13 {
		return fmt.Errorf("invalid IHDR chunk length: expected 13, got %d", chunkLength)
	}

		typ, err:= ReadChunkType(file)
	if err != nil{
		return err
	}

	if typ != "IHDR" {
		return fmt.Errorf("invalid chunk type: expected IHDR, got %s", typ)
	}


	//ihdr always 13 bytes long
	var chunkBuf [13]byte
	_, err = file.Read(chunkBuf[:])
	if err != nil {
		return err
	}

	ihdr.Length = chunkLength
	ihdr.Width = uint32(chunkBuf[0])<<24 | uint32(chunkBuf[1])<<16 | uint32(chunkBuf[2])<<8 | uint32(chunkBuf[3])
	ihdr.Height = uint32(chunkBuf[4])<<24 | uint32(chunkBuf[5])<<16 | uint32(chunkBuf[6])<<8 | uint32(chunkBuf[7])

	ihdr.BitDepth = chunkBuf[8]
	ihdr.ColorType = chunkBuf[9]
	ihdr.CompressionMethod = chunkBuf[10]
	ihdr.FilterMethod = chunkBuf[11]
	ihdr.InterlaceMethod = chunkBuf[12]
	crc, err:= ReadChunkCRC(file)
	if err != nil{
		return err
	}

	ihdr.CRC = crc


	return nil
}


func ReadChunkLength(file *os.File) (uint32, error){
	var length [4]byte
	_, err := file.Read(length[:])

	if err != nil{
		return 0, fmt.Errorf("err reading chunk length: %v", err)
	}

	ret := uint32(length[0] << 24) | uint32(length[1])<<16 | uint32(length[2])<<8 | uint32(length[3])
	return ret, nil
}

func ReadChunkCRC(file *os.File) (uint32, error){
	var crc [4]byte
	_, err := file.Read(crc[:])

	if err != nil{
		return 0, fmt.Errorf("err reading chunk crc: %v", err)
	}

	ret := uint32(crc[0] << 24) | uint32(crc[1])<<16 | uint32(crc[2])<<8 | uint32(crc[3])
	return ret, nil
}

func ReadChunkType(file *os.File) (string , error){
	var typ [4]byte
	_, err := file.Read(typ[:])

	if err != nil{
		return "", fmt.Errorf("err reading chunk typ: %v", err)
	}

	return string(typ[:]), nil
}