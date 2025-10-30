package pg

import "gorm.io/gorm"

func boolPtr(value bool) *bool {
	return &value
}

func preloadReportRelations(
	db *gorm.DB,
) *gorm.DB {
	return db.
		Preload("User", "users.deleted = FALSE OR users.deleted IS NULL").
		Preload("CompletedWork", "completed_works.deleted = FALSE OR completed_works.deleted IS NULL").
		Preload("CompletedWork.Task").
		Preload("HelpRequests", "help_requests.deleted = FALSE OR help_requests.deleted IS NULL").
		Preload("TomorrowPlans", "tomorrow_plans.deleted = FALSE OR tomorrow_plans.deleted IS NULL").
		Preload("ReportProblems", "report_problems.deleted = FALSE OR report_problems.deleted IS NULL").
		Preload("ReportProblems.Problem", "problems.deleted = FALSE OR problems.deleted IS NULL")
}
