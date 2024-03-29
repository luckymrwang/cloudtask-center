package driver

import "cloudtask-center/cache/driver/types"
import "cloudtask/common/models"

import (
	"fmt"
	"sort"
	"strings"
)

type Initialize func(parameters types.Parameters) (StorageDriver, error)

var (
	initializes      = make(map[types.Backend]Initialize)
	supportedBackend = func() string {
		keys := make([]string, 0, len(initializes))
		for k := range initializes {
			keys = append(keys, string(k))
		}
		sort.Strings(keys)
		return strings.Join(keys, ",")
	}()
)

//StorageDriver is exported
type StorageDriver interface {
	Open() error
	Close()
	SetConfigParameters(parameters types.Parameters)
	GetLocationsName() []string
	GetLocation(location string) *models.WorkLocation
	GetLocationSimpleJobs(location string) []*models.SimpleJob
	GetSimpleJob(jobid string) *models.SimpleJob
	GetJobs() []*models.Job
	GetStateJobs(state int) []*models.Job
	GetLocationJobs(location string) []*models.Job
	GetGroupJobs(groupid string) []*models.Job
	GetJob(jobid string) *models.Job
	SetJob(job *models.Job)
	SetJobLog(joblog *models.JobLog)
}

func NewDriver(backend types.Backend, parameters types.Parameters) (StorageDriver, error) {
	if init, exists := initializes[backend]; exists {
		return init(parameters)
	}
	return nil, fmt.Errorf("%s %s", types.ErrDriverNotSupported, supportedBackend)
}

func AddDriver(backend types.Backend, init Initialize) {
	initializes[backend] = init
}
