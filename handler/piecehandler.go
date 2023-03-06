package handler

import (
	"bufio"
	"gen-piece-commitment/inited"
	"gen-piece-commitment/model"
	"gen-piece-commitment/util"
	"github.com/filecoin-project/go-commp-utils/ffiwrapper"
	"github.com/filecoin-project/go-padreader"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/gammazero/workerpool"
	"github.com/ipfs/go-cid"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	StorageCarPathPrefix = "/Volumes/new-apfs/devnet/deals-data/car" // todo modify
)

var minerProofTypes sync.Map

func GetCarReader(path string) (io.Reader, error) {

	if strings.HasPrefix("http", path) {

	} else if strings.HasPrefix("/", path) {

	}
	return nil, nil
}

func ImportDeal(minerID, importFile string) error {
	file, err := os.Open(importFile)
	if err != nil {
		return err
	}

	defer file.Close()
	pool := workerpool.New(100)
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
			log.Errorf("getProofType of miner (%s) failed: %v", minerID, err)
			return nil
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
				log.Errorf("Import deal line %d failed: %v proposal_cid: %v", lineNo, err.Error(), proposalCid)
				return
			}
			succNum.Add(1)
			log.Infof("import deal line %d success. proposal_cid: %v", lineNo, proposalCid)
		})
	}
	pool.StopWait()
	log.Infof("Import deal by %s end. total: %d succ: %d", importFile, totalNum, succNum.Load())
	return nil
}

func StartGenPieceCommitmentTask() error {
	pool := workerpool.New(100)
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
		for _, dl := range deals {
			deal := dl
			// todo remove
			ok, err := inited.Redis.Lock(deal.ProposalCid, 1, time.Second*3600)
			if err != nil {
				log.Errorf("redis lock proposalCid: %v err: %v", deal.ProposalCid, err)
			}
			if !ok {
				log.Warnf("skip redis lock current proposalCid: %v failed. myabe other thread doing", deal.ProposalCid)
				continue
			}
			dbHandler.UpdateDealState(deal.ProposalCid, model.DealCheckStateGeneratingCid)
			carPath := StorageCarPathPrefix + "/" + deal.Car

			pool.Submit(func() {
				var taskSucc bool
				defer func() {
					inited.Redis.Unlock(deal.ProposalCid)
					if !taskSucc {
						dbHandler.UpdateDealState(deal.ProposalCid, model.DealCheckStateGenerateCidFailed)
					}
				}()
				if ok, err := util.IsPathExists(carPath); err != nil {
					log.Errorf("check path %v exists failed! dealID: %v err: %v", carPath, deal.ProposalCid, err.Error())
					return
				} else if !ok {
					log.Errorf("check file not exist: %v dealID: %v", carPath, deal.ProposalCid)
					return
				}
				sf, err := os.Stat(carPath)
				if err != nil {
					log.Errorf("stat file  %v err: %v ", carPath, err)
					return
				}
				carSize := sf.Size()
				fi, err := os.Open(carPath)
				if err != nil {
					log.Errorf("open %v err: %v", carPath, err.Error())
					return
				}
				defer fi.Close()
				log.Infof("generatePieceCommitment start. miner: %s proofType: %d sectorSize: %d car: %s", deal.Miner, deal.ProofType, carSize, carPath)
				pieceCid, perr := generatePieceCommitment(deal.ProofType, fi, uint64(carSize))
				if perr != nil {
					log.Errorf("generatePieceCommitment failed dealID: %v err: %v", deal.ProposalCid, perr.Error())
					return
				}
				log.Infof("generatePieceCommitment success! dealID: %v pieceCid: %v", deal.ProposalCid, pieceCid)
				if err := dbHandler.UpdatePieceCid(deal.ProposalCid, pieceCid.String()); err != nil {
					log.Errorf("updatePieceCid failed dealID: %v err: %v", deal.ProposalCid, err.Error())
					return
				}
				taskSucc = true
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
