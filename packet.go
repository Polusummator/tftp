package tftp

import (
	"encoding/binary"
	"fmt"
)

const (
	opRRQ   = uint16(1)
	opWRQ   = uint16(2)
	opDATA  = uint16(3)
	opACK   = uint16(4)
	opERROR = uint16(5)
)

/*
RRQ/WRQ packet

 2 bytes     string    1 byte     string   1 byte
 ------------------------------------------------
| Opcode |  Filename  |   0  |    Mode    |   0  |
 ------------------------------------------------
*/

type packetRRQ []byte
type packetWRQ []byte

func packRQ(opcode uint16, filename string, mode string) []byte {
	packet := make([]byte, 4+len(filename)+len(mode))
	binary.BigEndian.PutUint16(packet[0:], opcode)
	copy(packet[2:], filename)
	copy(packet[3+len(filename):], mode)
	return packet
}

func unpackRQ(packet []byte) (filename string, mode string, err error) {
	opcode := binary.BigEndian.Uint16(packet[0:])
	if opcode < 1 || opcode > 5 {
		return "", "", fmt.Errorf("invalid opcode: %d", opcode)
	}
	filenameEndPos := 2
	for filenameEndPos < len(packet) && packet[filenameEndPos] != 0 {
		filenameEndPos++
	}
	if filenameEndPos == len(packet) {
		return "", "", fmt.Errorf("invalid RQ filename format")
	}
	filename = string(packet[2:filenameEndPos])
	modeEndPos := 3 + len(filename)
	for modeEndPos < len(packet) && packet[modeEndPos] != 0 {
		modeEndPos++
	}
	if modeEndPos != len(packet)-1 {
		return "", "", fmt.Errorf("invalid RQ mode format")
	}
	mode = string(packet[(3 + len(filename)):modeEndPos])
	return filename, mode, nil
}

func packRRQ(filename string, mode string) packetRRQ {
	return packRQ(opRRQ, filename, mode)
}

func unpackRRQ(packet packetRRQ) (filename string, mode string, err error) {
	return unpackRQ(packet)
}

func packWRQ(packet []byte, filename string, mode string) packetWRQ {
	return packRQ(opWRQ, filename, mode)
}

func unpackWRQ(packet packetWRQ) (filename string, mode string, err error) {
	return unpackRQ(packet)
}

/*
DATA packet

 2 bytes     2 bytes      n bytes
 ----------------------------------
| Opcode |   Block #  |   Data     |
 ----------------------------------
*/

type packetDATA []byte

func packDATA(packet []byte, block uint16, data []byte) int {

}

func unpackDATA(packet packetDATA) (block uint16, data []byte, err error) {

}

/*
ACK packet

 2 bytes     2 bytes
 ---------------------
| Opcode |   Block #  |
 ---------------------
*/

type packetACK []byte

func packACK(packet []byte, block uint16) int {

}

func unpackACK(packet packetACK) (block uint16, err error) {

}

/*
ERROR packet

 2 bytes     2 bytes      string    1 byte
 -----------------------------------------
| Opcode |  ErrorCode |   ErrMsg   |   0  |
 -----------------------------------------
*/

type packetERROR []byte

func packERROR(packet []byte, errorCode uint16, errorMsg uint16) int {

}

func unpackERROR(packet packetERROR) (errorCode uint16, errorMsg uint16, err error) {

}
