package tftp

const (
	ErrUndefined         = 0
	ErrFileNotFound      = 1
	ErrAccessViolation   = 2
	ErrDiskFull          = 3
	ErrIllegalOperation  = 4
	ErrUnknownTransferID = 5
	ErrFileExists        = 6
	ErrNoSuchUser        = 7
)

var errorMessages = map[uint16]string{
	ErrUndefined:         "Not defined",
	ErrFileNotFound:      "File not found",
	ErrAccessViolation:   "Access violation",
	ErrDiskFull:          "Disk full or allocation exceeded",
	ErrIllegalOperation:  "Illegal TFTP operation",
	ErrUnknownTransferID: "Unknown transfer ID",
	ErrFileExists:        "File already exists",
	ErrNoSuchUser:        "No such user",
}

func getErrorMessage(code uint16) string {
	if msg, ok := errorMessages[code]; ok {
		return msg
	}
	return errorMessages[ErrUndefined]
}
