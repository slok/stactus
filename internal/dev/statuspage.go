package dev

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/slok/stactus/internal/model"
	storagememory "github.com/slok/stactus/internal/storage/memory"
)

func NewGithubRepository() (*storagememory.Repository, error) {
	return newStatusPageURLRepository("https://www.githubstatus.com/")
}

func NewDigitalOceanRepository() (*storagememory.Repository, error) {
	return newStatusPageURLRepository("https://status.digitalocean.com/")
}

func NewTwilioRepository() (*storagememory.Repository, error) {
	return newStatusPageURLRepository("https://status.twilio.com/")
}

func NewHashicorpRepository() (*storagememory.Repository, error) {
	return newStatusPageURLRepository("https://status.hashicorp.com/")
}

func NewDiscordRepository() (*storagememory.Repository, error) {
	return newStatusPageURLRepository("https://discordstatus.com/")
}

func NewCloudflareRepository() (*storagememory.Repository, error) {
	return newStatusPageURLRepository("https://www.cloudflarestatus.com/")
}

// Same as newStatusPageRepository but instead it will download the data form the API.
func newStatusPageURLRepository(url string) (*storagememory.Repository, error) {
	// Sanitize URL
	url = strings.TrimSpace(url)
	url = strings.TrimSuffix(url, "/")
	url = url + "/"

	// Download data.
	resp, err := http.Get(url + "api/v2/components.json")
	if err != nil {
		return nil, fmt.Errorf("could not download components form the API: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed, HTTP status code: %d ", resp.StatusCode)
	}
	defer resp.Body.Close()
	d, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read components data from resp: %w", err)
	}
	componentsRawJSON := string(d)

	resp, err = http.Get(url + "/api/v2/incidents.json")
	if err != nil {
		return nil, fmt.Errorf("could not download incidents form the API: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed, HTTP status code: %d ", resp.StatusCode)
	}
	defer resp.Body.Close()
	d, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read incidents data from resp: %w", err)
	}
	incidentsRawJSON := string(d)

	return newStatusPageRepository(componentsRawJSON, incidentsRawJSON)
}

// newStatusPageRepository knows how to map Atlassian status page API components/incidents into stactus model, helpful for development.
func newStatusPageRepository(componentsRawJSON string, incidentsRawJSON string) (*storagememory.Repository, error) {
	// Map systems (components).
	jsonComponents := struct {
		Components []struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
			GroupID     string `json:"group_id"`
			IsGroup     bool   `json:"group"`
		} `json:"components"`
	}{}
	err := json.Unmarshal([]byte(componentsRawJSON), &jsonComponents)
	if err != nil {
		return nil, fmt.Errorf("could not parse JSON components: %w", err)
	}

	// Map group ID to names.
	groups := map[string]string{}
	for _, g := range jsonComponents.Components {
		if !g.IsGroup {
			continue
		}
		groups[g.ID] = g.Name
	}

	systems := []model.System{}
	for _, comp := range jsonComponents.Components {
		// Ignore groups.
		if comp.IsGroup {
			continue
		}

		// If part of a group add the prefix on the name.
		name := comp.Name
		if comp.GroupID != "" {
			groupName, ok := groups[comp.GroupID]
			if !ok {
				return nil, fmt.Errorf("invalid group: %q", comp.GroupID)
			}
			name = fmt.Sprintf("%s / %s", groupName, name)
		}

		systems = append(systems, model.System{
			ID:          comp.ID,
			Name:        name,
			Description: comp.Description,
		})
	}

	// Map IRs.
	jsonIncidents := struct {
		Incidents []struct {
			ID         string    `json:"id"`
			Name       string    `json:"name"`
			Status     string    `json:"status"`
			Impact     string    `json:"impact"`
			CreatedAt  time.Time `json:"created_at"`
			ResolvedAt time.Time `json:"resolved_at"`
			Components []struct {
				ID string `json:"id"`
			} `json:"components"`
			IncidentUpdates []struct {
				Status    string    `json:"status"`
				Body      string    `json:"body"`
				CreatedAt time.Time `json:"created_at"`
			} `json:"incident_updates"`
		} `json:"incidents"`
	}{}
	err = json.Unmarshal([]byte(incidentsRawJSON), &jsonIncidents)
	if err != nil {
		return nil, fmt.Errorf("could not parse JSON incidents: %w", err)
	}

	irs := []model.IncidentReport{}
	for _, ir := range jsonIncidents.Incidents {
		componets := []string{}
		for _, c := range ir.Components {
			componets = append(componets, c.ID)
		}

		timeline := []model.IncidentReportEvent{}
		for _, update := range ir.IncidentUpdates {
			timeline = append(timeline, model.IncidentReportEvent{
				TS:          update.CreatedAt,
				Description: update.Body,
				Kind:        mapStatusPageUpdateStatusToModel(update.Status),
			})
		}

		// Set end date (for us resolved) if resolved.
		var end time.Time
		if ir.Status == "resolved" {
			end = ir.ResolvedAt
		}

		irs = append(irs, model.IncidentReport{
			ID:        ir.ID,
			Name:      ir.Name,
			Start:     ir.CreatedAt,
			End:       end,
			Impact:    mapStatusPageImpactToModel(ir.Impact),
			SystemIDs: componets,
			Timeline:  timeline,
		})
	}

	memRepo := storagememory.NewRepository(systems, irs)

	return &memRepo, nil
}

func mapStatusPageUpdateStatusToModel(status string) model.IncidentUpdateKind {
	switch strings.ToLower(status) {
	case "investigating":
		return model.IncidentUpdateKindInvestigating
	case "resolved":
		return model.IncidentUpdateKindResolved
	default:
		return model.IncidentUpdateKindUpdate
	}
}

func mapStatusPageImpactToModel(impact string) model.IncidentImpact {
	switch strings.ToLower(impact) {
	case "minor":
		return model.IncidentImpactMinor
	case "critical":
		return model.IncidentImpactCritical
	case "major":
		return model.IncidentImpactMajor
	default:
		return model.IncidentImpactNone
	}
}
