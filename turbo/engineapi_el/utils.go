package engineapi_el

import (
	libcommon "github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon-lib/common/hexutility"
)

func makeFCUresponce(status string, lastValidHash libcommon.Hash, errMsg error, payloadID *hexutility.Bytes) *ForkChoiceUpdatedResponse {

	return &ForkChoiceUpdatedResponse{
		PayloadStatus: PayloadStatus{
			Status:          status,
			LatestValidHash: lastValidHash,
			ValidationError: &StringifiedError{errMsg},
		},
		PayloadID: payloadID,
	}
}
