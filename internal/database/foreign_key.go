package database

type ForeignKey struct {
	Table      string `json:"table"`
	Column     string `json:"column"`
	RefTable   string `json:"refTable"`
	RefColumn  string `json:"refColumn"`
	Constraint string `json:"constraint"`
}
