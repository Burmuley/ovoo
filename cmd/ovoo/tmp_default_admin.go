package main

import (
	"context"
	"errors"
	"log/slog"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/services"
)

func makeDefaultAdmin(svcGw *services.ServiceGateway, admin map[string]string) error {
	adminUser := entities.User{
		FirstName: admin["firstName"],
		LastName:  admin["lastName"],
		Login:     admin["login"],
	}
	if _, err := svcGw.Users.Create(context.Background(), adminUser); err != nil {
		if errors.Is(err, entities.ErrDuplicateEntry) {
			slog.Info("default admin user already present in the repository, not creating")
			return nil
		} else {
			return err
		}
	}

	slog.Info("created default admin user")
	return nil
}
