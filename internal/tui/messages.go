package tui

// Request messages (from child to parent).

// AddAccountRequestMsg is sent when the user wants to add a new account.
type AddAccountRequestMsg struct{}

// DeleteAccountRequestMsg is sent when the user wants to delete an account.
type DeleteAccountRequestMsg struct {
	Email Email
}

// OAuth flow messages.

// OAuthURLMsg is sent when the OAuth URL is available.
type OAuthURLMsg struct {
	URL string
}

// OAuthCompleteMsg is sent when the OAuth flow completes.
type OAuthCompleteMsg struct {
	Email string
	Err   error
}

// Account management messages.

// AccountDeletedMsg is sent when an account is deleted.
type AccountDeletedMsg struct {
	Email string
	Err   error
}

// Setup messages.

// SetupRequiredMsg is sent when setup is needed.
type SetupRequiredMsg struct{}

// SetupCompleteMsg is sent when setup completes.
type SetupCompleteMsg struct {
	Err error
}

// CredentialsOKMsg is sent when credentials are valid.
type CredentialsOKMsg struct{}

// CalendarManagerStageMsg communicates stage transitions from calendar manager to parent.
type CalendarManagerStageMsg struct {
	Stage string // "loading", "selecting", "saving", "success", "error"
	Err   error  // Set when Stage is "error"
}
