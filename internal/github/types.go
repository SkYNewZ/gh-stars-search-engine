package github

import "time"

/*
	query ($cursor: String) {
	  viewer {
	    login
	    starredRepositories(first: 3, after: $cursor) {
	      totalCount
	      pageInfo {
	        endCursor
	        startCursor
	        hasNextPage
	        hasPreviousPage
	      }
	      repositories: edges {
	        starredAt
	        repository: node {
	          id
	          nameWithOwner
	          description
	          url
	          r1:object(expression: "HEAD:README.md") {
	            ... on Blob {
	              text
	            }
	          }
	          r2:object(expression: "HEAD:readme.md") {
	            ... on Blob {
	              text
	            }
	          }
	          primaryLanguage {
	            id
	            name
	            color
	          }
	        }
	      }
	    }
	  }
	  rateLimit {
	    cost
	    limit
	    remaining
	    used
	    resetAt
	  }
	}
*/
type query struct {
	// Viewer is the currently authenticated user
	Viewer struct {
		// Login is the username of the user
		Login string `graphql:"login" json:"login"`

		// StarredRepositories is the list of repositories starred by the user
		StarredRepositories *StarredRepositories `graphql:"starredRepositories(first: $count, after: $cursor)" json:"starred_repositories"`
	} `graphql:"viewer" json:"viewer"`

	// RateLimit contains the rate limit information
	RateLimit *RateLimit `graphql:"rateLimit" json:"rate_limit"`
}

// StarredRepositories is the list of repositories starred by the user.
type StarredRepositories struct {
	TotalCount   int                  `graphql:"totalCount" json:"total_count"`
	PageInfo     *PageInfo            `graphql:"pageInfo"   json:"page_info"`
	Repositories []*StarredRepository `graphql:"edges"      json:"repositories"`
}

// PageInfo contains information about the current page.
type PageInfo struct {
	StartCursor     string `graphql:"startCursor"     json:"start_cursor"`
	EndCursor       string `graphql:"endCursor"       json:"end_cursor"`
	HasNextPage     bool   `graphql:"hasNextPage"     json:"has_next_page"`
	HasPreviousPage bool   `graphql:"hasPreviousPage" json:"has_previous_page"`
}

// RateLimit contains the rate limit information.
type RateLimit struct {
	Cost      int       `graphql:"cost"      json:"cost"`
	Limit     int       `graphql:"limit"     json:"limit"`
	Remaining int       `graphql:"remaining" json:"remaining"`
	Used      int       `graphql:"used"      json:"used"`
	ResetAt   time.Time `graphql:"resetAt"   json:"reset_at"`
}

// StarredRepository is a repository starred by the user.
type StarredRepository struct {
	StarredAt  time.Time   `graphql:"starredAt" json:"starred_at"`
	Repository *Repository `graphql:"node"      json:"repository"`
}

// Repository is a GitHub repository.
type Repository struct {
	ID            string `graphql:"id"            json:"id"`
	NameWithOwner string `graphql:"nameWithOwner" json:"name_with_owner"`
	Description   string `graphql:"description"   json:"description"`
	URL           string `graphql:"url"           json:"url"`

	R1     *readme `graphql:"r1:object(expression: \"HEAD:README.md\")" json:"-"`      // content of README.md
	R2     *readme `graphql:"r2:object(expression: \"HEAD:readme.md\")" json:"-"`      // content of readme.md
	Readme string  `graphql:"-"                                         json:"readme"` // computed field

	PrimaryLanguage struct {
		ID    string `graphql:"id"    json:"id"`
		Name  string `graphql:"name"  json:"name"`
		Color string `graphql:"color" json:"color"`
	} `graphql:"primaryLanguage" json:"primary_language"`
}

type readme struct {
	Blob struct {
		Text string `graphql:"text" json:"-"`
	} `graphql:"... on Blob" json:"-"`
}

// GetID returns the repository ID.
func (r *Repository) GetID() string {
	return r.ID
}
