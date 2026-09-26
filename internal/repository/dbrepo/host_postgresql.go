package dbrepo

import (
	"context"
	"log"
	"time"

	"github.com/GangaRamPrasad2004/sentinal/internal/models"
	"github.com/GangaRamPrasad2004/sentinal/internal/repository/dbrepo/sqlc"
)

// InsertHost inserts a host into the database using sqlc type-safe query
func (m *postgresDBRepo) InsertHost(h models.Host) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	newID, err := m.q.InsertHost(ctx, sqlc.InsertHostParams{
		HostName:      h.HostName,
		CanonicalName: h.CanonicalName,
		Url:           h.URL,
		Ip:            h.IP,
		Ipv6:          h.IPV6,
		Location:      h.Location,
		Os:            h.OS,
		Active:        int32(h.Active),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	})

	if err != nil {
		log.Println(err)
		return int(newID), err
	}

	// add host services and set to inactive
	serviceIDs, err := m.q.ListServiceIDs(ctx)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	for _, svcID := range serviceIDs {
		err = m.q.InsertHostServiceDefault(ctx, sqlc.InsertHostServiceDefaultParams{
			HostID:    newID,
			ServiceID: svcID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
		if err != nil {
			return int(newID), err
		}
	}

	return int(newID), nil
}

// GetHostByID gets a host by id and returns models.Host using sqlc type-safe query
func (m *postgresDBRepo) GetHostByID(id int) (models.Host, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	hostRow, err := m.q.GetHostByID(ctx, int32(id))
	if err != nil {
		return models.Host{}, err
	}

	h := models.Host{
		ID:            int(hostRow.ID),
		HostName:      hostRow.HostName,
		CanonicalName: hostRow.CanonicalName,
		URL:           hostRow.Url,
		IP:            hostRow.Ip,
		IPV6:          hostRow.Ipv6,
		Location:      hostRow.Location,
		OS:            hostRow.Os,
		Active:        int(hostRow.Active),
		CreatedAt:     hostRow.CreatedAt,
		UpdatedAt:     hostRow.UpdatedAt,
	}

	// get all services for host
	hsRows, err := m.q.GetHostServicesForHost(ctx, int32(h.ID))
	if err != nil {
		log.Println(err)
		return h, err
	}

	var hostServices []models.HostService
	for _, r := range hsRows {
		hs := models.HostService{
			ID:             int(r.ID),
			HostID:         int(r.HostID),
			ServiceID:      int(r.ServiceID),
			Active:         int(r.Active),
			ScheduleNumber: int(r.ScheduleNumber),
			ScheduleUnit:   r.ScheduleUnit,
			LastCheck:      r.LastCheck,
			Status:         r.Status,
			CreatedAt:      r.CreatedAt,
			UpdatedAt:      r.UpdatedAt,
			Service: models.Services{
				ID:          int(r.ServiceIDRef),
				ServiceName: r.ServiceName,
				Active:      int(r.ServiceActive),
				Icon:        r.Icon,
				CreatedAt:   r.ServiceCreatedAt,
				UpdatedAt:   r.ServiceUpdatedAt,
			},
			LastMessage: r.LastMessage,
		}
		hostServices = append(hostServices, hs)
	}

	h.HostServices = hostServices

	return h, nil
}

// UpdateHost updates a host in the database using sqlc type-safe query
func (m *postgresDBRepo) UpdateHost(h models.Host) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.q.UpdateHost(ctx, sqlc.UpdateHostParams{
		HostName:      h.HostName,
		CanonicalName: h.CanonicalName,
		Url:           h.URL,
		Ip:            h.IP,
		Ipv6:          h.IPV6,
		Os:            h.OS,
		Active:        int32(h.Active),
		Location:      h.Location,
		UpdatedAt:     time.Now(),
		ID:            int32(h.ID),
	})

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// GetAllServiceStatusCounts gets count of all service statuses using sqlc
func (m *postgresDBRepo) GetAllServiceStatusCounts() (int, int, int, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	counts, err := m.q.GetAllServiceStatusCounts(ctx)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	return int(counts.Pending), int(counts.Healthy), int(counts.Warning), int(counts.Problem), nil
}

// AllHosts returns a slice of hosts using sqlc type-safe query
func (m *postgresDBRepo) AllHosts() ([]models.Host, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	hostRows, err := m.q.AllHosts(ctx)
	if err != nil {
		return nil, err
	}

	var hosts []models.Host
	for _, hr := range hostRows {
		h := models.Host{
			ID:            int(hr.ID),
			HostName:      hr.HostName,
			CanonicalName: hr.CanonicalName,
			URL:           hr.Url,
			IP:            hr.Ip,
			IPV6:          hr.Ipv6,
			Location:      hr.Location,
			OS:            hr.Os,
			Active:        int(hr.Active),
			CreatedAt:     hr.CreatedAt,
			UpdatedAt:     hr.UpdatedAt,
		}

		serviceRows, err := m.q.GetHostServicesForHost(ctx, hr.ID)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		var hostServices []models.HostService
		for _, sr := range serviceRows {
			hs := models.HostService{
				ID:             int(sr.ID),
				HostID:         int(sr.HostID),
				ServiceID:      int(sr.ServiceID),
				Active:         int(sr.Active),
				ScheduleNumber: int(sr.ScheduleNumber),
				ScheduleUnit:   sr.ScheduleUnit,
				LastCheck:      sr.LastCheck,
				Status:         sr.Status,
				CreatedAt:      sr.CreatedAt,
				UpdatedAt:      sr.UpdatedAt,
				Service: models.Services{
					ID:          int(sr.ServiceIDRef),
					ServiceName: sr.ServiceName,
					Active:      int(sr.ServiceActive),
					Icon:        sr.Icon,
					CreatedAt:   sr.ServiceCreatedAt,
					UpdatedAt:   sr.ServiceUpdatedAt,
				},
				LastMessage: sr.LastMessage,
			}
			hostServices = append(hostServices, hs)
		}
		h.HostServices = hostServices
		hosts = append(hosts, h)
	}

	return hosts, nil
}

// UpdateHostServiceStatus updates the active status of a host service using sqlc
func (m *postgresDBRepo) UpdateHostServiceStatus(hostID, serviceID, active int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.q.UpdateHostServiceStatus(ctx, sqlc.UpdateHostServiceStatusParams{
		Active:    int32(active),
		HostID:    int32(hostID),
		ServiceID: int32(serviceID),
	})
}

// UpdateHostService updates a host service in the database using sqlc
func (m *postgresDBRepo) UpdateHostService(hs models.HostService) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.q.UpdateHostService(ctx, sqlc.UpdateHostServiceParams{
		HostID:         int32(hs.HostID),
		ServiceID:      int32(hs.ServiceID),
		Active:         int32(hs.Active),
		ScheduleNumber: int32(hs.ScheduleNumber),
		ScheduleUnit:   hs.ScheduleUnit,
		LastCheck:      hs.LastCheck,
		Status:         hs.Status,
		UpdatedAt:      hs.UpdatedAt,
		LastMessage:    hs.LastMessage,
		ID:             int32(hs.ID),
	})
}

// GetServicesByStatus returns all active services with a given status using sqlc
func (m *postgresDBRepo) GetServicesByStatus(status string) ([]models.HostService, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.q.GetServicesByStatus(ctx, status)
	if err != nil {
		return nil, err
	}

	var services []models.HostService
	for _, r := range rows {
		h := models.HostService{
			ID:             int(r.ID),
			HostID:         int(r.HostID),
			ServiceID:      int(r.ServiceID),
			Active:         int(r.Active),
			ScheduleNumber: int(r.ScheduleNumber),
			ScheduleUnit:   r.ScheduleUnit,
			LastCheck:      r.LastCheck,
			Status:         r.Status,
			CreatedAt:      r.CreatedAt,
			UpdatedAt:      r.UpdatedAt,
			HostName:       r.HostName,
			Service: models.Services{
				ServiceName: r.ServiceName,
			},
			LastMessage: r.LastMessage,
		}
		services = append(services, h)
	}

	return services, nil
}

// GetHostServiceByID gets a host service by id using sqlc
func (m *postgresDBRepo) GetHostServiceByID(id int) (models.HostService, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	r, err := m.q.GetHostServiceByID(ctx, int32(id))
	if err != nil {
		log.Println(err)
		return models.HostService{}, err
	}

	hs := models.HostService{
		ID:             int(r.ID),
		HostID:         int(r.HostID),
		ServiceID:      int(r.ServiceID),
		Active:         int(r.Active),
		ScheduleNumber: int(r.ScheduleNumber),
		ScheduleUnit:   r.ScheduleUnit,
		LastCheck:      r.LastCheck,
		Status:         r.Status,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
		Service: models.Services{
			ID:          int(r.ServiceIDRef),
			ServiceName: r.ServiceName,
			Active:      int(r.ServiceActive),
			Icon:        r.Icon,
			CreatedAt:   r.ServiceCreatedAt,
			UpdatedAt:   r.ServiceUpdatedAt,
		},
		HostName:    r.HostName,
		LastMessage: r.LastMessage,
	}

	return hs, nil
}

// GetServicesToMonitor gets all host services we want to monitor using sqlc
func (m *postgresDBRepo) GetServicesToMonitor() ([]models.HostService, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.q.GetServicesToMonitor(ctx)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var services []models.HostService
	for _, r := range rows {
		h := models.HostService{
			ID:             int(r.ID),
			HostID:         int(r.HostID),
			ServiceID:      int(r.ServiceID),
			Active:         int(r.Active),
			ScheduleNumber: int(r.ScheduleNumber),
			ScheduleUnit:   r.ScheduleUnit,
			LastCheck:      r.LastCheck,
			Status:         r.Status,
			CreatedAt:      r.CreatedAt,
			UpdatedAt:      r.UpdatedAt,
			Service: models.Services{
				ID:          int(r.ServiceIDRef),
				ServiceName: r.ServiceName,
				Active:      int(r.ServiceActive),
				Icon:        r.Icon,
				CreatedAt:   r.ServiceCreatedAt,
				UpdatedAt:   r.ServiceUpdatedAt,
			},
			HostName:    r.HostName,
			LastMessage: r.LastMessage,
		}
		services = append(services, h)
	}

	return services, nil
}

// GetHostServiceByHostIDServiceID gets a host service by host id and service id using sqlc
func (m *postgresDBRepo) GetHostServiceByHostIDServiceID(hostID, serviceID int) (models.HostService, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	r, err := m.q.GetHostServiceByHostIDServiceID(ctx, sqlc.GetHostServiceByHostIDServiceIDParams{
		HostID:    int32(hostID),
		ServiceID: int32(serviceID),
	})
	if err != nil {
		log.Println(err)
		return models.HostService{}, err
	}

	hs := models.HostService{
		ID:             int(r.ID),
		HostID:         int(r.HostID),
		ServiceID:      int(r.ServiceID),
		Active:         int(r.Active),
		ScheduleNumber: int(r.ScheduleNumber),
		ScheduleUnit:   r.ScheduleUnit,
		LastCheck:      r.LastCheck,
		Status:         r.Status,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
		Service: models.Services{
			ID:          int(r.ServiceIDRef),
			ServiceName: r.ServiceName,
			Active:      int(r.ServiceActive),
			Icon:        r.Icon,
			CreatedAt:   r.ServiceCreatedAt,
			UpdatedAt:   r.ServiceUpdatedAt,
		},
		HostName:    r.HostName,
		LastMessage: r.LastMessage,
	}

	return hs, nil
}

// InsertEvent inserts an event into the database using sqlc
func (m *postgresDBRepo) InsertEvent(e models.Event) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.q.InsertEvent(ctx, sqlc.InsertEventParams{
		HostServiceID: int32(e.HostServiceID),
		EventType:     e.EventType,
		HostID:        int32(e.HostID),
		ServiceName:   e.ServiceName,
		HostName:      e.HostName,
		Message:       e.Message,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	})
}

// GetAllEvents gets all events using sqlc
func (m *postgresDBRepo) GetAllEvents() ([]models.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.q.GetAllEvents(ctx)
	if err != nil {
		return nil, err
	}

	var events []models.Event
	for _, r := range rows {
		ev := models.Event{
			ID:            int(r.ID),
			EventType:     r.EventType,
			HostServiceID: int(r.HostServiceID),
			HostID:        int(r.HostID),
			ServiceName:   r.ServiceName,
			HostName:      r.HostName,
			Message:       r.Message,
			CreatedAt:     r.CreatedAt,
			UpdatedAt:     r.UpdatedAt,
		}
		events = append(events, ev)
	}

	return events, nil
}
