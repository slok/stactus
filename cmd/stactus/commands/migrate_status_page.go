package commands

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"gopkg.in/yaml.v3"

	"github.com/slok/stactus/internal/log"
	"github.com/slok/stactus/internal/model"
	"github.com/slok/stactus/internal/storage/atlassianstatuspage"
	"github.com/slok/stactus/internal/storage/memory"
	utilfs "github.com/slok/stactus/internal/util/fs"
	apiv1 "github.com/slok/stactus/pkg/api/v1"
)

type MigrateStatusPageCommand struct {
	cmd        *kingpin.CmdClause
	rootConfig *RootCommand

	statusPageURL string
	outPath       string
}

// NewMigrateStatusPageCommand returns a generator with the github status page theme.
func NewMigrateStatusPageCommand(rootConfig *RootCommand, app MigrateCommand) *MigrateStatusPageCommand {
	cmd := app.Cmd.Command("status-page", "Migrate Atlassian status page to stactus files.")
	c := &MigrateStatusPageCommand{
		cmd:        cmd,
		rootConfig: rootConfig,
	}

	cmd.Flag("status-page-url", "The URL to Atlassian status page API (E.g: https://www.githubstatus.com).").Required().Short('u').StringVar(&c.statusPageURL)
	cmd.Flag("out", "The directory where all the generated files will be written.").Required().Short('o').StringVar(&c.outPath)

	return c
}

func (c *MigrateStatusPageCommand) Name() string { return c.cmd.FullCommand() }
func (c *MigrateStatusPageCommand) Run(ctx context.Context) (err error) {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(err)

	logger := c.rootConfig.Logger

	logger.Infof("Retrieving data...")
	repo, err := atlassianstatuspage.NewStatusPageRepository(c.statusPageURL)
	if err != nil {
		return fmt.Errorf("could not create atlassian status page repository: %w", err)
	}

	logger.Infof("Migrating %s...", c.statusPageURL)

	return migrateStatusPageRepository(ctx, logger, repo, c.outPath)
}

func migrateStatusPageRepository(ctx context.Context, logger log.Logger, repo *memory.Repository, outPath string) error {
	var fileManager utilfs.FileManager = utilfs.StdFileManager

	// Write stactus file.
	{
		settings, err := repo.GetStatusPageSettings(ctx)
		if err != nil {
			return fmt.Errorf("could not get settings: %w", err)
		}

		apiStactus := apiv1.StactusV1{
			Version: apiv1.StactusVersionV1,
			Name:    settings.Name,
			URL:     settings.URL,
		}

		systems, err := repo.ListAllSystems(ctx)
		if err != nil {
			return fmt.Errorf("could not list systems: %w", err)
		}
		for _, s := range systems {
			apiStactus.Systems = append(apiStactus.Systems, apiv1.StactusV1System{
				ID:          s.ID,
				Name:        s.Name,
				Description: s.Description,
			})
		}

		data, err := yaml.Marshal(apiStactus)
		if err != nil {
			return fmt.Errorf("could not marshal to yaml stactus file: %w", err)
		}

		// Write.
		fpath := filepath.Join(outPath, "stactus.yaml")
		err = fileManager.WriteFile(ctx, fpath, data)
		if err != nil {
			return fmt.Errorf("could not write %q: %w", fpath, err)
		}
	}

	// Write IRs.
	{

		irs, err := repo.ListAllIncidentReports(ctx)
		if err != nil {
			return fmt.Errorf("could not list irs: %w", err)
		}

		// Map each IR.
		for _, ir := range irs {
			// Map to API.
			timeline := []apiv1.IncidentV1TimelineEvent{}
			for _, event := range ir.Timeline {
				investigating := false
				resolved := false
				switch event.Kind {
				case model.IncidentUpdateKindInvestigating:
					investigating = true
				case model.IncidentUpdateKindResolved:
					resolved = true
				}

				timeline = append(timeline, apiv1.IncidentV1TimelineEvent{
					TS:            event.TS.UTC().Format(time.DateTime),
					Description:   event.Description,
					Investigating: investigating,
					Resolved:      resolved,
				})
			}

			// Revert timeline so on the yaml the ones in the last position are the latest events.
			slices.Reverse(timeline)

			apiIR := apiv1.IncidentV1{
				Version:  apiv1.IncidentVersionV1,
				ID:       ir.ID,
				Name:     ir.Name,
				Systems:  ir.SystemIDs,
				Impact:   strings.ToLower(string(ir.Impact)),
				Timeline: timeline,
			}

			data, err := yaml.Marshal(apiIR)
			if err != nil {
				return fmt.Errorf("could not marshal to yaml %q incident: %w", ir.ID, err)
			}

			// Write.
			fpath := filepath.Join(outPath, "incidents", ir.ID+".yaml")
			err = fileManager.WriteFile(ctx, fpath, data)
			if err != nil {
				return fmt.Errorf("could not write %q: %w", fpath, err)
			}

			logger.Debugf("Written %q incident", ir.ID)
		}
	}

	return nil
}
