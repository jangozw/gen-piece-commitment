package handler

import (
	"gen-piece-commitment/inited"
	"gen-piece-commitment/model"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

type Deal struct {
	ProofType   abi.RegisteredSealProof
	Car         string
	ProposalCid string
	Miner       string
}
type StateHandler interface {
	GetUncheckDeals(int) ([]Deal, error)
	UpdatePieceCid(string, string) error
}
type DBHandler struct {
	Db *gorm.DB
}

func NewDBHandler() *DBHandler {
	return &DBHandler{
		inited.DB,
	}
}
func (s *DBHandler) GetUncheckDeals(num int) ([]Deal, error) {
	var deals []model.TDealInfo
	query := s.Db.Model(&model.TDealInfo{}).Where("check_state = ?", model.DealCheckStateWait).Order("id asc").Limit(num).Find(&deals)
	if err := query.Error; err != nil {
		return nil, err
	}
	var ret []Deal
	for _, d := range deals {
		ret = append(ret, Deal{
			ProposalCid: d.ProposalCid,
			ProofType:   d.ProofType,
			Car:         d.Car,
			Miner:       d.Miner,
		})
	}
	return ret, nil
}

func (s *DBHandler) UpdatePieceCid(proposalCID, pieceCid string) error {
	return s.Db.Model(&model.TDealInfo{}).Where("proposal_cid =?", proposalCID).Update(map[string]interface{}{
		"piece_cid":   pieceCid,
		"check_state": model.DealCheckStateGeneratedCid,
	}).Error
}

func (s *DBHandler) UpdateDealState(proposalCid string, state model.DealCheckState) error {
	update := map[string]interface{}{
		"check_state": state,
	}
	return s.Db.Model(&model.TDealInfo{}).Where("proposal_cid=?", proposalCid).Update(update).Error
}

func (s *DBHandler) ImportDeal(proposalCid string, car string, miner string, proofType abi.RegisteredSealProof) error {

	if proposalCid == "" || car == "" {
		return errors.Errorf("check data")
	}

	var data = model.TDealInfo{}
	if err := s.Db.Model(&model.TDealInfo{}).Where("proposal_cid=?", proposalCid).First(&data).Error; err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	if data.ID > 0 && data.PieceCid != "" {
		return nil
		//return errors.("Skip this deal areadly generated pieceCid")
	}

	if data.ID > 0 {
		update := map[string]interface{}{
			"proposal_cid": proposalCid,
			"car":          car,
			"miner":        miner,
			"proof_type":   proofType,
		}
		return s.Db.Model(&model.TDealInfo{}).Where("id=?", data.ID).Update(update).Error
	} else {
		return s.Db.Model(&model.TDealInfo{}).Create(&model.TDealInfo{
			ProposalCid: proposalCid,
			Car:         car,
			Miner:       miner,
			ProofType:   proofType,
		}).Error
	}
}
