package models

type User42 struct {
	ID          int    `json:"id"`
	Username    string `json:"login"`
	Email       string `json:"email"`
	School_year string `json:"pool_year"`
	Is_active   bool   `json:"active?"`
}

type UserMessage struct {

	Db_id      uint   `json:"db_id,omitempty" gorm:"primaryKey;column:id;autoIncrement"`

	Username   string `json:"username" gorm:"column:username"`
	Email      string `json:"email" gorm:"column:email"`
	Role       string `json:"role" gorm:"column:role"`
	IntraID    int    `json:"intra_id" gorm:"column:intra_id"`
	SchoolYear string `json:"school_year" gorm:"column:school_year"`
	IsActive   bool   `json:"is_active" gorm:"column:is_active"`
}
