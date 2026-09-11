package model

// UserTokenState is the login/refresh response. Both fields are nullable: an
// unauthenticated refresh reports {"access_token":null,"expires_in":null}.
type UserTokenState struct {
	AccessToken *string `json:"access_token"`
	ExpiresIn   *int64  `json:"expires_in"`
}

// NewUserTokenState builds a populated token state.
func NewUserTokenState(accessToken string, expiresIn int64) UserTokenState {
	token := accessToken
	seconds := expiresIn
	return UserTokenState{AccessToken: &token, ExpiresIn: &seconds}
}

// EmptyUserTokenState builds the all-null token state.
func EmptyUserTokenState() UserTokenState { return UserTokenState{} }

// NewUserTokenStateWithExpiry reports an expiry even when no token could be
// minted, which is what refreshing an unusable token for a known principal does.
func NewUserTokenStateWithExpiry(accessToken string, expiresIn int64) UserTokenState {
	seconds := expiresIn
	state := UserTokenState{ExpiresIn: &seconds}
	if accessToken != "" {
		token := accessToken
		state.AccessToken = &token
	}
	return state
}
