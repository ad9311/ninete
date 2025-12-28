package prog

type ContextKey int

const (
	KeyCurrentUser ContextKey = iota
	KeyExpense
	KeyRecurrentExpense
)
