package feed_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/slok/stactus/internal/model"
	"github.com/slok/stactus/internal/storage/feed"
	utilfs "github.com/slok/stactus/internal/util/fs"
)

func TestRepositoryCreateHistoryFeed(t *testing.T) {
	t0, _ := time.Parse(time.RFC3339, "1912-06-23T01:02:03Z")

	tests := map[string]struct {
		ui       func() model.UI
		expFeeds map[string]string
		expErr   bool
	}{
		"Correct data should render correctly the metrics": {
			ui: func() model.UI {
				return model.UI{
					Settings: model.StatusPageSettings{
						Name: "Test",
						URL:  "https://status.slok.dev",
					},
					History: []*model.IncidentReport{
						{ID: "ir3", Name: "IR 3", SystemIDs: []string{"s3"}, Start: t0.Add(100 * time.Minute), Timeline: []model.IncidentReportEvent{
							{TS: t0.Add(110 * time.Minute), Description: "d33", Kind: model.IncidentUpdateKindUpdate},
							{TS: t0.Add(109 * time.Minute), Description: "d32", Kind: model.IncidentUpdateKindUpdate},
							{TS: t0.Add(100 * time.Minute), Description: "[d31](https://slok.dev)", Kind: model.IncidentUpdateKindInvestigating},
						}},
						{ID: "ir2", Name: "IR 2", SystemIDs: []string{"s2"}, Start: t0.Add(200 * time.Minute), End: t0.Add(220 * time.Minute), Duration: 20 * time.Minute, Timeline: []model.IncidentReportEvent{
							{TS: t0.Add(220 * time.Minute), Description: "d24", Kind: model.IncidentUpdateKindResolved},
							{TS: t0.Add(215 * time.Minute), Description: "d23", Kind: model.IncidentUpdateKindUpdate},
							{TS: t0.Add(209 * time.Minute), Description: "d22", Kind: model.IncidentUpdateKindUpdate},
							{TS: t0.Add(200 * time.Minute), Description: "**d21**", Kind: model.IncidentUpdateKindInvestigating},
						}},
						{ID: "ir1", Name: "IR 1", SystemIDs: []string{"s1"}, Start: t0.Add(300 * time.Minute), End: t0.Add(320 * time.Minute), Duration: 20 * time.Minute, Timeline: []model.IncidentReportEvent{
							{TS: t0.Add(320 * time.Minute), Description: "d12", Kind: model.IncidentUpdateKindResolved},
							{TS: t0.Add(300 * time.Minute), Description: "d11", Kind: model.IncidentUpdateKindInvestigating},
						}},
					},
				}
			},
			expFeeds: map[string]string{
				"test/history-feed.atom": `
<?xml version="1.0" encoding="UTF-8"?><feed xmlns="http://www.w3.org/2005/Atom">
  <title>Test - Incident history</title>
  <id>https://status.slok.dev</id>
  <updated>1912-06-23T01:02:03Z</updated>
  <subtitle>Test status page</subtitle>
  <link href="https://status.slok.dev" rel="alternate"></link>
  <author>
    <name>Test</name>
  </author>
  <entry>
    <title>IR 3</title>
    <updated>1912-06-23T02:52:03Z</updated>
    <id>https://status.slok.dev/ir/ir3</id>
    <content type="html">&#xA;&#xA;&lt;p&gt;&#xA;    &lt;small&gt;1912-06-23T02:52:03Z&lt;/small&gt;&#xA;    &lt;br /&gt;&#xA;    &lt;strong&gt;update&lt;/strong&gt;&#xA;     - &lt;p&gt;d33&lt;/p&gt;&#xA;&#xA;&lt;/p&gt;&#xA;&#xA;&lt;p&gt;&#xA;    &lt;small&gt;1912-06-23T02:51:03Z&lt;/small&gt;&#xA;    &lt;br /&gt;&#xA;    &lt;strong&gt;update&lt;/strong&gt;&#xA;     - &lt;p&gt;d32&lt;/p&gt;&#xA;&#xA;&lt;/p&gt;&#xA;&#xA;&lt;p&gt;&#xA;    &lt;small&gt;1912-06-23T02:42:03Z&lt;/small&gt;&#xA;    &lt;br /&gt;&#xA;    &lt;strong&gt;investigating&lt;/strong&gt;&#xA;     - &lt;p&gt;&lt;a href=&#34;https://slok.dev&#34;&gt;d31&lt;/a&gt;&lt;/p&gt;&#xA;&#xA;&lt;/p&gt;&#xA;&#xA;</content>
    <link href="https://status.slok.dev/ir/ir3" rel="alternate" type="text/html"></link>
  </entry>
  <entry>
    <title>IR 2</title>
    <updated>1912-06-23T04:42:03Z</updated>
    <id>https://status.slok.dev/ir/ir2</id>
    <content type="html">&#xA;&#xA;&lt;p&gt;&#xA;    &lt;small&gt;1912-06-23T04:42:03Z&lt;/small&gt;&#xA;    &lt;br /&gt;&#xA;    &lt;strong&gt;resolved&lt;/strong&gt;&#xA;     - &lt;p&gt;d24&lt;/p&gt;&#xA;&#xA;&lt;/p&gt;&#xA;&#xA;&lt;p&gt;&#xA;    &lt;small&gt;1912-06-23T04:37:03Z&lt;/small&gt;&#xA;    &lt;br /&gt;&#xA;    &lt;strong&gt;update&lt;/strong&gt;&#xA;     - &lt;p&gt;d23&lt;/p&gt;&#xA;&#xA;&lt;/p&gt;&#xA;&#xA;&lt;p&gt;&#xA;    &lt;small&gt;1912-06-23T04:31:03Z&lt;/small&gt;&#xA;    &lt;br /&gt;&#xA;    &lt;strong&gt;update&lt;/strong&gt;&#xA;     - &lt;p&gt;d22&lt;/p&gt;&#xA;&#xA;&lt;/p&gt;&#xA;&#xA;&lt;p&gt;&#xA;    &lt;small&gt;1912-06-23T04:22:03Z&lt;/small&gt;&#xA;    &lt;br /&gt;&#xA;    &lt;strong&gt;investigating&lt;/strong&gt;&#xA;     - &lt;p&gt;&lt;strong&gt;d21&lt;/strong&gt;&lt;/p&gt;&#xA;&#xA;&lt;/p&gt;&#xA;&#xA;</content>
    <link href="https://status.slok.dev/ir/ir2" rel="alternate" type="text/html"></link>
  </entry>
  <entry>
    <title>IR 1</title>
    <updated>1912-06-23T06:22:03Z</updated>
    <id>https://status.slok.dev/ir/ir1</id>
    <content type="html">&#xA;&#xA;&lt;p&gt;&#xA;    &lt;small&gt;1912-06-23T06:22:03Z&lt;/small&gt;&#xA;    &lt;br /&gt;&#xA;    &lt;strong&gt;resolved&lt;/strong&gt;&#xA;     - &lt;p&gt;d12&lt;/p&gt;&#xA;&#xA;&lt;/p&gt;&#xA;&#xA;&lt;p&gt;&#xA;    &lt;small&gt;1912-06-23T06:02:03Z&lt;/small&gt;&#xA;    &lt;br /&gt;&#xA;    &lt;strong&gt;investigating&lt;/strong&gt;&#xA;     - &lt;p&gt;d11&lt;/p&gt;&#xA;&#xA;&lt;/p&gt;&#xA;&#xA;</content>
    <link href="https://status.slok.dev/ir/ir1" rel="alternate" type="text/html"></link>
  </entry>
</feed>
`},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require := assert.New(t)
			assert := assert.New(t)

			fsm := utilfs.NewTestFileManager()
			repo, err := feed.NewFSRepository(feed.RepositoryConfig{
				FileManager:         fsm,
				AtomHistoryFilePath: "test/history-feed.atom",
				TimeNow:             func() time.Time { return t0 },
			})
			require.NoError(err)

			err = repo.CreateHistoryFeed(context.TODO(), test.ui())
			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				for k, v := range test.expFeeds {
					fsm.AssertEqual(t, k, v)
				}
			}
		})
	}
}
