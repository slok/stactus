package atlassianstatuspage_test

import (
	"context"
	"testing"
	"time"

	"github.com/slok/stactus/internal/model"
	"github.com/slok/stactus/internal/storage/atlassianstatuspage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	bigTestIncidentsJSON  = `{"page": {"id": "kctbh9vrtdwd","name": "GitHub","url": "https://www.githubstatus.com","time_zone": "Etc/UTC","updated_at": "2024-09-17T08:08:17.044Z"},"incidents": [{"id": "69sb0f8lydp4","name": "Incident with Pages and Actions","status": "resolved","created_at": "2024-09-16T21:31:02.719Z","updated_at": "2024-09-16T22:08:39.349Z","monitoring_at": null,"resolved_at": "2024-09-16T22:08:39.333Z","impact": "major","shortlink": "https://stspg.io/mc6sr407p78b","started_at": "2024-09-16T21:31:02.713Z","page_id": "kctbh9vrtdwd","incident_updates": [{"id": "gv96mtfg46fj","status": "resolved","body": "This incident has been resolved.","incident_id": "69sb0f8lydp4","created_at": "2024-09-16T22:08:39.333Z","updated_at": "2024-09-16T22:08:39.333Z","display_at": "2024-09-16T22:08:39.333Z","affected_components": [{"code": "vg70hn9s2tyj","name": "Pages","old_status": "operational","new_status": "operational"},{"code": "br0l2tvcx85d","name": "Actions","old_status": "partial_outage","new_status": "operational"}],"deliver_notifications": true,"custom_tweet": null,"tweet_id": null},{"id": "mfrjlr89z2dk","status": "investigating","body": "Actions is experiencing degraded performance. We are continuing to investigate.","incident_id": "69sb0f8lydp4","created_at": "2024-09-16T21:55:28.472Z","updated_at": "2024-09-16T21:55:28.472Z","display_at": "2024-09-16T21:55:28.472Z","affected_components": [{"code": "vg70hn9s2tyj","name": "Pages","old_status": "operational","new_status": "operational"},{"code": "br0l2tvcx85d","name": "Actions","old_status": "major_outage","new_status": "partial_outage"}],"deliver_notifications": true,"custom_tweet": null,"tweet_id": null},{"id": "h3jcv314m9fg","status": "investigating","body": "The team is investigating issues with some Actions jobs being queued for a long time and a percentage of jobs failing. A mitigation has been applied and jobs are starting to recover.","incident_id": "69sb0f8lydp4","created_at": "2024-09-16T21:53:48.129Z","updated_at": "2024-09-16T21:53:48.129Z","display_at": "2024-09-16T21:53:48.129Z","affected_components": [{"code": "vg70hn9s2tyj","name": "Pages","old_status": "operational","new_status": "operational"},{"code": "br0l2tvcx85d","name": "Actions","old_status": "major_outage","new_status": "major_outage"}],"deliver_notifications": true,"custom_tweet": null,"tweet_id": null},{"id": "x6m13rp5rqzz","status": "investigating","body": "Pages is operating normally.","incident_id": "69sb0f8lydp4","created_at": "2024-09-16T21:52:44.092Z","updated_at": "2024-09-16T21:52:44.092Z","display_at": "2024-09-16T21:52:44.092Z","affected_components": [{"code": "vg70hn9s2tyj","name": "Pages","old_status": "partial_outage","new_status": "operational"},{"code": "br0l2tvcx85d","name": "Actions","old_status": "major_outage","new_status": "major_outage"}],"deliver_notifications": true,"custom_tweet": null,"tweet_id": null},{"id": "r8yz2fpy4ss8","status": "investigating","body": "Actions is experiencing degraded availability. We are continuing to investigate.","incident_id": "69sb0f8lydp4","created_at": "2024-09-16T21:37:26.995Z","updated_at": "2024-09-16T21:37:26.995Z","display_at": "2024-09-16T21:37:26.995Z","affected_components": [{"code": "vg70hn9s2tyj","name": "Pages","old_status": "partial_outage","new_status": "partial_outage"},{"code": "br0l2tvcx85d","name": "Actions","old_status": "partial_outage","new_status": "major_outage"}],"deliver_notifications": true,"custom_tweet": null,"tweet_id": null},{"id": "9lpqk4965mzb","status": "investigating","body": "We are investigating reports of degraded performance for Actions and Pages","incident_id": "69sb0f8lydp4","created_at": "2024-09-16T21:31:02.798Z","updated_at": "2024-09-16T21:31:02.798Z","display_at": "2024-09-16T21:31:02.798Z","affected_components": [{"code": "vg70hn9s2tyj","name": "Pages","old_status": "operational","new_status": "partial_outage"},{"code": "br0l2tvcx85d","name": "Actions","old_status": "operational","new_status": "partial_outage"}],"deliver_notifications": true,"custom_tweet": null,"tweet_id": null}],"components": [{"id": "br0l2tvcx85d","name": "Actions","status": "operational","created_at": "2019-11-13T18:02:19.432Z","updated_at": "2024-09-16T22:08:39.305Z","position": 7,"description": "Workflows, Compute and Orchestration for GitHub Actions","showcase": false,"start_date": null,"group_id": null,"page_id": "kctbh9vrtdwd","group": false,"only_show_if_degraded": false},{"id": "vg70hn9s2tyj","name": "Pages","status": "operational","created_at": "2017-01-31T20:04:33.923Z","updated_at": "2024-09-16T21:52:44.065Z","position": 9,"description": "Frontend application and API servers for Pages builds","showcase": false,"start_date": null,"group_id": null,"page_id": "kctbh9vrtdwd","group": false,"only_show_if_degraded": false}],"reminder_intervals": null},{"id": "r3x7x31k7nn1","name": "Disruption with Git SSH","status": "resolved","created_at": "2024-09-16T13:29:47.163Z","updated_at": "2024-09-16T14:28:03.830Z","monitoring_at": null,"resolved_at": "2024-09-16T14:28:03.812Z","impact": "minor","shortlink": "https://stspg.io/36jl4xpkpcvz","started_at": "2024-09-16T13:29:47.155Z","page_id": "kctbh9vrtdwd","incident_updates": [{"id": "hq18g8q5kw28","status": "resolved","body": "This incident has been resolved.","incident_id": "r3x7x31k7nn1","created_at": "2024-09-16T14:28:03.812Z","updated_at": "2024-09-16T14:28:03.812Z","display_at": "2024-09-16T14:28:03.812Z","affected_components": null,"deliver_notifications": true,"custom_tweet": null,"tweet_id": null},{"id": "pv6g9j3y1sn7","status": "investigating","body": "We are no longer seeing dropped Git SSH connections and believe we have mitigated the incident.  We are continuing to monitor and investigate to prevent reoccurrence.","incident_id": "r3x7x31k7nn1","created_at": "2024-09-16T14:27:52.696Z","updated_at": "2024-09-16T14:27:52.696Z","display_at": "2024-09-16T14:27:52.696Z","affected_components": null,"deliver_notifications": true,"custom_tweet": null,"tweet_id": null},{"id": "xngt1d0zrgmc","status": "investigating","body": "We have taken suspected hosts out of rotation and have not seen any impact in the last 20 minutes.  We are continuing to monitor to ensure the problem is resolved and are investigating the cause.","incident_id": "r3x7x31k7nn1","created_at": "2024-09-16T14:11:07.786Z","updated_at": "2024-09-16T14:11:07.786Z","display_at": "2024-09-16T14:11:07.786Z","affected_components": null,"deliver_notifications": true,"custom_tweet": null,"tweet_id": null},{"id": "6rswdngzbjfg","status": "investigating","body": "We are seeing up to 2% of Git SSH connections failing.<br /><br />We have taken suspected problematic hosts out of rotation and are monitoring for recovery and continuing to investigate.","incident_id": "r3x7x31k7nn1","created_at": "2024-09-16T13:38:25.023Z","updated_at": "2024-09-16T13:38:25.023Z","display_at": "2024-09-16T13:38:25.023Z","affected_components": null,"deliver_notifications": true,"custom_tweet": null,"tweet_id": null},{"id": "twh4f08zp4x5","status": "investigating","body": "We are investigating failed connections for Git SSH.  Customers may be experiencing failed SSH connections both in CI and interactively.  Retrying the connection may be successful.  Git HTTP connections appear to be unaffected.","incident_id": "r3x7x31k7nn1","created_at": "2024-09-16T13:30:07.571Z","updated_at": "2024-09-16T13:30:07.571Z","display_at": "2024-09-16T13:30:07.571Z","affected_components": null,"deliver_notifications": true,"custom_tweet": null,"tweet_id": null},{"id": "58gfcp5sk35t","status": "investigating","body": "We are currently investigating this issue.","incident_id": "r3x7x31k7nn1","created_at": "2024-09-16T13:29:47.204Z","updated_at": "2024-09-16T13:29:47.204Z","display_at": "2024-09-16T13:29:47.204Z","affected_components": null,"deliver_notifications": true,"custom_tweet": null,"tweet_id": null}],"components": [],"reminder_intervals": null}]}`
	bigTestComponentsJSON = `{"page":{"id":"kctbh9vrtdwd","name":"GitHub","url":"https://www.githubstatus.com","time_zone":"Etc/UTC","updated_at":"2024-09-17T08:08:17.044Z"},"components":[{"id":"8l4ygp009s5s","name":"Git Operations","status":"operational","created_at":"2017-01-31T20:05:05.370Z","updated_at":"2024-08-27T23:26:32.478Z","position":1,"description":"Performance of git clones, pulls, pushes, and associated operations","showcase":false,"start_date":null,"group_id":null,"page_id":"kctbh9vrtdwd","group":false,"only_show_if_degraded":false},{"id":"4230lsnqdsld","name":"Webhooks","status":"operational","created_at":"2019-11-13T18:00:24.256Z","updated_at":"2024-09-13T07:13:24.599Z","position":2,"description":"Real time HTTP callbacks of user-generated and system events","showcase":false,"start_date":null,"group_id":null,"page_id":"kctbh9vrtdwd","group":false,"only_show_if_degraded":false}]}`
)

func TestNewJSONStatusPageRepository(t *testing.T) {
	tests := map[string]struct {
		componentsJSON string
		incidentsJSON  string
		expSettings    model.StatusPageSettings
		expSystems     []model.System
		expIncidents   []model.IncidentReport
		expErr         bool
	}{

		"Components, settings and incidents should be loaded correctly.": {
			componentsJSON: bigTestComponentsJSON,
			incidentsJSON:  bigTestIncidentsJSON,
			expSettings:    model.StatusPageSettings{Name: "GitHub", URL: "https://www.githubstatus.com"},
			expSystems: []model.System{
				{ID: "8l4ygp009s5s", Name: "Git Operations", Description: "Performance of git clones, pulls, pushes, and associated operations"},
				{ID: "4230lsnqdsld", Name: "Webhooks", Description: "Real time HTTP callbacks of user-generated and system events"},
			},
			expIncidents: []model.IncidentReport{
				{
					ID:        "69sb0f8lydp4",
					Name:      "Incident with Pages and Actions",
					SystemIDs: []string{"br0l2tvcx85d", "vg70hn9s2tyj"},
					Impact:    model.IncidentImpactMajor,
					Start:     time.Date(2024, time.September, 16, 21, 31, 2, 798000000, time.UTC),
					End:       time.Date(2024, time.September, 16, 22, 8, 39, 333000000, time.UTC),
					Duration:  2256535000000 * time.Nanosecond,
					Timeline: []model.IncidentReportEvent{
						{
							Description: "This incident has been resolved.",
							TS:          time.Date(2024, time.September, 16, 22, 8, 39, 333000000, time.UTC),
							Kind:        model.IncidentUpdateKindResolved,
						},
						{
							Description: "Actions is experiencing degraded performance. We are continuing to investigate.",
							TS:          time.Date(2024, time.September, 16, 21, 55, 28, 472000000, time.UTC),
							Kind:        model.IncidentUpdateKindInvestigating,
						},
						{
							Description: "The team is investigating issues with some Actions jobs being queued for a long time and a percentage of jobs failing. A mitigation has been applied and jobs are starting to recover.",
							TS:          time.Date(2024, time.September, 16, 21, 53, 48, 129000000, time.UTC),
							Kind:        model.IncidentUpdateKindInvestigating,
						},
						{
							Description: "Pages is operating normally.",
							TS:          time.Date(2024, time.September, 16, 21, 52, 44, 92000000, time.UTC),
							Kind:        model.IncidentUpdateKindInvestigating,
						},
						{
							Description: "Actions is experiencing degraded availability. We are continuing to investigate.",
							TS:          time.Date(2024, time.September, 16, 21, 37, 26, 995000000, time.UTC),
							Kind:        model.IncidentUpdateKindInvestigating,
						},
						{
							Description: "We are investigating reports of degraded performance for Actions and Pages",
							TS:          time.Date(2024, time.September, 16, 21, 31, 2, 798000000, time.UTC),
							Kind:        model.IncidentUpdateKindInvestigating,
						},
					},
				},

				{
					ID:        "r3x7x31k7nn1",
					Name:      "Disruption with Git SSH",
					SystemIDs: []string{},
					Impact:    model.IncidentImpactMinor,
					Start:     time.Date(2024, time.September, 16, 13, 29, 47, 204000000, time.UTC),
					End:       time.Date(2024, time.September, 16, 14, 28, 3, 812000000, time.UTC),
					Duration:  3496608000000 * time.Nanosecond,
					Timeline: []model.IncidentReportEvent{
						{
							Description: "This incident has been resolved.",
							TS:          time.Date(2024, time.September, 16, 14, 28, 3, 812000000, time.UTC),
							Kind:        model.IncidentUpdateKindResolved,
						},
						{
							Description: "We are no longer seeing dropped Git SSH connections and believe we have mitigated the incident.  We are continuing to monitor and investigate to prevent reoccurrence.",
							TS:          time.Date(2024, time.September, 16, 14, 27, 52, 696000000, time.UTC),
							Kind:        model.IncidentUpdateKindInvestigating,
						},
						{
							Description: "We have taken suspected hosts out of rotation and have not seen any impact in the last 20 minutes.  We are continuing to monitor to ensure the problem is resolved and are investigating the cause.", TS: time.Date(2024, time.September, 16, 14, 11, 7, 786000000, time.UTC),
							Kind: model.IncidentUpdateKindInvestigating,
						},
						{
							Description: "We are seeing up to 2% of Git SSH connections failing.<br /><br />We have taken suspected problematic hosts out of rotation and are monitoring for recovery and continuing to investigate.",
							TS:          time.Date(2024, time.September, 16, 13, 38, 25, 23000000, time.UTC),
							Kind:        model.IncidentUpdateKindInvestigating,
						},
						{
							Description: "We are investigating failed connections for Git SSH.  Customers may be experiencing failed SSH connections both in CI and interactively.  Retrying the connection may be successful.  Git HTTP connections appear to be unaffected.",
							TS:          time.Date(2024, time.September, 16, 13, 30, 7, 571000000, time.UTC),
							Kind:        model.IncidentUpdateKindInvestigating,
						},
						{
							Description: "We are currently investigating this issue.",
							TS:          time.Date(2024, time.September, 16, 13, 29, 47, 204000000, time.UTC),
							Kind:        model.IncidentUpdateKindInvestigating,
						},
					},
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			repo, err := atlassianstatuspage.NewJSONStatusPageRepository(test.componentsJSON, test.incidentsJSON)
			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				gotSettings, err := repo.GetStatusPageSettings(context.TODO())
				require.NoError(err)
				assert.Equal(test.expSettings, *gotSettings)

				gotSystems, err := repo.ListAllSystems(context.TODO())
				require.NoError(err)
				assert.Equal(test.expSystems, gotSystems)

				gotIncidents, err := repo.ListAllIncidentReports(context.TODO())
				require.NoError(err)
				assert.Equal(test.expIncidents, gotIncidents)
			}
		})
	}
}
