package alerts

import (
	"net/http"
	"sort"
	"strings"

	"github.com/riportdev/riport/plus/capabilities/alerting/entities/rules"
	"github.com/riportdev/riport/server/api/errors"
	"github.com/riportdev/riport/share/query"
)

var SupportedProblemsListFields = map[string][]string{
	"fields[problems]": {
		"problem_id",
		"rule_id",
		"client_id",
		"client_name",
		"active",
		"created_at",
		"resolved_at",
	},
}

var SupportedProblemsSorts = map[string]bool{
	"problem_id":  false,
	"rule_id":     true,
	"client_id":   true,
	"client_name": true,
	"active":      true,
	"created_at":  true,
	"resolved_at": true,
}

var SupportedProblemsFilters = map[string]bool{
	"problem_id":      true,
	"rule_id":         true,
	"client_id":       true,
	"client_name":     true,
	"active":          true,
	"created_at[gt]":  true,
	"resolved_at[gt]": true,
	"created_at[lt]":  true,
	"resolved_at[lt]": true,
	"created_at[eq]":  true,
	"resolved_at[eq]": true,
}

var SupportedProblemsFields = map[string]map[string]bool{
	"problems": {
		"problem_id":  true,
		"rule_id":     true,
		"client_id":   true,
		"client_name": true,
		"active":      true,
		"created_at":  true,
		"resolved_at": true,
	},
}

func SortProblemsByRuleID(problems []*rules.Problem, desc bool) {
	sort.Slice(problems, func(i, j int) bool {
		less := strings.ToLower(string(problems[i].RuleID)) < strings.ToLower(string(problems[j].RuleID))
		if desc {
			return !less
		}
		return less
	})
}

func SortProblemsByClientID(problems []*rules.Problem, desc bool) {
	sort.Slice(problems, func(i, j int) bool {
		less := strings.ToLower(problems[i].ClientID) < strings.ToLower(problems[j].ClientID)
		if desc {
			return !less
		}
		return less
	})
}

func SortProblemsByState(problems []*rules.Problem, desc bool) {
	sort.Slice(problems, func(i, j int) bool {
		less := problems[i].Active == problems[j].Active
		if desc {
			return !less
		}
		return less
	})
}

func SortProblemsByCreatedAt(problems []*rules.Problem, desc bool) {
	sort.Slice(problems, func(i, j int) bool {
		less := problems[i].CreatedAt.Before(problems[j].CreatedAt)
		if desc {
			return !less
		}
		return less
	})
}

func SortProblemsByResolvedAt(problems []*rules.Problem, desc bool) {
	sort.Slice(problems, func(i, j int) bool {
		less := problems[i].ResolvedAt.Before(problems[j].ResolvedAt.ToTime())
		if desc {
			return !less
		}
		return less
	})
}

func SortProblemsFunc(sorts []query.SortOption) (sortFunc func(a []*rules.Problem, desc bool), desc bool, err error) {
	if len(sorts) < 1 {
		return SortProblemsByCreatedAt, true, nil
	}
	if len(sorts) > 1 {
		return nil, false, errors.APIError{
			Message:    "Only one sort field is supported for problems.",
			HTTPStatus: http.StatusBadRequest,
		}
	}

	switch sorts[0].Column {
	case "rule_id":
		sortFunc = SortProblemsByRuleID
	case "client_id":
		sortFunc = SortProblemsByClientID
	case "state":
		sortFunc = SortProblemsByState
	case "created_at":
		sortFunc = SortProblemsByCreatedAt
	case "resolved_at":
		sortFunc = SortProblemsByResolvedAt
	}

	return sortFunc, !sorts[0].IsASC, nil
}
