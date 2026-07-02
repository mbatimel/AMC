package models

type CreateUserResponse struct {
	User User `json:"user"`
}

type GetUserResponse struct {
	User User `json:"user"`
}

type ListUsersResponse struct {
	Items      []UserListItem `json:"items"`
	Pagination Pagination     `json:"pagination"`
}

type UpdateUserResponse struct {
	User User `json:"user"`
}

type DeleteUserResponse struct {
	Deleted bool `json:"deleted"`
}

type ActivateUserResponse struct {
	User User `json:"user"`
}

type DeactivateUserResponse struct {
	User User `json:"user"`
}
