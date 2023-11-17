package api

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"io"
	"log/slog"
	"net/http"
	"strconv"
)

// AliasesGet -
func (a *Api) AliasesGet(c *gin.Context) {
	c.Header("Content-Type", "application/json; charset=UTF-8")
	stAliases, err := a.db.Aliases()
	if err != nil {
		c.JSON(500, ErrBackendError)
	}

	aliases := make(AliasesGetResponse, 0, len(stAliases))
	for _, v := range stAliases {
		aliases = append(aliases, aliasFromState(v))
	}
	indented, _ := strconv.ParseBool(c.Query("indented"))
	if indented {
		c.IndentedJSON(http.StatusOK, aliases)
		return
	}

	c.JSON(http.StatusOK, aliases)
}

// AliasesPost -
func (a *Api) AliasesPost(c *gin.Context) {
	c.Header("Content-Type", "application/json; charset=UTF-8")
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		slog.Debug("error reading request body", err)
		c.JSON(400, ErrBadRequest)
		return
	}

	aliasReq := Alias{}
	if err := json.Unmarshal(raw, &aliasReq); err != nil {
		slog.Debug("error parsing request body", err)
		c.JSON(400, ErrBadRequest)
		return
	}

	aliasReq.Id = ""
	if len(aliasReq.ProtectedAddress.Id) < 1 {
		slog.Debug("missing protected_address.id value", c)
		c.JSON(400, ErrBadRequest)
		return
	}

	stAlias := aliasToState(aliasReq)
	newAlias, err := a.db.CreateAlias(stAlias)
	if err != nil {
		slog.Debug("error creating Alias record in state", err)
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			c.JSON(400, ErrDuplicateEntry)
			return
		}

		c.JSON(500, ErrBackendError)
		return
	}
	response := AliasesPostResponse{Alias: aliasFromState(newAlias)}

	c.JSON(http.StatusOK, response)
}

// AliasesDeleteByEmail -
func (a *Api) AliasesDeleteByEmail(c *gin.Context) {
	email := c.Param("email")
	if !validateEmail(email) {
		slog.Debug("error validating email: ", email)
		c.JSON(400, ErrBadRequest)
		return
	}

	alias, ok := a.db.GetAliasByEmail(email)
	if !ok {
		slog.Debug("could not find alias to delete: ", email)
		c.JSON(400, ErrNotFound)
		return
	}

	if err := a.db.DeleteAliasByEmail(email); err != nil {
		slog.Debug("error deleting Alias record in state", err)
		c.JSON(500, ErrBackendError)
		return
	}

	response := AliasEmailDeleteResponse{
		Alias: aliasFromState(alias),
	}

	c.JSON(http.StatusOK, response)
}

// AliasesGetByEmail -
func (a *Api) AliasesGetByEmail(c *gin.Context) {
	email := c.Param("email")
	if !validateEmail(email) {
		slog.Debug("error validating email: ", email)
		c.JSON(400, ErrBadRequest)
		return
	}

	alias, ok := a.db.GetAliasByEmail(email)
	if !ok {
		slog.Debug("could not find alias by email: ", email)
		c.JSON(404, ErrNotFound)
		return
	}

	response := AliasEmailGetResponse{Alias: aliasFromState(alias)}
	c.JSON(http.StatusOK, response)
}

// AliasesDeleteById -
func (a *Api) AliasesDeleteById(c *gin.Context) {
	id := c.Param("id")
	if !validateId(id) {
		slog.Debug("error validating id: ", id)
		c.JSON(400, ErrBadRequest)
		return
	}

	alias, ok := a.db.GetAliasById(id)
	if !ok {
		slog.Debug("could not find alias to delete: ", id)
		c.JSON(400, ErrNotFound)
		return
	}

	if err := a.db.DeleteAliasById(id); err != nil {
		slog.Debug("error deleting Alias record in state", err)
		c.JSON(500, ErrBackendError)
		return
	}

	response := AliasEmailDeleteResponse{
		Alias: aliasFromState(alias),
	}

	c.JSON(http.StatusOK, response)
}

// AliasesGetById -
func (a *Api) AliasesGetById(c *gin.Context) {
	id := c.Param("id")
	if !validateId(id) {
		slog.Debug("error validating id: ", id)
		c.JSON(400, ErrBadRequest)
		return
	}

	alias, ok := a.db.GetAliasById(id)
	if !ok {
		slog.Debug("could not find alias by id: ", id)
		c.JSON(404, ErrNotFound)
		return
	}

	response := AliasEmailGetResponse{Alias: aliasFromState(alias)}
	c.JSON(http.StatusOK, response)
}
