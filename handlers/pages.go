package handlers

import (
	"net/http"
	"pong-htmx/models"
	"pong-htmx/repository"
	"pong-htmx/utils"
	"pong-htmx/views/pages"

	"github.com/labstack/echo/v4"
)

type PagesHandler struct {
	scoreRepo *repository.ScoresRepository
}

func NewPagesHandler(scoreRepo *repository.ScoresRepository) *PagesHandler {
	return &PagesHandler{
		scoreRepo: scoreRepo,
	}
}

func (h *PagesHandler) Index(ctx echo.Context) error {
	return utils.Render(ctx, http.StatusOK, pages.Index(ctx.Request().Context()))
}

func (h *PagesHandler) About(ctx echo.Context) error {
	return utils.Render(ctx, http.StatusOK, pages.About(ctx.Request().Context()))
}

func (h *PagesHandler) Contact(ctx echo.Context) error {
	return utils.Render(ctx, http.StatusOK, pages.Contact(ctx.Request().Context()))
}

func (h *PagesHandler) Login(ctx echo.Context) error {
	return utils.Render(ctx, http.StatusOK, pages.Login(ctx.Request().Context(), false, "", utils.TemplateError{}))
}

func (h *PagesHandler) Ping(ctx echo.Context) error {
	return ctx.JSONBlob(http.StatusOK, []byte(`{"status": "Server Healthy & Running ✅"}`))
}

func (h *PagesHandler) StartGamePage(ctx echo.Context) error {
	return utils.Render(ctx, http.StatusOK, pages.StartGame(ctx.Request().Context()))
}

func (h *PagesHandler) Score(ctx echo.Context) error {
	return utils.Render(ctx, http.StatusOK, pages.Score(ctx.Request().Context()))
}

func (h *PagesHandler) Profile(ctx echo.Context) error {
	user := ctx.Get("user").(models.User)
	return utils.Render(ctx, http.StatusOK, pages.ProfilePage(ctx.Request().Context(), user))
}
