package data

import (
	"encoding/json"
	"os/exec"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/mrf/dcc/internal/config"
)

// PullRequest represents a GitHub pull request
type PullRequest struct {
	Number         int
	Title          string
	Repo           string
	URL            string
	AgeDays        int64
	ReviewDecision string
	CIStatus       string // "success", "failure", "pending", or ""
}

// PrsPanel holds PR data for display
type PrsPanel struct {
	NeedsReview []PullRequest
	YourPrs     []PullRequest
	IsLoading   bool
}

// ghSearchPr represents the JSON structure from gh CLI
type ghSearchPr struct {
	Number     int    `json:"number"`
	Title      string `json:"title"`
	URL        string `json:"url"`
	CreatedAt  string `json:"createdAt"`
	Repository struct {
		NameWithOwner string `json:"nameWithOwner"`
	} `json:"repository"`
	ReviewDecision string `json:"reviewDecision"`
}

// ghCheckRollup represents the status check rollup from gh pr view
type ghCheckRollup struct {
	StatusCheckRollup []struct {
		State      string `json:"state"`
		Status     string `json:"status"`
		Conclusion string `json:"conclusion"`
	} `json:"statusCheckRollup"`
}

// FetchPrs retrieves PRs from GitHub CLI
func FetchPrs(cfg config.PrsConfig) PrsPanel {
	if !cfg.Enabled {
		return PrsPanel{}
	}

	// Check if gh CLI is available
	if _, err := exec.LookPath("gh"); err != nil {
		return PrsPanel{}
	}

	var wg sync.WaitGroup
	var needsReview, yourPrs []PullRequest
	var mu sync.Mutex

	wg.Add(2)

	// Fetch PRs that need review
	go func() {
		defer wg.Done()
		prs := fetchReviewRequested()
		mu.Lock()
		needsReview = prs
		mu.Unlock()
	}()

	// Fetch user's own PRs
	go func() {
		defer wg.Done()
		prs := fetchAuthoredPrs()
		mu.Lock()
		yourPrs = prs
		mu.Unlock()
	}()

	wg.Wait()

	// Fetch CI status for authored PRs
	enrichWithCIStatus(yourPrs)

	// Sort needs_review by age (oldest first)
	sort.Slice(needsReview, func(i, j int) bool {
		return needsReview[i].AgeDays > needsReview[j].AgeDays
	})

	// Sort your_prs by age (newest first)
	sort.Slice(yourPrs, func(i, j int) bool {
		return yourPrs[i].AgeDays < yourPrs[j].AgeDays
	})

	return PrsPanel{
		NeedsReview: needsReview,
		YourPrs:     yourPrs,
	}
}

func fetchReviewRequested() []PullRequest {
	cmd := exec.Command("gh", "search", "prs",
		"--review-requested=@me",
		"--state=open",
		"--json", "number,title,url,createdAt,repository,reviewDecision",
		"--limit", "10")

	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	return parseGhPrs(output)
}

func fetchAuthoredPrs() []PullRequest {
	cmd := exec.Command("gh", "search", "prs",
		"--author=@me",
		"--state=open",
		"--json", "number,title,url,createdAt,repository,reviewDecision",
		"--limit", "10")

	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	return parseGhPrs(output)
}

// fetchCIStatus fetches CI check status for a single PR
func fetchCIStatus(repo string, number int) string {
	cmd := exec.Command("gh", "pr", "view", strconv.Itoa(number),
		"--repo", repo,
		"--json", "statusCheckRollup")

	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	var rollup ghCheckRollup
	if err := json.Unmarshal(output, &rollup); err != nil {
		return ""
	}

	if len(rollup.StatusCheckRollup) == 0 {
		return ""
	}

	hasFailure := false
	hasPending := false
	for _, check := range rollup.StatusCheckRollup {
		switch {
		case check.Conclusion == "FAILURE" || check.Conclusion == "TIMED_OUT" ||
			check.Conclusion == "CANCELLED" || check.State == "FAILURE" || check.State == "ERROR":
			hasFailure = true
		case check.Status == "IN_PROGRESS" || check.Status == "QUEUED" ||
			check.Status == "PENDING" || check.State == "PENDING":
			hasPending = true
		}
	}

	switch {
	case hasFailure:
		return "failure"
	case hasPending:
		return "pending"
	default:
		return "success"
	}
}

// enrichWithCIStatus fetches CI status concurrently for all PRs
func enrichWithCIStatus(prs []PullRequest) {
	var wg sync.WaitGroup
	for i := range prs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			prs[i].CIStatus = fetchCIStatus(prs[i].Repo, prs[i].Number)
		}()
	}
	wg.Wait()
}

func parseGhPrs(output []byte) []PullRequest {
	var ghPrs []ghSearchPr
	if err := json.Unmarshal(output, &ghPrs); err != nil {
		return nil
	}

	prs := make([]PullRequest, 0, len(ghPrs))
	now := time.Now()

	for _, ghPr := range ghPrs {
		createdAt, err := time.Parse(time.RFC3339, ghPr.CreatedAt)
		if err != nil {
			continue
		}

		ageDays := int64(now.Sub(createdAt).Hours() / 24)

		prs = append(prs, PullRequest{
			Number:         ghPr.Number,
			Title:          ghPr.Title,
			Repo:           ghPr.Repository.NameWithOwner,
			URL:            ghPr.URL,
			AgeDays:        ageDays,
			ReviewDecision: ghPr.ReviewDecision,
		})
	}

	return prs
}

// OpenFirstPr opens the first PR that needs review (or first of user's PRs) in browser
func OpenFirstPr(panel PrsPanel) error {
	var url string
	if len(panel.NeedsReview) > 0 {
		url = panel.NeedsReview[0].URL
	} else if len(panel.YourPrs) > 0 {
		url = panel.YourPrs[0].URL
	} else {
		return nil
	}

	return exec.Command("open", url).Run()
}
