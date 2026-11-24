package engine_test

import (
	"context"
	"testing"

	"refweir.dev/internal/engine"
	"refweir.dev/internal/store"
)

func TestEvaluateMerge_Ruleset(t *testing.T) {
	ctx := context.Background()

	// 1. No rulesets -> allowed
	res := engine.EvaluateMerge(ctx, nil, "main", "main", "user_1", "", 0, nil)
	if !res.Allowed || res.Verdict != "mergeable" {
		t.Fatalf("expected allowed, got %+v", res)
	}

	// 2. Pull Request rule requiring 2 approvals
	rsPR := store.Ruleset{
		Enforcement:      "active",
		Kind:             "merge_branch",
		IncludedRefNames: []string{"~DEFAULT_BRANCH"},
		Rules: []store.RulesetRule{
			{
				RuleType: "pull_request",
				Parameters: map[string]any{
					"requiredApprovingReviewCount": float64(2),
				},
			},
		},
	}

	// 0 approvals -> blocked
	res = engine.EvaluateMerge(ctx, []store.Ruleset{rsPR}, "main", "main", "user_1", "", 0, nil)
	if res.Allowed || res.Verdict != "blocked" || len(res.Blockers) != 1 {
		t.Fatalf("expected blocked with 1 blocker, got %+v", res)
	}
	if res.Blockers[0].Kind != "required_approvals" {
		t.Fatalf("expected kind required_approvals, got %s", res.Blockers[0].Kind)
	}

	// 2 approvals -> allowed
	res = engine.EvaluateMerge(ctx, []store.Ruleset{rsPR}, "main", "main", "user_1", "", 2, nil)
	if !res.Allowed || res.Verdict != "mergeable" {
		t.Fatalf("expected allowed with 2 approvals, got %+v", res)
	}

	// 3. Status checks rule
	rsChecks := store.Ruleset{
		Enforcement:      "active",
		Kind:             "merge_branch",
		IncludedRefNames: []string{"refs/heads/main"},
		Rules: []store.RulesetRule{
			{
				RuleType: "require_status_checks",
				Parameters: map[string]any{
					"requiredChecks": []any{
						map[string]any{
							"groupKey": "ci",
							"runKey":   "build",
							"name":     "CI Build",
						},
					},
				},
			},
		},
	}

	// Missing check run -> blocked
	res = engine.EvaluateMerge(ctx, []store.Ruleset{rsChecks}, "main", "main", "user_1", "", 0, nil)
	if res.Allowed || len(res.Blockers) != 1 || res.Blockers[0].Kind != "required_checks" {
		t.Fatalf("expected blocked on required_checks, got %+v", res)
	}

	// Passing check run -> allowed
	passingRun := store.CheckRun{
		Key:        "build",
		Name:       "CI Build",
		Status:     "completed",
		Conclusion: "success",
	}
	res = engine.EvaluateMerge(ctx, []store.Ruleset{rsChecks}, "main", "main", "user_1", "", 0, []store.CheckRun{passingRun})
	if !res.Allowed || res.Verdict != "mergeable" {
		t.Fatalf("expected allowed with passing check run, got %+v", res)
	}

	// 4. Bypass actor
	rsWithBypass := rsPR
	rsWithBypass.BypassActors = []store.RulesetBypassActor{
		{
			BypassMode: "always",
			User:       &store.BypassUser{ID: "admin_user"},
		},
	}

	res = engine.EvaluateMerge(ctx, []store.Ruleset{rsWithBypass}, "main", "main", "admin_user", "", 0, nil)
	if !res.Allowed || res.Verdict != "mergeable" {
		t.Fatalf("expected bypass actor to be allowed, got %+v", res)
	}
}

func TestEvaluateMerge_RequireLinearHistory(t *testing.T) {
	ctx := context.Background()

	rsLinear := store.Ruleset{
		Enforcement:      "active",
		Kind:             "merge_branch",
		IncludedRefNames: []string{"refs/heads/main"},
		Rules: []store.RulesetRule{
			{
				RuleType: "require_linear_history",
			},
		},
	}

	// 1. Standard merge commit -> blocked
	optsMerge := engine.MergeEvaluationOptions{
		MergeMethod: "merge",
	}
	res := engine.EvaluateMergeWithOptions(ctx, []store.Ruleset{rsLinear}, "main", "main", "user_1", "", 0, nil, optsMerge)
	if res.Allowed || len(res.Blockers) != 1 || res.Blockers[0].Kind != "require_linear_history" {
		t.Fatalf("expected blocked with require_linear_history blocker, got %+v", res)
	}

	// 2. Squash merge -> allowed
	optsSquash := engine.MergeEvaluationOptions{
		MergeMethod: "squash",
	}
	res = engine.EvaluateMergeWithOptions(ctx, []store.Ruleset{rsLinear}, "main", "main", "user_1", "", 0, nil, optsSquash)
	if !res.Allowed || res.Verdict != "mergeable" {
		t.Fatalf("expected squash merge to be allowed, got %+v", res)
	}

	// 3. Rebase merge -> allowed
	optsRebase := engine.MergeEvaluationOptions{
		MergeMethod: "rebase",
	}
	res = engine.EvaluateMergeWithOptions(ctx, []store.Ruleset{rsLinear}, "main", "main", "user_1", "", 0, nil, optsRebase)
	if !res.Allowed || res.Verdict != "mergeable" {
		t.Fatalf("expected rebase merge to be allowed, got %+v", res)
	}
}

func TestEvaluateMerge_RequireSignedCommits(t *testing.T) {
	ctx := context.Background()

	rsSigned := store.Ruleset{
		Enforcement:      "active",
		Kind:             "merge_branch",
		IncludedRefNames: []string{"refs/heads/main"},
		Rules: []store.RulesetRule{
			{
				RuleType: "require_signed_commits",
			},
		},
	}

	// 1. Unsigned commits present -> blocked
	signedFalse := false
	optsUnsigned := engine.MergeEvaluationOptions{
		CommitsSigned:        &signedFalse,
		UnsignedCommitsCount: 2,
	}
	res := engine.EvaluateMergeWithOptions(ctx, []store.Ruleset{rsSigned}, "main", "main", "user_1", "", 0, nil, optsUnsigned)
	if res.Allowed || len(res.Blockers) != 1 || res.Blockers[0].Kind != "require_signed_commits" {
		t.Fatalf("expected blocked with require_signed_commits blocker, got %+v", res)
	}

	// 2. All commits verified -> allowed
	signedTrue := true
	optsSigned := engine.MergeEvaluationOptions{
		CommitsSigned:        &signedTrue,
		UnsignedCommitsCount: 0,
	}
	res = engine.EvaluateMergeWithOptions(ctx, []store.Ruleset{rsSigned}, "main", "main", "user_1", "", 0, nil, optsSigned)
	if !res.Allowed || res.Verdict != "mergeable" {
		t.Fatalf("expected all signed commits to be allowed, got %+v", res)
	}
}

