package conventions

import (
	"fmt"
	"path/filepath"
	"strings"
)

// PrometheusMetricsPathName is the path where metrics will be served.
const PrometheusMetricsPathName = "metrics"

// IRHistoryAtomFeedPathName is the path where history Atom feed will be created.
const IRHistoryAtomFeedPathName = "history-feed.atom"

// IRDetailURL standardizes the URL for serving an incident report detail on an URL.
func IRDetailURL(baseURL, irID string) string {
	baseURL = strings.TrimSuffix(baseURL, "/")
	return fmt.Sprintf("%s/ir/%s", baseURL, irID)
}

// IRDetailFilePath standardizes the file path for read/storing an incident report detail on an FS.
func IRDetailFilePath(basePath, irID string) string {
	basePath = filepath.Clean(basePath)
	return fmt.Sprintf("%s/ir/%s.html", basePath, irID)
}

// IRHistoryURL standardizes the URL for serving an incident report history on an URL.
func IRHistoryURL(baseURL string, page int) string {
	baseURL = strings.TrimSuffix(baseURL, "/")
	return fmt.Sprintf("%s/history/%d", baseURL, page)
}

// IRHistoryFilePath standardizes the file path for read/storing an incident report history on an FS.
func IRHistoryFilePath(basePath string, page int) string {
	basePath = filepath.Clean(basePath)
	return fmt.Sprintf("%s/history/%d.html", basePath, page)
}

// IndexFilePath standardizes the file path for read/storing the site index on an FS.
func IndexFilePath(basePath string) string {
	basePath = filepath.Clean(basePath)
	return fmt.Sprintf("%s/index.html", basePath)
}
