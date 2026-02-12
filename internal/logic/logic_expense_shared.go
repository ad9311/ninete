package logic

type expenseBaseParams struct {
	CategoryID  int    `validate:"required,gt=0"`
	Description string `validate:"required,min=3,max=50"`
	Amount      uint64 `validate:"required,gt=0"`
}
