package board_column

import (
	"time"

	"github.com/google/uuid"
)

type BoardColumn struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	BoardID   uuid.UUID `gorm:"type:uuid;not null"`
	Name      string    `gorm:"type:varchar(255);not null"`
	Position  int       `gorm:"type:integer;not null;default:0"`
	IsBacklog bool      `gorm:"type:boolean;not null;default:false"`
	IsHidden  bool      `gorm:"type:boolean;not null;default:false"`
	IsDone    bool      `gorm:"type:boolean;not null;default:false"`
	Color     string    `gorm:"type:varchar(7);default:'#6B7280'"`
	WipLimit  *int      `gorm:"type:integer"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (BoardColumn) TableName() string {
	return "board_columns"
}
