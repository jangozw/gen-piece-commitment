package handler

import (
	"bufio"
	"gen-piece-commitment/inited"
	"gen-piece-commitment/util"
	"github.com/filecoin-project/go-commp-utils/ffiwrapper"
	"github.com/filecoin-project/go-padreader"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/gammazero/workerpool"
	"github.com/ipfs/go-cid"
	"github.com/pkg/errors"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var minerProofTypes sync.Map

func ImportDeal(minerID, importFile string) error {
	file, err := os.Open(importFile)
	if err != nil {
		return err
	}

	defer file.Close()
	pool := workerpool.New(1000)
	dbHandler := NewDBHandler()
	bf := bufio.NewReader(file)
	totalNum := -1
	var succNum atomic.Int32
	getProofType := func(miner string, pt abi.RegisteredSealProof) (abi.RegisteredSealProof, error) {
		if proofType, ok := minerProofTypes.Load(miner); ok {
			pt = proofType.(abi.RegisteredSealProof)
			if pt > 0 {
				return pt, nil
			}
		}
		proofType, err := util.GetMinerProofType(miner)
		if err != nil {
			if pt > 0 {
				log.Warnf("miner (%s) GetMinerProofType failed: %v, use default proof type: %v, miner:", miner, err.Error(), pt)
				return pt, nil
			}
			log.Errorf("miner (%s) GetMinerProofType failed: %v", miner, err.Error())
			return 0, nil
		}
		minerProofTypes.Store(miner, proofType)
		return proofType, nil
	}
	var proofType abi.RegisteredSealProof
	if minerID != "" {
		pt, err := getProofType(minerID, 0)
		if err != nil || pt == 0 {
			return errors.Errorf("getProofType of miner (%s) failed: %v", minerID, err)
		}
		proofType = pt
	}

	for {
		totalNum++
		line, err := bf.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Errorf("Import deal read line %d err: %s", totalNum, err.Error())
			continue
		}
		line = strings.TrimRight(line, "\n")
		arr := strings.Split(line, " ")
		if len(arr) < 2 {
			log.Errorf("Import deal read line %d format failed err: %s", totalNum, "arr length < 2")
			continue
		}
		proposalCid := arr[0]
		car := arr[1]
		var miner string
		var minerPt abi.RegisteredSealProof
		if minerID == "" {
			if len(arr) < 3 {
				log.Errorf("Import deal read line %d format failed: %s", totalNum, "arr length < 3")
				continue
			}
			miner = arr[2]
			if len(arr) >= 4 {
				pt, _ := strconv.Atoi(arr[3])
				if pt > 0 {
					minerPt = abi.RegisteredSealProof(pt)
				}
			}
			minerPt, err = getProofType(miner, minerPt)
			if err != nil || minerPt == 0 {
				log.Errorf("Import deal read get %s proof type err: %v", miner, err)
				continue
			}
		} else {
			miner = minerID
			minerPt = proofType
		}

		if proposalCid == "" || miner == "" || car == "" || minerPt == 0 {
			log.Errorf("check import line: %d err", totalNum)
			continue
		}

		lineNo := totalNum
		pool.Submit(func() {
			if err := dbHandler.ImportDeal(proposalCid, car, miner, minerPt); err != nil {
				log.Errorf("Import deal failed: %v proposal_cid: %v", err.Error(), proposalCid)
				return
			}
			succNum.Add(1)
			log.Infof("import success! line: %d proposal_cid: %v", lineNo, proposalCid)
		})
	}
	pool.StopWait()
	log.Infof("Imort deal by %s end. total: %d succ: %d", importFile, totalNum, succNum.Load())
	return nil
}

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
				if ok, err := util.IsPathExists(deal.Car); err != nil {
					log.Errorf("check path %v exists failed! dealID: %v err: %v", deal.Car, deal.ProposalCid, err.Error())
					return
				} else if !ok {
					log.Errorf("check file not exist: %v dealID: %v", deal.Car, deal.ProposalCid)
					return
				}
				/*			bytes, err := os.ReadFile(deal.Car)
							if err != nil {
								log.Errorf("get file %v size is err: %v ", deal.Car, err.Error())
								return
							}
							if len(bytes) == 0 {
								log.Errorf("deal car file %v bytes is 0", deal.Car)
								return
							}*/
				fi, err := os.Open(deal.Car)
				if err != nil {
					log.Errorf("open %v err: %v", deal.Car, err.Error())
					return
				}
				defer fi.Close()
				bf := bufio.NewReader(fi)
				carSize := bf.Size()

				log.Infof("generatePieceCommitment start. miner: %s proofType: %d sectorSize: %d car: %s", deal.Miner, deal.ProofType, carSize, deal.Car)

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
