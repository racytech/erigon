package engineapi_el

import (
	"encoding/binary"

	libcommon "github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon-lib/common/hexutility"
)

func makeFCUresponce(status string, lastValidHash libcommon.Hash, errMsg *string, payloadID *hexutility.Bytes) *ForkChoiceUpdatedResponse {

	return &ForkChoiceUpdatedResponse{
		PayloadStatus: PayloadStatus{
			Status:          status,
			LatestValidHash: lastValidHash,
			ValidationError: errMsg,
		},
		PayloadID: payloadID,
	}
}

func convertPayloadId(payloadId uint64) *hexutility.Bytes {
	encodedPayloadId := make([]byte, 8)
	binary.BigEndian.PutUint64(encodedPayloadId, payloadId)
	ret := hexutility.Bytes(encodedPayloadId)
	return &ret
}

func payloadIDtoUint64(payloadID hexutility.Bytes) uint64 {
	return binary.BigEndian.Uint64(payloadID)
}
