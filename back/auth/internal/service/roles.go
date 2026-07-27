package service

// Role codes as seeded in the roles table (back/migrations/pkg/migrations/data/20260705171948_access_roles.sql).
const (
	RoleCodeAdmin    = 0
	RoleCodeSupport  = 1
	RoleCodeSupplier = 2
	RoleCodeBuyer    = 3
)

// defaultSignUpRoleCode is the role assigned to every user created via RegisterIP/RegisterIndividual.
const defaultSignUpRoleCode = RoleCodeBuyer
