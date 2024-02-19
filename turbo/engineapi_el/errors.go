package engineapi_el

import libcommon "github.com/ledgerwatch/erigon-lib/common"

const (
	VALID              = "VALID"
	INVALID            = "INVALID"
	SYNCING            = "SYNCING"
	ACCEPTED           = "ACCEPTED"
	INVALID_BLOCK_HASH = "INVALID_BLOCK_HASH"
)

// Engine API error codes (sent to CL)
const (
	SERVER_ERROR         = -32000
	INVALID_REQUEST      = -32600
	METHOD_NOT_FOUND     = -32601
	INVALID_PARAMS_ERROR = -32602
	INTERNAL_ERROR       = -32603
	INVALID_JSON         = -32700

	UNKNOWN_PAYLOAD            = -38001
	INVALID_FORKCHOICE_STATE   = -38002
	INVALID_PAYLOAD_ATTRIBUTES = -38003
	TOO_LARGE_REQUEST          = -38004
	UNSUPPORTED_FORK_ERROR     = -38005
)

type engineAPIError struct {
	code int
	msg  string
}

func (e *engineAPIError) ErrorCode() int { return e.code }

func (e *engineAPIError) Error() string { return e.msg }

func makeError(code int, msg string) *engineAPIError {
	return &engineAPIError{code, msg}
}

var (
	STATUS_SYNCING = ForkChoiceUpdatedResponse{
		PayloadStatus: PayloadStatus{
			Status: SYNCING,
		},
		PayloadID: nil,
	}

	INVALID_PAYLOAD_ATTRIBUTES_ERR = engineAPIError{
		code: INVALID_PAYLOAD_ATTRIBUTES,
		msg:  "Invalid payload attributes",
	}
)

func payloadResponse(status string, validationError error, latestValidHash libcommon.Hash, crit error) *PayloadStatus {
	err := validationError.Error()
	return &PayloadStatus{
		Status:          status,
		ValidationError: &err,
		LatestValidHash: latestValidHash,
		CriticalError:   crit,
	}
}
