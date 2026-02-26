package rbac

const (
	ViewStocks         = "view_stocks"
	OrderStocks        = "order_stocks"
	ManageOwnPortfolio = "manage_own_portfolio"
	ManageOwnWatchlist = "manage_own_watchlist"
	ManageOwnProfile   = "manage_own_profile"
	ManageOwnAccounts  = "manage_own_accounts"
	ViewUsers          = "view_users"
	ViewAdminDashboard = "view_admin_dashboard"
	ViewAuditLogs      = "view_audit_logs"
	DeactivateUsers    = "deactivate_users"
	ReactivateUsers    = "reactivate_users"
	ManageRoles        = "manage_roles"
)

var validPermissions = map[string]struct{}{
	ViewStocks:         {},
	OrderStocks:        {},
	ManageOwnPortfolio: {},
	ManageOwnWatchlist: {},
	ManageOwnProfile:   {},
	ManageOwnAccounts:  {},
	ViewUsers:          {},
	ViewAdminDashboard: {},
	ViewAuditLogs:      {},
	DeactivateUsers:    {},
	ReactivateUsers:    {},
	ManageRoles:        {},
}

func IsValidPermission(p string) bool {
	_, ok := validPermissions[p]
	return ok
}
