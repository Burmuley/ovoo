package rest

import (
	"fmt"
	"net/http"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/services"
)

// GetUsers retrieves all users and returns them as a response.
func (a *Application) GetUsers(w http.ResponseWriter, r *http.Request) {
	cuser, err := userFromContext(r)
	if err != nil {
		a.errorLogNResponse(w, "getting users: identifying user", err)
		return
	}

	// filling filters
	filters := readFilters(r, []string{"login", "id", "type", "page_size", "page"})
	users, pgm, err := a.svcGw.Users.GetAll(r.Context(), cuser, filters)
	if err != nil {
		a.errorLogNResponse(w, "gettings users", err)
		return
	}

	usersData := make([]UserData, 0, len(users))
	for _, user := range users {
		usersData = append(usersData, userTResponse(user))
	}

	resp := GetUsersResponse{
		PaginationMetadata: pgmTMetadata(pgm),
		Users:              usersData,
	}

	a.successResponse(w, resp, http.StatusOK)
}

func (a *Application) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	cuser, err := userFromContext(r)
	if err != nil {
		a.errorLogNResponse(w, "getting user profile: identifying user", err)
	}

	resp := userTResponse(cuser)
	a.successResponse(w, resp, http.StatusOK)
}

// GetUserById retrieves a user by their ID and returns the user details.
func (a *Application) GetUserById(w http.ResponseWriter, r *http.Request) {
	cuser, err := userFromContext(r)
	if err != nil {
		a.errorLogNResponse(w, "getting user by id: identifying user", err)
	}

	userId := r.PathValue("id")
	user, err := a.svcGw.Users.GetById(r.Context(), cuser, entities.Id(userId))
	if err != nil {
		a.errorLogNResponse(w, "getting user by id", err)
		return
	}

	resp := userTResponse(user)
	a.successResponse(w, resp, http.StatusOK)
}

// CreateUser creates a new user based on the provided request and returns the created user details.
func (a *Application) CreateUser(w http.ResponseWriter, r *http.Request) {
	cuser, err := userFromContext(r)
	if err != nil {
		a.errorLogNResponse(w, "creating user: identifying user", err)
		return
	}

	req := CreateUserRequest{}
	if err := readBody(r.Body, &req); err != nil {
		a.errorLogNResponse(w, "parsing user create request", fmt.Errorf("%w: %w", entities.ErrValidation, err))
		return
	}

	user, err := a.svcGw.Users.Create(r.Context(), cuser, services.UserCreateCmd{
		Login:     req.Login,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Type:      entities.UserType(userTypeFStr(req.Type)),
		Password:  req.Password,
	})

	if err != nil {
		a.errorLogNResponse(w, "creating user", err)
		return
	}

	resp := userTResponse(user)
	a.successResponse(w, resp, http.StatusCreated)
}

// UpdateUser updates an existing user's information and returns the updated user details.
func (a *Application) UpdateUser(w http.ResponseWriter, r *http.Request) {
	cuser, err := userFromContext(r)
	if err != nil {
		a.errorLogNResponse(w, "updating user: identifying user", err)
	}

	userId := r.PathValue("id")
	req := UpdateUserRequest{}
	if err := readBody(r.Body, &req); err != nil {
		a.errorLogNResponse(w, "parsing user update request", fmt.Errorf("%w: %w", entities.ErrValidation, err))
		return
	}

	cmd := services.UserUpdateCmd{
		UserID:    entities.Id(userId),
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}
	if req.Type != nil {
		utyp := userTypeFStr(*req.Type)
		cmd.Type = &utyp
	}
	user, err := a.svcGw.Users.Update(r.Context(), cuser, cmd)
	if err != nil {
		a.errorLogNResponse(w, "updating user by id", fmt.Errorf("updating user: %w", err))
		return
	}

	resp := userTResponse(user)
	a.successResponse(w, resp, http.StatusCreated)
}

// DeleteUser deletes a user by their ID and returns the deleted user's details.
func (a *Application) DeleteUser(w http.ResponseWriter, r *http.Request) {
	cuser, err := userFromContext(r)
	if err != nil {
		a.errorLogNResponse(w, "deleting user: identifying user", err)
	}

	userId := r.PathValue("id")
	user, err := a.svcGw.Users.Delete(r.Context(), cuser, entities.Id(userId))
	if err != nil {
		a.errorLogNResponse(w, "deleting user by id", err)
		return
	}

	resp := userTResponse(user)
	a.successResponse(w, resp, http.StatusCreated)
}
