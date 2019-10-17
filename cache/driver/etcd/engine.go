package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"cloudtask-center/cache/driver/types"
	"cloudtask/common/models"
	"cloudtask/libtools/gounits/httpx"
	"cloudtask/libtools/gzkwrapper"
)

var defaultTimeout = 10 * time.Second

var (
	workLocationKey = "/cloudtask/location"
)

//Engine is exported
type Engine struct {
	client *httpx.HttpClient
	node   *gzkwrapper.Node
}

//NewEngine is exported
func NewEngine(hosts string, readPageSize int) *Engine {
	client := httpx.NewClient().
		SetTransport(&http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 60 * time.Second,
			}).DialContext,
			DisableKeepAlives:     false,
			MaxIdleConns:          25,
			MaxIdleConnsPerHost:   25,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   http.DefaultTransport.(*http.Transport).TLSHandshakeTimeout,
			ExpectContinueTimeout: http.DefaultTransport.(*http.Transport).ExpectContinueTimeout,
		})

	node := gzkwrapper.NewNode(hosts)
	err := node.Open()
	if err != nil {
		fmt.Println("connect etcd fail:", err)
	}

	return &Engine{
		client: client,
		node:   node,
	}
}

func (engine *Engine) SetConfigParameters(rawAPIURL string, readPageSize int) {
}

func (engine *Engine) getLocation(location string) (*models.WorkLocation, error) {
	workRaw, err := engine.node.Get(workLocationKey)
	if err != nil {
		return nil, err
	}

	workLocation := &models.WorkLocation{}
	if err := json.Unmarshal(workRaw, workLocation); err != nil {
		return nil, err
	}

	return workLocation, nil
}

func (engine *Engine) postLocation(workLocation *models.WorkLocation) error {
	workJson, err := json.Marshal(workLocation)
	if err != nil {
		return err
	}

	return engine.node.Set(workLocationKey, workJson)
}

func (engine *Engine) readLocationsName() ([]string, error) {
	// 暂时只支持一个location
	return []string{"myCluster"}, nil
}

func (engine *Engine) readSimpleJobs(query map[string][]string) ([]*models.SimpleJob, error) {
	fmt.Println("read simple jobs")

	var (
		pageIndex = 1
		jobs      = []*models.SimpleJob{}
	)

	query["fields"] = []string{`["jobid","name","location","groupid","servers","enabled","stat"]`}
	for {
		query["pageIndex"] = []string{strconv.Itoa(pageIndex)}
		respData, err := engine.client.Get(context.Background(), "/sys_jobs", query, nil)
		if err != nil {
			return nil, err
		}

		defer respData.Close()
		statusCode := respData.StatusCode()
		if statusCode >= http.StatusBadRequest {
			return nil, fmt.Errorf("HTTP GET sys_jobs jobs failure %d.", statusCode)
		}

		respSimpleJobs, err := parseSimpleJobsResponse(respData)
		if err != nil {
			return nil, err
		}

		jobs = append(jobs, respSimpleJobs.Jobs...)
		if len(jobs) != respSimpleJobs.TotalRows {
			pageIndex++
			continue
		}
		break
	}
	return jobs, nil
}

func (engine *Engine) readJobs(query map[string][]string) ([]*models.Job, error) {
	fmt.Println("read jobs")

	var (
		pageIndex = 1
		jobs      = []*models.Job{}
	)

	for {
		query["pageIndex"] = []string{strconv.Itoa(pageIndex)}
		respData, err := engine.client.Get(context.Background(), "/sys_jobs", query, nil)
		if err != nil {
			return nil, err
		}

		defer respData.Close()
		statusCode := respData.StatusCode()
		if statusCode >= http.StatusBadRequest {
			return nil, fmt.Errorf("HTTP GET sys_jobs jobs failure %d.", statusCode)
		}

		respJobs, err := parseJobsResponse(respData)
		if err != nil {
			return nil, err
		}

		jobs = append(jobs, respJobs.Jobs...)
		if len(jobs) != respJobs.TotalRows {
			pageIndex++
			continue
		}
		break
	}
	return jobs, nil
}

func (engine *Engine) getSimpleJob(jobid string) (*models.SimpleJob, error) {

	job, err := engine.getJob(jobid)
	if err != nil {
		return nil, err
	}

	return &models.SimpleJob{
		JobId:    job.JobId,
		Name:     job.Name,
		Location: job.Location,
		GroupId:  job.GroupId,
		Servers:  job.Servers,
		Enabled:  job.Enabled,
		Stat:     job.Stat,
	}, nil
}

func (engine *Engine) getJob(jobid string) (*models.Job, error) {
	fmt.Println("get job")

	respData, err := engine.client.Get(context.Background(), "/sys_jobs/"+jobid, nil, nil)
	if err != nil {
		return nil, err
	}

	defer respData.Close()
	statusCode := respData.StatusCode()
	if statusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("HTTP GET sys_jobs %s failure %d.", jobid, statusCode)
	}

	if statusCode == http.StatusNoContent {
		return nil, types.ErrDriverResourceNotFound
	}

	job := &models.Job{}
	if err := respData.JSON(job); err != nil {
		return nil, err
	}
	return job, nil
}

func (engine *Engine) putJob(job *models.Job) error {
	// TODO:
	fmt.Println("put job")
	return nil
}

func (engine *Engine) postJobLog(jobLog *models.JobLog) error {
	// TODO:
	fmt.Println("post job log")
	return nil
}
