package handlers

import (
	"errors"
	"net/http"
	"strings"

	"emplacc-api/api/v1/dto"
	"emplacc-api/api/v1/dto/request"
	"emplacc-api/api/v1/presenter"
	"emplacc-api/internal/app"
	"emplacc-api/internal/domain"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type TeamHandler struct {
	service app.TeamService
}

func NewTeamHandler(service app.TeamService) *TeamHandler {
	return &TeamHandler{service: service}
}

func RegisterTeamRoutes(group *echo.Group, service app.TeamService) {
	handler := NewTeamHandler(service)

	group.GET("/teams", handler.ListTeams)
	group.GET("/teams/:id", handler.GetTeam)
	group.GET("/projects/:project_id/teams", handler.ListTeamsByProject)
	group.POST("/teams", handler.CreateTeam)
	group.PATCH("/teams/:id", handler.UpdateTeam)
	group.DELETE("/teams/:id", handler.DeleteTeam)
	group.POST("/teams/:id/members", handler.AddUserToTeam)
	group.DELETE("/teams/:id/members/:user_id", handler.RemoveUserFromTeam)
	group.POST("/teams/:id/projects", handler.AddTeamToProject)
	group.DELETE("/teams/:id/projects/:project_id", handler.RemoveTeamFromProject)
}

func (h *TeamHandler) ListTeams(c echo.Context) error {
	params := app.PaginationParams{
		Page:     parsePositiveInt(c.QueryParam("page"), 1),
		PageSize: parsePositiveInt(c.QueryParam("page_size"), 20),
	}

	page, err := h.service.ListTeams(c.Request().Context(), params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewError("teams_fetch_failed", "failed to list teams"))
	}

	pagination, payload := presenter.MapTeamsPage(page)
	return c.JSON(http.StatusOK, dto.NewPaginatedResponse(payload, pagination))
}

func (h *TeamHandler) GetTeam(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid team identifier"))
	}

	team, err := h.service.GetTeam(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.JSON(http.StatusNotFound, dto.NewError("team_not_found", "team not found"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("teams_fetch_failed", "failed to get team"))
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse(presenter.ToTeamDTO(team)))
}

func (h *TeamHandler) ListTeamsByProject(c echo.Context) error {
	projectID, err := parseUUID(c.Param("project_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid project identifier"))
	}

	teams, err := h.service.ListTeamsByProject(c.Request().Context(), projectID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewError("teams_fetch_failed", "failed to list teams"))
	}

	items := make([]response.Team, len(teams))
	for i := range teams {
		items[i] = presenter.ToTeamDTO(&teams[i])
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse(items))
}

func (h *TeamHandler) CreateTeam(c echo.Context) error {
	var req request.CreateTeam
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "failed to parse request body"))
	}

	input := app.CreateTeamInput{
		Name:        req.Name,
		Description: sanitizeStringPtr(req.Description),
	}

	team, err := h.service.CreateTeam(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("team_create_failed", "failed to create team"))
	}

	return c.JSON(http.StatusCreated, dto.NewSuccessResponse(presenter.ToTeamDTO(team)))
}

func (h *TeamHandler) UpdateTeam(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid team identifier"))
	}

	var req request.UpdateTeam
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "failed to parse request body"))
	}

	input := app.UpdateTeamInput{
		Name:        sanitizeStringPtr(req.Name),
		Description: sanitizeStringPtr(req.Description),
	}

	team, err := h.service.UpdateTeam(c.Request().Context(), id, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		case errors.Is(err, domain.ErrNotFound):
			return c.JSON(http.StatusNotFound, dto.NewError("team_not_found", "team not found"))
		default:
			return c.JSON(http.StatusInternalServerError, dto.NewError("team_update_failed", "failed to update team"))
		}
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse(presenter.ToTeamDTO(team)))
}

func (h *TeamHandler) DeleteTeam(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid team identifier"))
	}

	if err := h.service.DeleteTeam(c.Request().Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.JSON(http.StatusNotFound, dto.NewError("team_not_found", "team not found"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("team_delete_failed", "failed to delete team"))
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *TeamHandler) AddUserToTeam(c echo.Context) error {
	teamID, err := parseUUID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid team identifier"))
	}

	var req request.AddTeamMember
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "failed to parse request body"))
	}

	userID, err := parseUUID(req.UserID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "user_id must be a valid UUID"))
	}

	input := app.TeamMemberInput{
		TeamID:         teamID,
		UserID:         userID,
		Specialization: sanitizeStringPtr(req.Specialization),
	}

	if err := h.service.AddUserToTeam(c.Request().Context(), input); err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("team_member_add_failed", "failed to add member"))
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *TeamHandler) RemoveUserFromTeam(c echo.Context) error {
	teamID, err := parseUUID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid team identifier"))
	}
	userID, err := parseUUID(c.Param("user_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid user identifier"))
	}

	if err := h.service.RemoveUserFromTeam(c.Request().Context(), teamID, userID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.JSON(http.StatusNotFound, dto.NewError("team_member_not_found", "team member not found"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("team_member_remove_failed", "failed to remove member"))
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *TeamHandler) AddTeamToProject(c echo.Context) error {
	teamID, err := parseUUID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid team identifier"))
	}

	var req request.AddTeamProject
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "failed to parse request body"))
	}

	projectID, err := parseUUID(req.ProjectID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", "project_id must be a valid UUID"))
	}

	input := app.TeamProjectInput{
		TeamID:    teamID,
		ProjectID: projectID,
	}

	if err := h.service.AddTeamToProject(c.Request().Context(), input); err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return c.JSON(http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("team_project_add_failed", "failed to attach team to project"))
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *TeamHandler) RemoveTeamFromProject(c echo.Context) error {
	teamID, err := parseUUID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid team identifier"))
	}
	projectID, err := parseUUID(c.Param("project_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("invalid_id", "invalid project identifier"))
	}

	if err := h.service.RemoveTeamFromProject(c.Request().Context(), teamID, projectID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.JSON(http.StatusNotFound, dto.NewError("team_project_not_found", "team-project link not found"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("team_project_remove_failed", "failed to detach team from project"))
	}

	return c.NoContent(http.StatusNoContent)
}

func sanitizeStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	res := trimmed
	return &res
}
