package data

import "testing"

const samplePrJSON = `[
	{
		"createdAt": "2025-12-10T20:20:26Z",
		"number": 123,
		"repository": {"nameWithOwner": "example-org/example-repo"},
		"title": "Add new feature",
		"url": "https://github.com/example-org/example-repo/pull/123"
	},
	{
		"createdAt": "2025-12-11T23:29:52Z",
		"number": 456,
		"repository": {"nameWithOwner": "example-org/example-repo"},
		"title": "Fix bug in parser",
		"url": "https://github.com/example-org/example-repo/pull/456"
	}
]`

func TestParsePrJSONExtractsFieldsCorrectly(t *testing.T) {
	prs := parseGhPrs([]byte(samplePrJSON))

	if len(prs) != 2 {
		t.Fatalf("expected 2 PRs, got %d", len(prs))
	}

	if prs[0].Number != 123 {
		t.Errorf("expected number 123, got %d", prs[0].Number)
	}
	if prs[0].Title != "Add new feature" {
		t.Errorf("unexpected title: %s", prs[0].Title)
	}
	if prs[0].Repo != "example-org/example-repo" {
		t.Errorf("unexpected repo: %s", prs[0].Repo)
	}
	if prs[0].URL != "https://github.com/example-org/example-repo/pull/123" {
		t.Errorf("unexpected URL: %s", prs[0].URL)
	}
}

func TestParsePrJSONHandlesEmptyArray(t *testing.T) {
	prs := parseGhPrs([]byte("[]"))

	if len(prs) != 0 {
		t.Errorf("expected 0 PRs, got %d", len(prs))
	}
}

func TestParsePrJSONHandlesInvalidJSON(t *testing.T) {
	prs := parseGhPrs([]byte("not valid json"))

	if prs != nil && len(prs) != 0 {
		t.Errorf("expected nil or empty for invalid JSON, got %d PRs", len(prs))
	}
}

func TestParsePrJSONCalculatesAge(t *testing.T) {
	prs := parseGhPrs([]byte(samplePrJSON))

	if len(prs) < 1 {
		t.Fatal("expected at least 1 PR")
	}

	// Age should be positive (in the past)
	if prs[0].AgeDays < 0 {
		t.Errorf("expected positive age, got %d", prs[0].AgeDays)
	}
}
