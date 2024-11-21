package main

import (
	"bytes"
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
	Data					[]byte
	Length				uint32
}

func main() {

	var ihdr 		IHDR
	var idat		IDAT
	var pHYs 		pHYs
	var length 	uint32
	var typ 		string
	//var ChunkType string 
	//open cover image
	cover, err := OpenImage("badgopher.png")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer cover.Close()

	//iterate through data chunks of cover image
	err = ReadIHDR(cover, &ihdr)
	if err != nil {
		log.Fatal(err)
		return
	}

	// getIHDR chunk
	fmt.Printf("%+v\n",ihdr)

	for {
    length, err = ReadChunkLength(cover)
    if err != nil {
        log.Printf("Error reading chunk length: %v", err)
        break
    }

    typ, err = ReadChunkType(cover)
    if err != nil {
        log.Printf("Error reading chunk type: %v", err)
        break
    }

    switch typ {
    case ChunkTypepHYs:
        err = ReadpHYs(cover, &pHYs)
        if err != nil {
            log.Printf("Error processing pHYs chunk: %v", err)
        } else {
            fmt.Printf("Processed pHYs chunk: %+v\n", pHYs)
        }
    case ChunkTypeIDAT:
        err = ReadIDAT(cover, &idat, length)
        if err != nil {
            log.Printf("Error processing IDAT chunk: %v", err)
        } else {
            fmt.Printf("Processed IDAT chunk of length: %d\n", length)
        }
    case ChunkTypeIEND:
        fmt.Printf("Processed IEND chunk: %+v\n", typ)
        break
    default:
        log.Printf("Skipping unknown chunk type: %s of length: %d", typ, length)
        if _, err := cover.Seek(int64(length+4), io.SeekCurrent); err != nil {
            log.Printf("Error seeking past unknown chunk: %v", err)
            break
        }
    }

		if typ == ChunkTypeIEND {
			fmt.Println("Processed entire image. Total IDAT length: %v", idat.Length)
			break
		}
	}

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

    idat.Data = append(idat.Data, newDataBuf...)
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