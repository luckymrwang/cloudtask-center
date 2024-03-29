package etcd

import (
	"errors"
	"net/url"
	"path"
	"strconv"
	"sync"

	"cloudtask-center/cache/driver"
	"cloudtask-center/cache/driver/types"
	"cloudtask/common/models"
	"cloudtask/libtools/gounits/logger"
)

const (
	defaultReadPageSize = 512
)

var (
	ErrNgCloudStorageDriverURLInvalid = errors.New("ngcloud storage driver apiURL invalid.")
)

//NgCloudStorageDriver is exported
type NgCloudStorageDriver struct {
	sync.RWMutex
	driver.StorageDriver
	engine *Engine
}

func init() {
	driver.AddDriver(types.ETCD, New)
}

//New is exported
func New(parameters types.Parameters) (driver.StorageDriver, error) {
	return &NgCloudStorageDriver{
		engine: NewEngine(parameters["hosts"].(string), 0),
	}, nil
}

func parseEngineConfigs(parameters types.Parameters) (string, int, error) {
	var (
		value        interface{}
		ret          bool
		readPageSize int
		rawAPIURL    string
	)

	readPageSize = defaultReadPageSize
	value, ret = parameters["readpagesize"]
	if ret {
		if pValue, err := strconv.Atoi(value.(string)); err == nil {
			readPageSize = pValue
		}
	}

	value, ret = parameters["apiurl"]
	if !ret {
		return "", 0, ErrNgCloudStorageDriverURLInvalid
	}

	pRawURL, err := url.Parse(value.(string))
	if err != nil {
		return "", 0, ErrNgCloudStorageDriverURLInvalid
	}

	scheme := pRawURL.Scheme
	if scheme == "" {
		scheme = "http"
	}

	rawAPIURL = scheme + "://" + pRawURL.Host + path.Clean(pRawURL.Path)
	if pRawURL.RawQuery != "" {
		rawAPIURL = rawAPIURL + "?" + pRawURL.RawQuery
	}
	return rawAPIURL, readPageSize, nil
}

//Open is exported
func (driver *NgCloudStorageDriver) Open() error {

	return nil
}

//Close is exported
func (driver *NgCloudStorageDriver) Close() {
}

//SetConfigParameters is exported
func (driver *NgCloudStorageDriver) SetConfigParameters(parameters types.Parameters) {

	rawAPIURL, readPageSize, err := parseEngineConfigs(parameters)
	if err != nil {
		logger.ERROR("[#cache#] ngcloud driver parse configs error, %s", err.Error())
		return
	}
	driver.engine.SetConfigParameters(rawAPIURL, readPageSize)
	logger.ERROR("[#cache#] ngcloud driver configs changed, %s %s", rawAPIURL, readPageSize)
}

//GetLocationsName is exported
func (driver *NgCloudStorageDriver) GetLocationsName() []string {

	driver.RLock()
	defer driver.RUnlock()
	names, err := driver.engine.readLocationsName()
	if err != nil {
		logger.ERROR("[#cache#] engine read locations name error, %s", err.Error())
		return []string{}
	}
	return names
}

//GetLocation is exported
func (driver *NgCloudStorageDriver) GetLocation(location string) *models.WorkLocation {

	driver.RLock()
	defer driver.RUnlock()
	workLocation, err := driver.engine.getLocation(location)
	if err != nil {
		logger.ERROR("[#cache#] engine read location %s error, %s", location, err.Error())
		return nil
	}
	return workLocation
}

//GetLocationSimpleJobs is exported
func (driver *NgCloudStorageDriver) GetLocationSimpleJobs(location string) []*models.SimpleJob {

	driver.RLock()
	defer driver.RUnlock()
	query := map[string][]string{"f_location": []string{location}}
	jobs, err := driver.engine.readSimpleJobs(query)
	if err != nil {
		logger.ERROR("[#cache#] engine read simple jobs %+v error, %s", query, err.Error())
		return []*models.SimpleJob{}
	}
	return jobs
}

//GetSimpleJob is exported
func (driver *NgCloudStorageDriver) GetSimpleJob(jobid string) *models.SimpleJob {

	driver.RLock()
	defer driver.RUnlock()
	job, err := driver.engine.getSimpleJob(jobid)
	if err != nil {
		logger.ERROR("[#cache#] engine get simple job %s error, %s", jobid, err.Error())
		return nil
	}
	return job
}

//GetJobs is exported
func (driver *NgCloudStorageDriver) GetJobs() []*models.Job {

	driver.RLock()
	defer driver.RUnlock()
	query := map[string][]string{}
	jobs, err := driver.engine.readJobs(query)
	if err != nil {
		logger.ERROR("[#cache#] engine read jobs %+v error, %s", query, err.Error())
		return []*models.Job{}
	}
	return jobs
}

//GetStateJobs is exported
func (driver *NgCloudStorageDriver) GetStateJobs(state int) []*models.Job {

	driver.RLock()
	defer driver.RUnlock()
	query := map[string][]string{"f_stat": []string{strconv.Itoa(state)}}
	jobs, err := driver.engine.readJobs(query)
	if err != nil {
		logger.ERROR("[#cache#] engine read jobs %+v error, %s", query, err.Error())
		return []*models.Job{}
	}
	return jobs
}

//GetLocationJobs is exported
func (driver *NgCloudStorageDriver) GetLocationJobs(location string) []*models.Job {

	driver.RLock()
	defer driver.RUnlock()
	query := map[string][]string{"f_location": []string{location}}
	jobs, err := driver.engine.readJobs(query)
	if err != nil {
		logger.ERROR("[#cache#] engine read jobs %+v error, %s", query, err.Error())
		return []*models.Job{}
	}
	return jobs
}

//GetGroupJobs is exported
func (driver *NgCloudStorageDriver) GetGroupJobs(groupid string) []*models.Job {

	driver.RLock()
	defer driver.RUnlock()
	query := map[string][]string{"f_groupid": []string{groupid}}
	jobs, err := driver.engine.readJobs(query)
	if err != nil {
		logger.ERROR("[#cache#] engine read jobs %+v error, %s", query, err.Error())
		return []*models.Job{}
	}
	return jobs
}

//GetJob is exported
func (driver *NgCloudStorageDriver) GetJob(jobid string) *models.Job {

	driver.RLock()
	defer driver.RUnlock()
	job, err := driver.engine.getJob(jobid)
	if err != nil {
		logger.ERROR("[#cache#] engine get job %s error, %s", jobid, err.Error())
		return nil
	}
	return job
}

//SetJob is exported
func (driver *NgCloudStorageDriver) SetJob(job *models.Job) {

	driver.Lock()
	defer driver.Unlock()
	if err := driver.engine.putJob(job); err != nil {
		logger.ERROR("[#cache#] engine set job %s error, %s", job.JobId, err.Error())
	}
}

//SetJobLog is exported
func (driver *NgCloudStorageDriver) SetJobLog(joblog *models.JobLog) {

	driver.Lock()
	defer driver.Unlock()
	if err := driver.engine.postJobLog(joblog); err != nil {
		logger.ERROR("[#cache#] engine post job %s error, %s", joblog.JobId, err.Error())
	}
}
