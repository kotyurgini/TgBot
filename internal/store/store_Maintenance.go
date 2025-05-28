package store

import (
	"context"
	"fmt"
)

func (s *Store) MaintenanceStatus() bool {
	return s.appSetting.Maintenance
}

func (s *Store) SetMaintenance(ctx context.Context, maintenance bool) error {
	prevValue := s.appSetting.Maintenance
	s.appSetting.Maintenance = maintenance
	if _, err := s.driver.AppSettingUpdate(ctx, s.appSetting); err != nil {
		s.appSetting.Maintenance = prevValue
		return fmt.Errorf("failed to set maintenance mode: %w", err)
	}
	return nil
}
