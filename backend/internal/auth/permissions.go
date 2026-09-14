package auth

// PlaceholderPermissions 是 S0 RBAC 占位（B7）。不代表已实现患者/任务/报告 CRUD。
var PlaceholderPermissions = []string{
	"patient.view",
	"task.create",
	"report.view",
}

func PermissionsOrPlaceholder(in []string) []string {
	if len(in) == 0 {
		out := make([]string, len(PlaceholderPermissions))
		copy(out, PlaceholderPermissions)
		return out
	}
	return in
}
