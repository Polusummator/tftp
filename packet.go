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

func packRQ(packet []byte, opcode uint16, filename string, mode string) (n int) {
	n = 4 + len(filename) + len(mode)
	binary.BigEndian.PutUint16(packet, opcode)
	copy(packet[2:], filename)
	packet[2+len(filename)] = 0
	copy(packet[3+len(filename):], mode)
	packet[3+len(filename)+len(mode)] = 0
	return n
}

func unpackRQ(packet []byte) (filename string, mode string, err error) {
	opcode := binary.BigEndian.Uint16(packet)
	if opcode != opRRQ && opcode != opWRQ {
		return "", "", fmt.Errorf("invalid RQ opcode: %d", opcode)
	}
	filenameEndPos := 2
	for filenameEndPos < len(packet) && packet[filenameEndPos] != 0 {
		filenameEndPos++
	}
	if filenameEndPos >= len(packet)-2 {
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

func packRRQ(packet []byte, filename string, mode string) int {
	return packRQ(packet, opRRQ, filename, mode)
}

func unpackRRQ(packet packetRRQ) (filename string, mode string, err error) {
	return unpackRQ(packet)
}

func packWRQ(packet []byte, filename string, mode string) int {
	return packRQ(packet, opWRQ, filename, mode)
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

func packDATA(block uint16, data []byte) packetDATA {
	packet := make([]byte, 4+len(data))
	binary.BigEndian.PutUint16(packet, opDATA)
	binary.BigEndian.PutUint16(packet[2:], block)
	copy(packet[4:], data)
	return packet
}

func unpackDATA(packet packetDATA) (block uint16, data []byte, err error) {
	opcode := binary.BigEndian.Uint16(packet)
	if opcode != opDATA {
		return 0, nil, fmt.Errorf("invalid DATA opcode: %d", opcode)
	}
	block = binary.BigEndian.Uint16(packet[2:])
	return block, packet[4:], nil
}

/*
ACK packet

 2 bytes     2 bytes
 ---------------------
| Opcode |   Block #  |
 ---------------------
*/

type packetACK []byte

func packACK(block uint16) packetACK {
	packet := make([]byte, 4)
	binary.BigEndian.PutUint16(packet, opACK)
	binary.BigEndian.PutUint16(packet[2:], block)
	return packet
}

func unpackACK(packet packetACK) (block uint16, err error) {
	opcode := binary.BigEndian.Uint16(packet)
	if opcode != opACK {
		return 0, fmt.Errorf("invalid ACK opcode: %d", opcode)
	}
	block = binary.BigEndian.Uint16(packet[2:])
	return block, nil
}

/*
ERROR packet

 2 bytes     2 bytes      string    1 byte
 -----------------------------------------
| Opcode |  ErrorCode |   ErrMsg   |   0  |
 -----------------------------------------
*/

type packetERROR []byte

func packERROR(errorCode uint16, errMsg string) packetERROR {
	packet := make([]byte, 5+len(errMsg))
	binary.BigEndian.PutUint16(packet, opERROR)
	binary.BigEndian.PutUint16(packet[2:], errorCode)
	copy(packet[4:], errMsg)
	return packet
}

func unpackERROR(packet packetERROR) (errorCode uint16, errMsg string, err error) {
	opcode := binary.BigEndian.Uint16(packet)
	if opcode != opERROR {
		return 0, "", fmt.Errorf("invalid ERROR opcode: %d", opcode)
	}
	errorCode = binary.BigEndian.Uint16(packet[2:])
	errMsgEndPos := 4
	for errMsgEndPos < len(packet) && packet[errMsgEndPos] != 0 {
		errMsgEndPos++
	}
	if errMsgEndPos != len(packet)-1 {
		return 0, "", fmt.Errorf("invalid ERROR errMsg format")
	}
	errMsg = string(packet[4:errMsgEndPos])
	return errorCode, errMsg, nil
}
