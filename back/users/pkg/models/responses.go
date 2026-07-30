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

type GetProfileResponse struct {
	Profile Profile `json:"profile"`
}

type UpdateProfileResponse struct {
	Profile Profile `json:"profile"`
}

type ListUserClientsResponse struct {
	Items []UserClientListItem `json:"items"`
}

type GetClientDetailsResponse struct {
	Details ClientDetails `json:"details"`
}

type GetClientConditionsResponse struct {
	Conditions ClientConditions `json:"conditions"`
}

type SwitchActiveClientResponse struct {
	ActiveClient UserClientListItem `json:"active_client"`
}

type ListFavoritesResponse struct {
	Items []Favorite `json:"items"`
}

type AddFavoriteResponse struct {
	Favorite Favorite `json:"favorite"`
}

type DeleteFavoritesResponse struct {
	Deleted int `json:"deleted"`
}
