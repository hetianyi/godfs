package bridgev2

import "io"

type OperationHandler struct {
	OperationCode byte
	Handler func(metaJson string, bodyReader *io.Reader)
}
