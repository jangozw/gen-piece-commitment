package handler

import (
	"gen-piece-commitment/inited"
	"gen-piece-commitment/util"
	"github.com/filecoin-project/go-commp-utils/ffiwrapper"
	"github.com/filecoin-project/go-padreader"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/gammazero/workerpool"
	"github.com/ipfs/go-cid"
	"io"
	"os"
	"time"
)

func StartGenPieceCommitmentTask() error {
	pool := workerpool.New(10)
	dbHandler := NewDBHandler()
	for {
		deals, err := dbHandler.GetUncheckDeals(10)
		if err != nil {
			log.Errorf("getUncheckDeals err: %v", err.Error())
			continue
		}
		if len(deals) == 0 {
			log.Debugf("no unchecked deals found, watiting 1 s")
			time.Sleep(1 * time.Second)
			continue
		}
		for _, deal := range deals {
			ok, err := inited.Redis.Lock(deal.ProposalCid, 1, time.Second*3600)
			if err != nil {
				log.Errorf("redis lock proposalCid: %v err: %v", deal.ProposalCid, err)
			}
			if !ok {
				log.Warnf("skip redis lock current proposalCid: %v failed. myabe other thread doing", deal.ProposalCid)
				continue
			}
			pool.Submit(func() {
				if ok, err := util.IsPathExists(deal.CarPath); err != nil {
					log.Errorf("check path %v exists failed! dealID: %v err: %v", deal.CarPath, deal.ProposalCid, err.Error())
					return
				} else if !ok {
					log.Errorf("check file not exist: %v dealID: %v", deal.CarPath, deal.ProposalCid)
					return
				}
				bytes, err := os.ReadFile(deal.CarPath)
				if err != nil {
					log.Errorf("get file %v size is err: %v ", deal.CarPath, err.Error())
					return
				}
				if len(bytes) == 0 {
					log.Errorf("deal car file %v bytes is 0", deal.CarPath)
					return
				}
				fi, err := os.Open(deal.CarPath)
				if err != nil {
					log.Errorf("open %v err: %v", deal.CarPath, err.Error())
					return
				}
				defer fi.Close()
				carSize := len(bytes)
				pieceCid, perr := generatePieceCommitment(deal.ProofType, fi, uint64(carSize))
				defer func() {
					if ok, err := inited.Redis.Unlock(deal.ProposalCid); err != nil || !ok {
						log.Errorf("redis unlock deal %v failed. %v %v", deal.ProposalCid, ok, err)
					}
				}()
				if perr != nil {
					log.Errorf("generatePieceCommitment failed dealID: %v err: %v", deal.ProposalCid, perr.Error())
					return
				}
				log.Infof("generatePieceCommitment success! dealID: %v pieceCid: %v", deal.ProposalCid, pieceCid)
				if err := dbHandler.UpdatePieceCid(deal.ProposalCid, pieceCid.String()); err != nil {
					log.Errorf("updatePieceCid failed dealID: %v err: %v", deal.ProposalCid, err.Error())
					return
				}
			})
		}
	}
}

func generatePieceCommitment(rt abi.RegisteredSealProof, rd io.Reader, pieceSize uint64) (cid.Cid, error) {
	paddedReader, paddedSize := padreader.New(rd, pieceSize)
	commitment, err := ffiwrapper.GeneratePieceCIDFromFile(rt, paddedReader, paddedSize)
	if err != nil {
		return cid.Undef, err
	}
	return commitment, nil
}
