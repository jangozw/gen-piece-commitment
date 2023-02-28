package util

import (
	"fmt"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/pkg/errors"
)

func GetMinerProofType(miner string) (abi.RegisteredSealProof, error) {
	if miner == "t01000" {
		return abi.RegisteredSealProof_StackedDrg2KiBV1_1, nil
	}

	url := "https://api2.filscout.com/api/v2/miner/" + miner
	type Resp struct {
		Code int `json:"code"`
		Data struct {
			SectorSize int64 `json:"sectorSize"`
		} `json:"data"`
	}
	res, err := HttpPost(url, nil, nil)
	if err != nil {
		return 0, err
	}
	if res.StatusCode == 200 {
		var resp Resp
		if err := res.Decode(&resp); err != nil {
			return 0, err
		}
		if resp.Data.SectorSize > 0 {
			switch resp.Data.SectorSize >> 30 {
			case 32:
				return abi.RegisteredSealProof_StackedDrg32GiBV1_1, nil
			case 64:
				return abi.RegisteredSealProof_StackedDrg64GiBV1_1, nil
			}
		}
	}
	return 0, errors.New(fmt.Sprintf("couldn't get miner %v proof type", miner))
}
