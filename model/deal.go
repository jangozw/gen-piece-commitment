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
	ProposalCid string                  `gorm:"column:proposal_cid;type:varchar(128);not null;default:''"` // 矿工号id, f01234
	ProofType   abi.RegisteredSealProof `gorm:"column:proof_type;type:int;not null;default:0"`             // 扇区id
	Miner       string                  `gorm:"column:miner;type:varchar(32);not null;default:''"`         // c1 md5
	PieceCid    string                  `gorm:"column:piece_cid;type:varchar(128);not null;default:''"`    // c1结果: base64UrlEncodeBytesToString
	Car         string                  `gorm:"column:car;type:varchar(256);not null;default:''"`          // c2证明结果
	Verified    bool                    `gorm:"column:verified;type:int;not null;default:0"`               // c2证明结果
	Model
}
