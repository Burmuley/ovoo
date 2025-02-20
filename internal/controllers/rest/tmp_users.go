package rest

import (
	"github.com/Burmuley/ovoo/internal/entities"
)

// TMP FUNCTION
// TODO: implementa auth and delete it
func (c *Controller) getFirstUser() entities.User {
	users, _ := c.svcGw.Users.GetAll(c.context)
	return users[0]
}
