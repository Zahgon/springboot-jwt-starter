package model

// RoleName enumerates the authorities a user may hold.
type RoleName string

// The two roles recognised by the application.
const (
	RoleUser  RoleName = "ROLE_USER"
	RoleAdmin RoleName = "ROLE_ADMIN"
)
