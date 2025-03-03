package rest

import (
	"fmt"
	"net/http"

	"github.com/Burmuley/ovoo/internal/entities"
)

// GetUsers retrieves all users and returns them as a response.
func (c *Controller) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := c.svcGw.Users.GetAll(c.context)
	if err != nil {
		c.errorLogNResponse(w, "gettings users", err)
		return
	}

	resp := make(GetUsersResponse, 0, len(users))
	for _, user := range users {
		resp = append(resp, userTResponse(user))
	}

	c.successResponse(w, resp, http.StatusOK)
}

// GetUserById retrieves a user by their ID and returns the user details.
func (c *Controller) GetUserById(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("id")
	user, err := c.svcGw.Users.GetById(c.context, entities.Id(userId))
	if err != nil {
		c.errorLogNResponse(w, "getting user by id", err)
		return
	}

	resp := userTResponse(user)
	c.successResponse(w, resp, http.StatusOK)
}

// CreateUser creates a new user based on the provided request and returns the created user details.
func (c *Controller) CreateUser(w http.ResponseWriter, r *http.Request) {
	req_body := CreateUserRequest{}
	if err := readBody(r.Body, &req_body); err != nil {
		c.errorLogNResponse(w, "parsing user create request", fmt.Errorf("%w: %w", entities.ErrValidation, err))
		return
	}

	user, err := c.svcGw.Users.Create(c.context, entities.User{
		Login:     req_body.Login,
		FirstName: req_body.FirstName,
		LastName:  req_body.LastName,
		Type:      entities.UserType(userTypeFStr(req_body.Type)),
	})
	if err != nil {
		c.errorLogNResponse(w, "creating user", err)
		return
	}

	resp := userTResponse(user)
	c.successResponse(w, resp, http.StatusCreated)
}

// UpdateUser updates an existing user's information and returns the updated user details.
func (c *Controller) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("id")
	user, err := c.svcGw.Users.GetById(c.context, entities.Id(userId))
	if err != nil {
		c.errorLogNResponse(w, "updating user by id", err)
		return
	}

	req_body := UpdateUserRequest{}
	if err := readBody(r.Body, &req_body); err != nil {
		c.errorLogNResponse(w, "parsing user update request", fmt.Errorf("%w: %w", entities.ErrValidation, err))
		return
	}

	user.FirstName = *req_body.FirstName
	user.LastName = *req_body.LastName
	user.Type = userTypeFStr(*req_body.Type)
	user, err = c.svcGw.Users.Update(c.context, user)
	if err != nil {
		c.errorLogNResponse(w, "updating user by id", fmt.Errorf("updating user: %w", err))
		return
	}

	resp := userTResponse(user)
	c.successResponse(w, resp, http.StatusCreated)
}

// DeleteUser deletes a user by their ID and returns the deleted user's details.
func (c *Controller) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("id")
	user, err := c.svcGw.Users.GetById(c.context, entities.Id(userId))
	if err != nil {
		c.errorLogNResponse(w, "check if user to delete by id exists", err)
		return
	}

	if err := c.svcGw.Users.Delete(c.context, entities.Id(userId)); err != nil {
		c.errorLogNResponse(w, "deleting user by id", err)
		return
	}

	resp := userTResponse(user)
	c.successResponse(w, resp, http.StatusCreated)
}
