package model

import "encoding/json"

// Authority is a granted role. Only the role name is exposed to clients, under
// the key "authority"; the row id and the raw name field are suppressed.
type Authority struct {
	ID   int64
	Name RoleName
}

// GetAuthority returns the wire representation of the role.
func (a Authority) GetAuthority() string { return string(a.Name) }

// MarshalJSON emits {"authority":"ROLE_USER"} and nothing else.
func (a Authority) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Authority string `json:"authority"`
	}{Authority: a.GetAuthority()})
}
