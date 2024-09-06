package system

import (
	"context"
	"fmt"

	"github.com/slok/stactus/internal/internalerrors"
	"github.com/slok/stactus/internal/model"
)

type ListAllSystemsReq struct{}

func (r *ListAllSystemsReq) validate() error {
	return nil
}

type SystemAndOngoingIncident struct {
	System model.System
	IR     *model.IncidentReport
}

type ListAllSystemsResp struct {
	Message string
	Systems []SystemAndOngoingIncident
}

func (s Service) ListAllSystems(ctx context.Context, req ListAllSystemsReq) (ListAllSystemsResp, error) {
	// Validate inputs.
	err := req.validate()
	if err != nil {
		return ListAllSystemsResp{Message: err.Error()}, internalerrors.ErrNotValid
	}

	// Get all systems.
	systems, err := s.sysGetter.ListAllSystems(ctx)
	if err != nil {
		return ListAllSystemsResp{}, fmt.Errorf("could not list systems: %w", err)
	}

	// TODO(slok): Get each of the systems ongoing incidents.

	sysAndIRs := []SystemAndOngoingIncident{}
	for _, sys := range systems {
		sysAndIRs = append(sysAndIRs, SystemAndOngoingIncident{
			System: sys,
		})
	}

	return ListAllSystemsResp{
		Systems: sysAndIRs,
	}, nil
}
