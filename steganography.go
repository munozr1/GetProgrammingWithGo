package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
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
	cover, err := ImagePixels("good_gopher.png")
	if err != nil {
		log.Fatalf("Failed to process cover image: %v", err)
	}

	secret, err := ImagePixels("evil_gopher.png")
	if err != nil {
		log.Fatalf("Failed to process secret image: %v", err)
	}


	compatible, errMsg := compatibility(&cover, &secret)
	if !compatible {
		log.Fatalf("Images are not compatible: %s", errMsg)
	}
	fmt.Println("Images are compatible")

	/*
	bpp := (int(cover.ihdr.BitDepth) * int(cover.ihdr.ColorType)) / 8
	expectedDataSize := int(cover.ihdr.Height) * (int(cover.ihdr.Width)*bpp + 1)
	if len(cover.data) != expectedDataSize || len(secret.data) != expectedDataSize {
			log.Fatalf("Data size mismatch: Expected %d bytes, got cover=%d, secret=%d",
					expectedDataSize, len(cover.data), len(secret.data))
	}
					*/


	samples := samplesPerPixel(cover.ihdr.ColorType)
	if samples == 0 {
			log.Fatalf("Unsupported color type: %d", cover.ihdr.ColorType)
	}
	bitsPerPixel := uint32(samples) * uint32(cover.ihdr.BitDepth)
	//bytesPerPixel := (bitsPerPixel + 7) / 8 // Round up to the nearest byte

	var newImage []byte

	for i := uint32(0); i < uint32(cover.ihdr.Height); i++ {
		rowStart := i * (1 + ((bitsPerPixel*cover.ihdr.Width + 7) / 8)) // Include filter byte
		filterByte := cover.data[rowStart] // First byte of the row
		newImage = append(newImage, filterByte)

		pixelDataStart := rowStart + 1
		rowWidthInBytes := ((bitsPerPixel * cover.ihdr.Width) + 7) / 8

		for j := uint32(0); j < rowWidthInBytes; j++ {
				cByte := cover.data[pixelDataStart+j]
				sByte := secret.data[pixelDataStart+j]

				cMSB := cByte & 0xF0
				sLSB := (sByte & 0xF0) >> 4
				nByte := cMSB | sLSB

				newImage = append(newImage, nByte)
		}
	}
	outputFile, err := os.Create("stego_image.png")
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer outputFile.Close()

	var buf bytes.Buffer

	// Write PNG signature
	buf.Write([]byte{137, 80, 78, 71, 13, 10, 26, 10})

	// Write IHDR chunk
	writeIHDRChunk(&buf, cover.ihdr)

	// Write IDAT chunk
	writeIDATChunk(&buf, newImage)

	// Write IEND chunk
	writeIENDChunk(&buf)

	// Save the file
	_, err = outputFile.Write(buf.Bytes())
	if err != nil {
		log.Fatalf("Failed to write output PNG: %v", err)
	}

	fmt.Println("Stego image successfully created as stego_image.png")
}

func writeIHDRChunk(buf *bytes.Buffer, ihdr IHDR) {
	var chunkData bytes.Buffer

	writeUint32(&chunkData, ihdr.Width)
	writeUint32(&chunkData, ihdr.Height)
	chunkData.WriteByte(ihdr.BitDepth)
	chunkData.WriteByte(ihdr.ColorType)
	chunkData.WriteByte(ihdr.CompressionMethod)
	chunkData.WriteByte(ihdr.FilterMethod)
	chunkData.WriteByte(ihdr.InterlaceMethod)

	writeChunk(buf, ChunkTypeIHDR, chunkData.Bytes())
}

func writeIDATChunk(buf *bytes.Buffer, imageData []byte) {
	var compressedData bytes.Buffer
	writer := zlib.NewWriter(&compressedData)
	_, err := writer.Write(imageData)
	if err != nil {
		log.Fatalf("Failed to compress IDAT data: %v", err)
	}
	writer.Close()

	writeChunk(buf, ChunkTypeIDAT, compressedData.Bytes())
}

func writeIENDChunk(buf *bytes.Buffer) {
	writeChunk(buf, ChunkTypeIEND, nil)
}

func writeChunk(buf *bytes.Buffer, chunkType string, data []byte) {
	writeUint32(buf, uint32(len(data)))
	buf.WriteString(chunkType)
	buf.Write(data)

	crc := calculateCRC(chunkType, data)
	writeUint32(buf, crc)
}

func writeUint32(buf *bytes.Buffer, value uint32) {
	buf.WriteByte(byte(value >> 24))
	buf.WriteByte(byte((value >> 16) & 0xFF))
	buf.WriteByte(byte((value >> 8) & 0xFF))
	buf.WriteByte(byte(value & 0xFF))
}


func generateCRCTable() [256]uint32 {
	const crc32Polynomial uint32 = 0xEDB88320
	var table [256]uint32
	for i := 0; i < 256; i++ {
			crc := uint32(i)
			for j := 0; j < 8; j++ {
					if crc&1 != 0 {
							crc = (crc >> 1) ^ crc32Polynomial
					} else {
							crc >>= 1
					}
			}
			table[i] = crc
	}
	return table
}

func calculateCRC(chunkType string, data []byte) uint32 {
	var crcTable = generateCRCTable()
	crc := uint32(0xFFFFFFFF)
	for _, b := range []byte(chunkType) {
		crc = crcTable[(crc^uint32(b))&0xFF] ^ (crc >> 8)
	}
	for _, b := range data {
		crc = crcTable[(crc^uint32(b))&0xFF] ^ (crc >> 8)
	}
	return crc ^ 0xFFFFFFFF
}

func compatibility(cover, secret *ImageData) (bool, string) {
    widthMatch := cover.ihdr.Width == secret.ihdr.Width
    if !widthMatch {
        return false, fmt.Sprintf("Width mismatch: cover=%d, secret=%d", cover.ihdr.Width, secret.ihdr.Width)
    }

    heightMatch := cover.ihdr.Height == secret.ihdr.Height
    if !heightMatch {
        return false, fmt.Sprintf("Height mismatch: cover=%d, secret=%d", cover.ihdr.Height, secret.ihdr.Height)
    }

    // Correct bpp calculation
    samplesCover := samplesPerPixel(cover.ihdr.ColorType)
    samplesSecret := samplesPerPixel(secret.ihdr.ColorType)
    if samplesCover == 0 || samplesSecret == 0 {
        return false, "Unsupported color type"
    }
    bitsPerPixelCover := int(cover.ihdr.BitDepth) * samplesCover
    bitsPerPixelSecret := int(secret.ihdr.BitDepth) * samplesSecret

    if bitsPerPixelCover != bitsPerPixelSecret {
        return false, fmt.Sprintf("Bits per pixel mismatch: cover=%d, secret=%d", bitsPerPixelCover, bitsPerPixelSecret)
    }

    // Calculate bytes per scanline (including filter byte)
    bytesPerScanline := 1 + ((bitsPerPixelCover*int(cover.ihdr.Width) + 7) / 8)
    expectedSize := int(cover.ihdr.Height) * bytesPerScanline

    dataSizeMatch := len(cover.data) == expectedSize && len(secret.data) == expectedSize
    if !dataSizeMatch {
        return false, fmt.Sprintf(
            "Data size mismatch: expected=%d bytes, got cover=%d, secret=%d",
            expectedSize, len(cover.data), len(secret.data),
        )
    }

    return true, ""
}

/*
func compatibility(cover, secret *ImageData) bool {
	widthMatch := cover.ihdr.Width == secret.ihdr.Width
	heightMatch := cover.ihdr.Height == secret.ihdr.Height
	bppCover := (int(cover.ihdr.BitDepth) * int(cover.ihdr.ColorType)) / 8
	bppSecret := (int(secret.ihdr.BitDepth) * int(secret.ihdr.ColorType)) / 8
	bppMatch := bppCover == bppSecret
	expectedSize := int(cover.ihdr.Height) * (int(cover.ihdr.Width)*bppCover + 1) // Include filter bytes
	dataSizeMatch := len(cover.data) == expectedSize && len(secret.data) == expectedSize

	return widthMatch && heightMatch && bppMatch && dataSizeMatch
}
	*/



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
			fmt.Printf("Skipping chunk: %s of length %d\n", typ, length)
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
	//bpp := (int(ihdr.BitDepth) * int(ihdr.ColorType)) / 8

	/*
	rawPixelData, err := ApplyFilterMethod(rawPNGData, int(ihdr.Width), int(ihdr.Height), bpp)
	if err != nil {
		return imgData, fmt.Errorf("error processing scanlines: %v", err)
	}
		*/
	rawPixelData := rawPNGData
	imgData.ihdr = ihdr
	imgData.data = rawPixelData
	return imgData, nil
}


func ApplyFilterMethod(data []byte, width int, height int, bytesPerPixel int) ([]byte, error) {
	var pixels []byte
	var bytesPerRow = width * bytesPerPixel
	var scanlineLength = bytesPerRow + 1

	for i := 0; i < height; i++ {
		start := i * scanlineLength
		if start+scanlineLength > len(data) {
			return nil, fmt.Errorf("Row %d: insufficient data for scanline", i)
		}

		filterByte := data[start]
		if filterByte != 0 { // Only method 0 is supported
			continue
			//return nil, fmt.Errorf("Unsupported filter method %d at row %d", filterByte, i)
		}

		// Include filter byte in the pixels data
		rawScanline := data[start : start+scanlineLength]
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


func ReadChunkLength(file *os.File) (uint32, error) {
    var length [4]byte
    _, err := file.Read(length[:])
    if err != nil {
        return 0, fmt.Errorf("err reading chunk length: %v", err)
    }

    ret := binary.BigEndian.Uint32(length[:])
    return ret, nil
}

func ReadChunkCRC(file *os.File) (uint32, error) {
    var crc [4]byte
    _, err := file.Read(crc[:])
    if err != nil {
        return 0, fmt.Errorf("err reading chunk crc: %v", err)
    }

    ret := binary.BigEndian.Uint32(crc[:])
    return ret, nil
}

func ReadIHDR(file *os.File, ihdr *IHDR) error {
    chunkLength, err := ReadChunkLength(file)
    if err != nil {
        return fmt.Errorf("ReadIHDR() => %v", err)
    }

    if chunkLength != 13 {
        return fmt.Errorf("invalid IHDR chunk length: expected 13, got %d", chunkLength)
    }

    typ, err := ReadChunkType(file)
    if err != nil {
        return err
    }

    if typ != "IHDR" {
        return fmt.Errorf("invalid chunk type: expected IHDR, got %s", typ)
    }

    var chunkBuf [13]byte
    _, err = file.Read(chunkBuf[:])
    if err != nil {
        return err
    }

    ihdr.Length = chunkLength
    ihdr.Width = binary.BigEndian.Uint32(chunkBuf[0:4])
    ihdr.Height = binary.BigEndian.Uint32(chunkBuf[4:8])

    ihdr.BitDepth = chunkBuf[8]
    ihdr.ColorType = chunkBuf[9]
    ihdr.CompressionMethod = chunkBuf[10]
    ihdr.FilterMethod = chunkBuf[11]
    ihdr.InterlaceMethod = chunkBuf[12]

    crc, err := ReadChunkCRC(file)
    if err != nil {
        return err
    }

    ihdr.CRC = crc

    return nil
}

func ReadpHYs(file *os.File, p *pHYs) error {
    var chunkBuf [9]byte
    _, err := file.Read(chunkBuf[:])
    if err != nil {
        return err
    }

    p.X = binary.BigEndian.Uint32(chunkBuf[0:4])
    p.Y = binary.BigEndian.Uint32(chunkBuf[4:8])
    p.UnitSpecifier = chunkBuf[8]

    crc, err := ReadChunkCRC(file)
    if err != nil {
        return err
    }

    p.CRC = crc

    return nil
}

func ReadChunkType(file *os.File) (string , error){
	var typ [4]byte
	_, err := file.Read(typ[:])

	if err != nil{
		return "", fmt.Errorf("err reading chunk typ: %v", err)
	}

	return string(typ[:]), nil
}

func samplesPerPixel(colorType byte) int {
    switch colorType {
    case 0:
        return 1 // Grayscale
    case 2:
        return 3 // RGB
    case 3:
        return 1 // Palette index (treated differently)
    case 4:
        return 2 // Grayscale with alpha
    case 6:
        return 4 // RGBA
    default:
        return 0 // Invalid color type
    }
}
