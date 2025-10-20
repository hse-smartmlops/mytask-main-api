package dto

import "emplacc-api/api/v1/dto/response"

type AttendancesListResponse struct {
	SuccessResponse[response.AttendancesPage]
}

type AttendanceResponse struct {
	SuccessResponse[response.Attendance]
}

type AttendancesByUserResponse struct {
	SuccessResponse[response.AttendancesByUser]
}

type TasksListResponse struct {
	SuccessResponse[response.TasksPage]
}

type TaskResponse struct {
	SuccessResponse[response.Task]
}

type ReportsListResponse struct {
	SuccessResponse[response.ReportsPage]
}

type ImprovedTaskReportResponse struct {
	SuccessResponse[response.ImprovedTaskReport]
}

type TaskBoardProjectResponse struct {
	SuccessResponse[response.TaskBoardProject]
}

type CompletedWorkItemResponse struct {
	SuccessResponse[response.CompletedWorkItem]
}

type ReportResponse struct {
	SuccessResponse[response.Report]
}

type TomorrowPlanItemResponse struct {
	SuccessResponse[response.TomorrowPlanItem]
}

type HelpRequestItemResponse struct {
	SuccessResponse[response.HelpRequestItem]
}

type HelpRequestsForUserResponse struct {
	SuccessResponse[response.HelpRequestsForUser]
}

type ReportUniversalResponse struct {
	SuccessResponse[response.ReportUniversal]
}

type BoardsListResponse struct {
	SuccessResponse[response.BoardsPage]
}

type BoardsListByProjectResponse struct {
	SuccessResponse[[]response.Board]
}

type BoardResponse struct {
	SuccessResponse[response.Board]
}

type UsersListResponse struct {
	SuccessResponse[response.UsersPage]
}

type UserResponse struct {
	SuccessResponse[response.User]
}

type SubscriptionsListResponse struct {
	SuccessResponse[response.SubscriptionsPage]
}

type SubscriptionResponse struct {
	SuccessResponse[response.Subscription]
}

type ForumMessagesListResponse struct {
	SuccessResponse[response.ForumMessagesPage]
}

type ForumMessageResponse struct {
	SuccessResponse[response.ForumMessage]
}

type ProblemsListResponse struct {
	SuccessResponse[response.ProblemsPage]
}

type ProblemResponse struct {
	SuccessResponse[response.Problem]
}

type TeamsListResponse struct {
	SuccessResponse[response.TeamsPage]
}

type TeamResponse struct {
	SuccessResponse[response.Team]
}

type ProjectsListResponse struct {
	SuccessResponse[response.ProjectsPage]
}

type ProjectResponse struct {
	SuccessResponse[response.Project]
}

type RolesListResponse struct {
	SuccessResponse[response.RolesPage]
}

type RoleResponse struct {
	SuccessResponse[response.Role]
}

type StatusesListResponse struct {
	SuccessResponse[response.StatusesPage]
}

type StatusResponse struct {
	SuccessResponse[response.Status]
}

type AuthResponse struct {
	SuccessResponse[response.Auth]
}

type RefreshResponse struct {
	SuccessResponse[response.Refresh]
}

type UserInfoResponse struct {
	SuccessResponse[response.UserInfo]
}

type TokenValidationResponse struct {
	SuccessResponse[response.TokenValidation]
}
