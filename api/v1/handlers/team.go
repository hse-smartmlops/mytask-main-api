package handlers

import (
	"errors"
	"net/http"

	"emplacc-api/api/v1/dto"
	"emplacc-api/api/v1/dto/response"
	"emplacc-api/api/v1/middleware"
	"emplacc-api/api/v1/presenter"
	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/domain"

	"github.com/labstack/echo/v4"
)

type TeamHandler struct {
	service ports.TeamService
}

func NewTeamHandler(service ports.TeamService) *TeamHandler {
	return &TeamHandler{service: service}
}

// @Summary Register Team Routes
// @Description Register routes for team management
// @Tags Teams
func RegisterTeamRoutes(group *echo.Group, service ports.TeamService) {
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

// @Summary List Teams
// @Description Retrieve a paginated list of teams
// @Tags Teams
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.TeamsListResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /teams [get]
func (h *TeamHandler) ListTeams(c echo.Context) error {
	pageNum, size := resolvePagination(c.QueryParam("page"), c.QueryParam("page_size"), 1, 20)
	params := ports.PaginationParams{Page: pageNum, PageSize: size}

	page, err := h.service.ListTeams(c.Request().Context(), params)
	if err != nil {
		return respondError(c, http.StatusInternalServerError, dto.NewError("teams_fetch_failed", "failed to list teams"))
	}

	pagination, payload := presenter.MapTeamsPage(page)
	return respondPaginated(c, http.StatusOK, payload, pagination)
}

// @Summary Get Team
// @Description Retrieve a team by its ID
// @Tags Teams
// @Accept json
// @Produce json
// @Param id path string true "Team ID"
// @Success 200 {object} dto.TeamResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /teams/{id} [get]
func (h *TeamHandler) GetTeam(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid team identifier"))
	}

	team, err := h.service.GetTeam(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return respondError(c, http.StatusNotFound, dto.NewError("team_not_found", "team not found"))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("teams_fetch_failed", "failed to get team"))
	}

	return respondSuccess(c, http.StatusOK, presenter.ToTeamDTO(team))
}

// @Summary List Teams By Project
// @Description Retrieve a list of teams associated with a specific project
// @Tags Teams
// @Accept json
// @Produce json
// @Param project_id path string true "Project ID"
// @Success 200 {object} dto.TeamsListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /projects/{project_id}/teams [get]
func (h *TeamHandler) ListTeamsByProject(c echo.Context) error {
	projectID, err := parseUUID(c.Param("project_id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid project identifier"))
	}

	teams, err := h.service.ListTeamsByProject(c.Request().Context(), projectID)
	if err != nil {
		return respondError(c, http.StatusInternalServerError, dto.NewError("teams_fetch_failed", "failed to list teams"))
	}

	items := make([]response.Team, len(teams))
	for i := range teams {
		items[i] = presenter.ToTeamDTO(&teams[i])
	}

	return respondSuccess(c, http.StatusOK, items)
}

// @Summary Create Team
// @Description Create a new team
// @Tags Teams
// @Accept json
// @Produce json
// @Param team body request.CreateTeam true "Team creation payload"
// @Success 201 {object} dto.TeamResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /teams [post]
func (h *TeamHandler) CreateTeam(c echo.Context) error {
	req, err := middleware.BindAndValidate(c, middleware.ValidateCreateTeamPayload)
	if err != nil {
		return middleware.RespondValidationError(c, err)
	}

	input := ports.CreateTeamInput{
		Name:        req.Name,
		Description: req.Description,
	}

	team, err := h.service.CreateTeam(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("team_create_failed", "failed to create team"))
	}

	return respondSuccess(c, http.StatusCreated, presenter.ToTeamDTO(team))
}

// @Summary Update Team
// @Description Update an existing team
// @Tags Teams
// @Accept json
// @Produce json
// @Param id path string true "Team ID"
// @Param team body request.UpdateTeam true "Team data"
// @Success 200 {object} dto.TeamResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /teams/{id} [patch]
func (h *TeamHandler) UpdateTeam(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid team identifier"))
	}

	req, err := middleware.BindAndValidate(c, middleware.ValidateUpdateTeamPayload)
	if err != nil {
		return middleware.RespondValidationError(c, err)
	}

	input := ports.UpdateTeamInput{
		Name:        req.Name,
		Description: req.Description,
	}

	team, err := h.service.UpdateTeam(c.Request().Context(), id, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		case errors.Is(err, domain.ErrNotFound):
			return respondError(c, http.StatusNotFound, dto.NewError("team_not_found", "team not found"))
		default:
			return respondError(c, http.StatusInternalServerError, dto.NewError("team_update_failed", "failed to update team"))
		}
	}

	return respondSuccess(c, http.StatusOK, presenter.ToTeamDTO(team))
}

// @Summary Delete Team
// @Description Delete a team by its ID
// @Tags Teams
// @Accept json
// @Produce json
// @Param id path string true "Team ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /teams/{id} [delete]
func (h *TeamHandler) DeleteTeam(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid team identifier"))
	}

	if err := h.service.DeleteTeam(c.Request().Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return respondError(c, http.StatusNotFound, dto.NewError("team_not_found", "team not found"))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("team_delete_failed", "failed to delete team"))
	}

	return c.NoContent(http.StatusNoContent)
}

// @Summary Add User to Team
// @Description Add a user to a team
// @Tags Teams
// @Accept json
// @Produce json
// @Param id path string true "Team ID"
// @Param member body request.AddTeamMember true "Team member data"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /teams/{id}/members [post]
func (h *TeamHandler) AddUserToTeam(c echo.Context) error {
	teamID, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid team identifier"))
	}

	req, err := middleware.BindAndValidate(c, middleware.ValidateAddTeamMemberPayload)
	if err != nil {
		return middleware.RespondValidationError(c, err)
	}

	input := ports.TeamMemberInput{
		TeamID:         teamID,
		UserID:         req.UserUUID,
		Specialization: req.Specialization,
	}

	if err := h.service.AddUserToTeam(c.Request().Context(), input); err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("team_member_add_failed", "failed to add member"))
	}

	return c.NoContent(http.StatusNoContent)
}

// @Summary Remove User from Team
// @Description Remove a user from a team
// @Tags Teams
// @Accept json
// @Produce json
// @Param id path string true "Team ID"
// @Param user_id path string true "User ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /teams/{id}/members/{user_id} [delete]
func (h *TeamHandler) RemoveUserFromTeam(c echo.Context) error {
	teamID, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid team identifier"))
	}
	userID, err := parseUUID(c.Param("user_id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid user identifier"))
	}

	if err := h.service.RemoveUserFromTeam(c.Request().Context(), teamID, userID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return respondError(c, http.StatusNotFound, dto.NewError("team_member_not_found", "team member not found"))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("team_member_remove_failed", "failed to remove member"))
	}

	return c.NoContent(http.StatusNoContent)
}

// @Summary Add Team to Project
// @Description Associate a team with a project
// @Tags Teams
// @Accept json
// @Produce json
// @Param id path string true "Team ID"
// @Param project body request.AddTeamProject true "Team-Project association data"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /teams/{id}/projects [post]
func (h *TeamHandler) AddTeamToProject(c echo.Context) error {
	teamID, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid team identifier"))
	}

	req, err := middleware.BindAndValidate(c, middleware.ValidateAddTeamProjectPayload)
	if err != nil {
		return middleware.RespondValidationError(c, err)
	}

	input := ports.TeamProjectInput{
		TeamID:    teamID,
		ProjectID: req.ProjectUUID,
	}

	if err := h.service.AddTeamToProject(c.Request().Context(), input); err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("team_project_add_failed", "failed to attach team to project"))
	}

	return c.NoContent(http.StatusNoContent)
}

// @Summary Add Team to Project
// @Description Associate a team with a project
// @Tags Teams
// @Accept json
// @Produce json
// @Param id path string true "Team ID"
// @Param project body request.AddTeamProject true "Team-Project association data"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /teams/{id}/projects [post]
func (h *TeamHandler) RemoveTeamFromProject(c echo.Context) error {
	teamID, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid team identifier"))
	}
	projectID, err := parseUUID(c.Param("project_id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid project identifier"))
	}

	if err := h.service.RemoveTeamFromProject(c.Request().Context(), teamID, projectID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return respondError(c, http.StatusNotFound, dto.NewError("team_project_not_found", "team-project link not found"))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("team_project_remove_failed", "failed to detach team from project"))
	}

	return c.NoContent(http.StatusNoContent)
}
