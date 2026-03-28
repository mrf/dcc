package data

import (
	"encoding/json"
	"os/exec"
	"sort"
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

// OpenPrByIndex opens the PR at the given cursor index in browser.
// Items are indexed across NeedsReview (up to 4) then YourPrs (up to 4).
func OpenPrByIndex(panel PrsPanel, idx int) error {
	needsReviewVisible := len(panel.NeedsReview)
	if needsReviewVisible > 4 {
		needsReviewVisible = 4
	}

	if idx < needsReviewVisible {
		return exec.Command("open", panel.NeedsReview[idx].URL).Run()
	}

	yourIdx := idx - needsReviewVisible
	yourPrsVisible := len(panel.YourPrs)
	if yourPrsVisible > 4 {
		yourPrsVisible = 4
	}

	if yourIdx >= 0 && yourIdx < yourPrsVisible {
		return exec.Command("open", panel.YourPrs[yourIdx].URL).Run()
	}

	return nil
}
