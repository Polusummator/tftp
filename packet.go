package tftp

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

func packRQ(packet []byte, opcode uint16, filename string, mode string) int {}

func unpackRQ(packet []byte) (opcode uint16, filename string, mode string, err error) {}

func packRRQ(filename string, mode string) int {

}

func unpackRRQ(packet packetRRQ) (filename string, mode string, err error) {

}

func packWRQ(packet []byte, filename string, mode string) int {

}

func unpackWRQ(packet packetWRQ) (filename string, mode string, err error) {

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
