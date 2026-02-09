package webkeys

type ContextKey string

const (
	TemplateData = ContextKey("templateData")
)

const (
	SessionIsUserSignedIn = "isUserSignedIn"
	SessionUserID         = "userID"
)
