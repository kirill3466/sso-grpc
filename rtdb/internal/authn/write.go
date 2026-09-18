package authn

import "strings"

type WriteAccess struct {
	userIDs map[int64]struct{}
	emails  map[string]struct{}
}

func NewWriteAccess(userIDs []int64, emails []string) WriteAccess {
	ids := make(map[int64]struct{}, len(userIDs))
	for _, id := range userIDs {
		if id > 0 {
			ids[id] = struct{}{}
		}
	}

	mails := make(map[string]struct{}, len(emails))
	for _, email := range emails {
		email = strings.ToLower(strings.TrimSpace(email))
		if email != "" {
			mails[email] = struct{}{}
		}
	}

	return WriteAccess{userIDs: ids, emails: mails}
}

func (a WriteAccess) Allows(claims *Claims) bool {
	if claims == nil {
		return false
	}
	if _, ok := a.userIDs[claims.UID]; ok {
		return true
	}
	_, ok := a.emails[strings.ToLower(claims.Email)]
	return ok
}
