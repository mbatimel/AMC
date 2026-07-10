package postgres

import (
	_ "embed"
)

//go:embed sql/createUser.sql
var sqlCreateUser string

//go:embed sql/getUserByEmail.sql
var sqlGetUserByEmail string

//go:embed sql/getUserByID.sql
var sqlGetUserByID string

//go:embed sql/updateUserPassword.sql
var sqlUpdateUserPassword string

//go:embed sql/updateUserStatus.sql
var sqlUpdateUserStatus string
