package handlers

import (
	"errors"
	"net/http"
	"strings"

	"emplacc-api/api/v1/dto"
	"emplacc-api/api/v1/dto/request"
	"emplacc-api/api/v1/dto/response"
	"emplacc-api/api/v1/middleware"
	"emplacc-api/api/v1/presenter"
	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/domain"

	"github.com/labstack/echo/v4"
)

var (
	_ request.CreateTeam
	_ request.UpdateTeam
	_ request.AddTeamMember
	_ request.AddTeamProject
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

	tgroup := group.Group("/team")
	{
		tgroup.GET("", handler.ListTeams)
		tgroup.GET("/:id", handler.GetTeam)
		tgroup.POST("", handler.CreateTeam)
		tgroup.PATCH("/:id", handler.UpdateTeam)
		tgroup.DELETE("/:id", handler.DeleteTeam)

		tgroup.POST("/project", handler.AddTeamToProjectLegacy)
		tgroup.DELETE("/project", handler.RemoveTeamFromProjectLegacy)

		tgroup.POST("/user", handler.AddUserToTeamLegacy)
		tgroup.DELETE("/user", handler.RemoveUserFromTeamLegacy)
		tgroup.GET("/user/:user_id", handler.ListTeamsByUser)
	}
}

// @Summary List Teams
// @Description Retrieve a paginated list of teams
// @Tags Teams
// @Accept json
// @Produce json
// @Param page query int true "Page number"
// @Param page_size query int true "Page size"
// @Success 200 {object} dto.TeamsListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /teams [get]
func (h *TeamHandler) ListTeams(c echo.Context) error {
	params, err := paginationParams(c)
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("pagination_required", err.Error()))
	}
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
	params, err := paginationParams(c)
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("pagination_required", err.Error()))
	}

	projectID, err := parseUUID(c.Param("project_id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid project identifier"))
	}

	teams, err := h.service.ListTeamsByProject(c.Request().Context(), projectID, params)
	if err != nil {
		return respondError(c, http.StatusInternalServerError, dto.NewError("teams_fetch_failed", "failed to list teams"))
	}

	pagination, payload := presenter.MapTeamsPage(teams)
	return respondPaginated(c, http.StatusOK, payload, pagination)
}

// @Summary List Teams By User
// @Description Retrieve teams associated with a specific user
// @Tags Teams
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} dto.TeamsListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /team/user/{id} [get]
func (h *TeamHandler) ListTeamsByUser(c echo.Context) error {
	params, err := paginationParams(c)
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("pagination_required", err.Error()))
	}

	userID, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid user identifier"))
	}

	teams, err := h.service.ListTeamsByUser(c.Request().Context(), userID, params)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", "user identifier is required"))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("teams_fetch_failed", "failed to list teams"))
	}

	pagination, payload := presenter.MapTeamsPage(teams)
	return respondPaginated(c, http.StatusOK, payload, pagination)
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

// @Summary Add Users to Team (legacy)
// @Description Add one or more users to a team using legacy payload format
// @Tags Teams
// @Accept json
// @Produce json
// @Param body body request.TeamAddUsers true "Legacy team-user association"
// @Success 200 {object} map[string]string
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /team/user [post]
func (h *TeamHandler) AddUserToTeamLegacy(c echo.Context) error {
	var req request.TeamAddUsers
	if err := c.Bind(&req); err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", "invalid request body"))
	}

	teamID := strings.TrimSpace(req.TeamID)
	if teamID == "" {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", "team_id is required"))
	}

	teamUUID, err := parseUUID(teamID)
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", "team_id must be a valid UUID"))
	}

	if len(req.UserIDs) == 0 {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", "user_ids are required"))
	}

	cleanedUserIDs := make([]string, 0, len(req.UserIDs))
	for _, rawUserID := range req.UserIDs {
		userID := strings.TrimSpace(rawUserID)
		if userID == "" {
			return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", "user_id cannot be empty"))
		}

		userUUID, err := parseUUID(userID)
		if err != nil {
			return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", "user_id must be a valid UUID"))
		}

		cleanedUserIDs = append(cleanedUserIDs, userUUID.String())

		input := ports.TeamMemberInput{
			TeamID: teamUUID,
			UserID: userUUID,
		}

		if err := h.service.AddUserToTeam(c.Request().Context(), input); err != nil {
			switch {
			case errors.Is(err, domain.ErrInvalidInput):
				return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
			case errors.Is(err, domain.ErrNotFound):
				return respondError(c, http.StatusNotFound, dto.NewError("team_not_found", "team or user not found"))
			default:
				return respondError(c, http.StatusInternalServerError, dto.NewError("team_member_add_failed", "failed to add member"))
			}
		}
	}

	payload := response.UsersAdd{
		TeamID:  teamUUID.String(),
		UserIDs: cleanedUserIDs,
		Message: "Team members added",
	}
	return respondSuccess(c, http.StatusOK, payload)
}

// @Summary Remove User from Team (legacy)
// @Description Remove a user from a team using legacy payload format
// @Tags Teams
// @Accept json
// @Produce json
// @Param body body request.TeamRemoveUser true "Legacy team-user removal"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /team/user [delete]
func (h *TeamHandler) RemoveUserFromTeamLegacy(c echo.Context) error {
	var req request.TeamRemoveUser
	if err := c.Bind(&req); err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", "invalid request body"))
	}

	teamID := strings.TrimSpace(req.TeamID)
	userID := strings.TrimSpace(req.UserID)
	if teamID == "" || userID == "" {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", "team_id and user_id are required"))
	}

	teamUUID, err := parseUUID(teamID)
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", "team_id must be a valid UUID"))
	}
	userUUID, err := parseUUID(userID)
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", "user_id must be a valid UUID"))
	}

	if err := h.service.RemoveUserFromTeam(c.Request().Context(), teamUUID, userUUID); err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		case errors.Is(err, domain.ErrNotFound):
			return respondError(c, http.StatusNotFound, dto.NewError("team_member_not_found", "team member not found"))
		default:
			return respondError(c, http.StatusInternalServerError, dto.NewError("team_member_remove_failed", "failed to remove member"))
		}
	}

	payload := response.TeamUserMessage{
		TeamID:  teamUUID.String(),
		UserID:  userUUID.String(),
		Message: "Team member removed",
	}
	return respondSuccess(c, http.StatusOK, payload)
}

// @Summary Add Team to Project (legacy)
// @Description Associate a team with a project using legacy payload
// @Tags Teams
// @Accept json
// @Produce json
// @Param body body request.TeamProjectRequest true "Team-project association"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /team/project [post]
func (h *TeamHandler) AddTeamToProjectLegacy(c echo.Context) error {
	var req request.TeamProjectRequest
	if err := c.Bind(&req); err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", "invalid request body"))
	}

	teamID := strings.TrimSpace(req.TeamID)
	projectID := strings.TrimSpace(req.ProjectID)
	if teamID == "" || projectID == "" {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", "team_id and project_id are required"))
	}

	teamUUID, err := parseUUID(teamID)
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", "team_id must be a valid UUID"))
	}
	projectUUID, err := parseUUID(projectID)
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", "project_id must be a valid UUID"))
	}

	input := ports.TeamProjectInput{TeamID: teamUUID, ProjectID: projectUUID}
	if err := h.service.AddTeamToProject(c.Request().Context(), input); err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		case errors.Is(err, domain.ErrNotFound):
			return respondError(c, http.StatusNotFound, dto.NewError("team_not_found", "team or project not found"))
		default:
			return respondError(c, http.StatusInternalServerError, dto.NewError("team_project_add_failed", "failed to associate team with project"))
		}
	}

	payload := response.TeamProjectMessage{
		TeamID:    teamUUID.String(),
		ProjectID: projectUUID.String(),
		Message:   "Project linked to team",
	}
	return respondSuccess(c, http.StatusOK, payload)
}

// @Summary Remove Team from Project (legacy)
// @Description Remove a team-project association using legacy payload
// @Tags Teams
// @Accept json
// @Produce json
// @Param body body request.TeamProjectRequest true "Team-project association"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /team/project [delete]
func (h *TeamHandler) RemoveTeamFromProjectLegacy(c echo.Context) error {
	var req request.TeamProjectRequest
	if err := c.Bind(&req); err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", "invalid request body"))
	}

	teamID := strings.TrimSpace(req.TeamID)
	projectID := strings.TrimSpace(req.ProjectID)
	if teamID == "" || projectID == "" {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", "team_id and project_id are required"))
	}

	teamUUID, err := parseUUID(teamID)
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", "team_id must be a valid UUID"))
	}
	projectUUID, err := parseUUID(projectID)
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", "project_id must be a valid UUID"))
	}

	if err := h.service.RemoveTeamFromProject(c.Request().Context(), teamUUID, projectUUID); err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		case errors.Is(err, domain.ErrNotFound):
			return respondError(c, http.StatusNotFound, dto.NewError("team_project_not_found", "association not found"))
		default:
			return respondError(c, http.StatusInternalServerError, dto.NewError("team_project_remove_failed", "failed to remove association"))
		}
	}

	payload := response.TeamProjectMessage{
		TeamID:    teamUUID.String(),
		ProjectID: projectUUID.String(),
		Message:   "Project unlinked from team",
	}
	return respondSuccess(c, http.StatusOK, payload)
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
