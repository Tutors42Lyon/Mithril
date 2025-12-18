package models

type ExerciseMessage struct {
	Db_id uint `json:"db_id,omitempty" gorm:"primaryKey;column:id;autoIncrement"`

	Title   string `json:"title" gorm:"column:title"`
	Yaml string `json:"yaml_file_path" gorm:"column:yaml_file_path"`
}

func (ExerciseMessage) TableName() string {
	return "exercises"
}
