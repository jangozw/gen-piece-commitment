package handler

import (
	"gen-piece-commitment/inited"
	"gen-piece-commitment/model"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/jinzhu/gorm"
)

type Deal struct {
	ProofType   abi.RegisteredSealProof
	CarPath     string
	ProposalCid string
	PieceCid    string
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
	query := s.Db.Model(&model.TDealInfo{}).Where("piece_cid = ''").Order("id asc").Limit(num).Find(&deals)
	if err := query.Error; err != nil {
		return nil, err
	}
	var ret []Deal
	for _, d := range deals {
		ret = append(ret, Deal{
			ProposalCid: d.ProposalCid,
			ProofType:   d.ProofType,
			PieceCid:    d.PieceCid,
			CarPath:     d.Car,
		})
	}
	return ret, nil
}

func (s *DBHandler) UpdatePieceCid(proposalCID, pieceCid string) error {
	return s.Db.Model(&model.TDealInfo{}).Where("proposal_cid =?", proposalCID).Update(map[string]interface{}{
		"piece_cid": pieceCid,
	}).Error
}
