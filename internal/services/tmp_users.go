package services

import (
	"context"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories/factory"
)

// TMP FUNCTION
// TODO: implementa auth and delete it
func getFirstUser(ctx context.Context, repoFactory factory.RepoFactory) entities.User {
	users, _ := repoFactory.Users.GetAll(ctx)
	return users[0]
}
