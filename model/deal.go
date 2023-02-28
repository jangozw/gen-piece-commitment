package model

import (
	"github.com/filecoin-project/go-state-types/abi"
	"time"
)

type Model struct {
	CreatedAt time.Time  `gorm:"column:created_at;type:datetime;default:null"`                           // 创建时间
	UpdatedAt time.Time  `gorm:"column:updated_at;type:datetime;default:null;default:CURRENT_TIMESTAMP"` // 更新时间
	DeletedAt *time.Time `gorm:"column:deleted_at;type:datetime;default:null"`
}

type TDealInfo struct {
	ID          uint64                  `gorm:"column:id;AUTO_INCREMENT;not null"`
	ProposalCid string                  `gorm:"column:proposal_cid;unique_index:pcid;type:varchar(128);not null;default:''"` // 交易id
	ProofType   abi.RegisteredSealProof `gorm:"column:proof_type;type:int;not null;default:0"`                               // 密封类型
	Miner       string                  `gorm:"column:miner;type:varchar(32);not null;default:''"`                           // miner id
	PieceCid    string                  `gorm:"column:piece_cid;type:varchar(128);not null;default:''"`                      // 根据car路径计算出piece cid
	Car         string                  `gorm:"column:car;type:varchar(256);not null;default:''"`                            // car路径
	Verified    bool                    `gorm:"column:verified;type:int;not null;default:0"`                                 // 是否是已验证交易
	Model
}
