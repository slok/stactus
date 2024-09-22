package feed

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/feeds"

	"github.com/slok/stactus/internal/conventions"
	"github.com/slok/stactus/internal/model"
	utilfs "github.com/slok/stactus/internal/util/fs"
)

type RepositoryConfig struct {
	FileManager         utilfs.FileManager
	AtomHistoryFilePath string
	HistoryItemsPerFeed int
	TimeNow             func() time.Time
}

func (c *RepositoryConfig) defaults() error {
	if c.FileManager == nil {
		c.FileManager = utilfs.StdFileManager
	}

	c.AtomHistoryFilePath = filepath.Clean(c.AtomHistoryFilePath)
	if c.AtomHistoryFilePath == "" {
		return fmt.Errorf("atom history feed file path is required")
	}

	if !strings.HasSuffix(c.AtomHistoryFilePath, conventions.IRHistoryAtomFeedPathName) {
		return fmt.Errorf("atom history feed must end with %q path", conventions.IRHistoryAtomFeedPathName)
	}

	if c.HistoryItemsPerFeed == 0 {
		c.HistoryItemsPerFeed = 25 // Same as Atlassian Status page.
	}

	if c.TimeNow == nil {
		c.TimeNow = func() time.Time { return time.Now().UTC() }
	}

	return nil
}

type Repository struct {
	fileManager         utilfs.FileManager
	atomHistoryFilePath string
	historyItemsPerFeed int
	timeNow             func() time.Time
}

func NewFSRepository(config RepositoryConfig) (*Repository, error) {
	err := config.defaults()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &Repository{
		fileManager:         config.FileManager,
		atomHistoryFilePath: config.AtomHistoryFilePath,
		historyItemsPerFeed: config.HistoryItemsPerFeed,
		timeNow:             config.TimeNow,
	}, nil
}

func (r Repository) CreateHistoryFeed(ctx context.Context, ui model.UI) error {
	now := r.timeNow()
	feed := &feeds.Feed{
		Title:       fmt.Sprintf("%s - Incident history", ui.Settings.Name),
		Link:        &feeds.Link{Rel: "alternate", Type: "text/html", Href: ui.Settings.URL},
		Description: fmt.Sprintf("%s status page", ui.Settings.Name),
		Author:      &feeds.Author{Name: ui.Settings.Name},
		Updated:     now,
		Id:          ui.Settings.URL,
	}

	history := ui.History
	if len(ui.History) > r.historyItemsPerFeed {
		history = ui.History[:r.historyItemsPerFeed]
	}

	for _, ir := range history {
		if len(ir.Timeline) < 1 {
			continue
		}

		var b bytes.Buffer
		data := []irHTMLItemData{}
		for _, e := range ir.Timeline {
			data = append(data, irHTMLItemData{
				Kind:        string(e.Kind),
				Description: e.Description,
				TS:          e.TS.UTC().Format(time.RFC3339),
			})
		}

		err := irHTMLItemsTpl.Execute(&b, data)
		if err != nil {
			return fmt.Errorf("could not render Atom entry content: %w", err)
		}

		url := conventions.IRDetailURL(ui.Settings.URL, ir.ID)
		feed.Add(&feeds.Item{
			Title:   ir.Name,
			Link:    &feeds.Link{Rel: "alternate", Type: "text/html", Href: url},
			Content: b.String(),
			Created: ir.Start,
			Updated: ir.Timeline[0].TS, // Latest.
			Id:      url,
		})
	}

	atomFeed, err := feed.ToAtom()
	if err != nil {
		return fmt.Errorf("could not render Atom feed: %w", err)
	}
	err = r.fileManager.WriteFile(ctx, r.atomHistoryFilePath, []byte(atomFeed))
	if err != nil {
		return fmt.Errorf("could not write Atom feed: %w", err)
	}

	// TODO(slok): Slack RSS: To receive live status updates in Slack, copy and paste the text below into the Slack channel of your choice.: `/feed subscribe https://linearstatus.com/slack.rss`

	return nil
}

type irHTMLItemData struct {
	TS          string
	Kind        string
	Description string
}

var irHTMLItemsTpl = template.Must(template.New("").Parse(`
{{ range . }}
<p>
    <small>{{ .TS }}</small>
    <br />
    <strong>{{ .Kind }}</strong>
     - {{ .Description }}
</p>
{{ end }}
`))
