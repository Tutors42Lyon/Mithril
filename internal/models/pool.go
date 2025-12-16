package models

type PoolMessage struct {
	Db_id uint `json:"db_id,omitempty" gorm:"primaryKey;column:id;autoIncrement"`

	Name   string `json:"name" gorm:"column:name"`
	Category      string `json:"category" gorm:"column:category"`
	Description       string `json:"description" gorm:"column:description;"`
	Is_published   bool   `json:"is_published" gorm:"column:is_published"`
}

func (PoolMessage) TableName() string {
	return "exercise_pools"
}