package enum

import (
	"database/sql/driver"
	"errors"
)

type UserRole string

const (
	RoleClient UserRole = "client"
	RoleAdmin  UserRole = "admin"
)

func (r *UserRole) String() string {
	return string(*r)
}

func (r *UserRole) Scan(value interface{}) error {
	if value == nil {
		*r = ""
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return errors.New("invalid role type")
	}
	switch UserRole(str) {
	case RoleClient, RoleAdmin:
		*r = UserRole(str)
		return nil
	default:
		return errors.New("unknown role: " + str)
	}
}

func (r *UserRole) Value() (driver.Value, error) {
	return string(*r), nil
}
