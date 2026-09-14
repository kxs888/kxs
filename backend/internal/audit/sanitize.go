package audit

import (
	"strings"

	"github.com/kxs888/kxs/backend/internal/obs"
)

var forbiddenDetailKeys = map[string]struct{}{
	"medical_record": {},
	"note_full":      {},
	"note":           {},
	"assessment":     {},
	"phi":            {},
	"password":       {},
	"passwd":         {},
	"access_token":   {},
	"refresh_token":  {},
	"jwt":            {},
	"token":          {},
	"authorization":  {},
	"body":           {},
	"content":        {},
}

// SanitizeDetail 去掉病历全文、JWT、口令等字段。未知键若值像 JWT 也丢弃。
func SanitizeDetail(in map[string]any) map[string]any {
	if in == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		if _, bad := forbiddenDetailKeys[strings.ToLower(k)]; bad {
			continue
		}
		if s, ok := v.(string); ok && (obs.LooksLikeJWT(s) || looksLikeBearer(s)) {
			continue
		}
		out[k] = v
	}
	return out
}

func looksLikeBearer(s string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(s)), "bearer ")
}
