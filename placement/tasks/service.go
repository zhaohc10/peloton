package tasks

import (
	"context"
	"errors"
	"time"

	"code.uber.internal/infra/peloton/.gen/peloton/private/resmgr"
	"code.uber.internal/infra/peloton/.gen/peloton/private/resmgrsvc"
	"code.uber.internal/infra/peloton/placement/config"
	"code.uber.internal/infra/peloton/placement/metrics"
	"code.uber.internal/infra/peloton/placement/models"
	log "github.com/sirupsen/logrus"
)

const (
	_timeout = 10 * time.Second
)

// Service will manage gangs/tasks and placements used by any placement strategy.
type Service interface {
	// Dequeue fetches some tasks from the resource manager.
	Dequeue(ctx context.Context, taskType resmgr.TaskType, batchSize int, timeout int) (assignments []*models.Assignment)

	// Enqueue returns dequeued tasks back to resource manager which as they were not placed.
	Enqueue(ctx context.Context, assignments []*models.Assignment, reason string)

	// SetPlacements sets placements in the resource manager.
	SetPlacements(ctx context.Context, placements []*resmgr.Placement)
}

// NewService will create a new task service.
func NewService(resourceManager resmgrsvc.ResourceManagerServiceYARPCClient,
	cfg *config.PlacementConfig,
	metrics *metrics.Metrics) Service {
	return &service{
		config:          cfg,
		resourceManager: resourceManager,
		metrics:         metrics,
	}
}

type service struct {
	config          *config.PlacementConfig
	resourceManager resmgrsvc.ResourceManagerServiceYARPCClient
	metrics         *metrics.Metrics
}

// Dequeue fetches some tasks from the resource manager.
func (s *service) Dequeue(
	ctx context.Context,
	taskType resmgr.TaskType,
	batchSize int,
	timeout int) []*models.Assignment {
	ctx, cancelFunc := context.WithTimeout(ctx, _timeout)
	defer cancelFunc()

	request := &resmgrsvc.DequeueGangsRequest{
		Limit:   uint32(batchSize),
		Type:    taskType,
		Timeout: uint32(timeout),
	}

	response, err := s.resourceManager.DequeueGangs(ctx, request)
	if err != nil {
		log.WithFields(log.Fields{
			log.ErrorKey: err,
			"request":    request,
			"response":   response,
		}).Error("dequeue gangs failed")
		return nil
	}

	numberOfTasks := 0
	for _, gang := range response.Gangs {
		numberOfTasks += len(gang.GetTasks())
	}

	// Create assignments from the tasks but without any offers
	assignments := make([]*models.Assignment, 0, numberOfTasks)
	now := time.Now()
	for _, gang := range response.Gangs {
		for _, task := range s.createTasks(gang, now) {
			assignments = append(assignments, models.NewAssignment(task))
		}
	}

	if len(assignments) > 0 {
		log.WithFields(log.Fields{
			"request":         request,
			"response":        response,
			"taskType":        taskType,
			"batchSize":       batchSize,
			"timeout":         timeout,
			"assignments_len": len(assignments),
			"assignments":     assignments,
		}).Debug("Dequeued gangs")
		log.WithField("tasks", len(assignments)).Info("Dequeued from task queue")
	}

	return assignments
}

// SetPlacements sets placements in the resource manager.
func (s *service) SetPlacements(ctx context.Context, placements []*resmgr.Placement) {
	if len(placements) == 0 {
		log.Debug("No task to place")
		return
	}

	setPlacementStart := time.Now()
	ctx, cancelFunc := context.WithTimeout(ctx, _timeout)
	defer cancelFunc()

	var request = &resmgrsvc.SetPlacementsRequest{
		Placements: placements,
	}
	response, err := s.resourceManager.SetPlacements(ctx, request)
	if err != nil {
		log.WithFields(log.Fields{
			"num_placements": len(placements),
			"error":          err.Error(),
		}).WithError(errors.New("failed to set placements"))
		return
	}

	log.WithFields(log.Fields{
		"request":  request,
		"response": response,
	}).Debug("set placements called")

	if response.GetError() != nil {
		log.WithFields(log.Fields{
			"num_placements": len(placements),
			"error":          response.Error.String(),
		}).Error("failed to place tasks")
		return
	}
	log.WithField("num_placements", len(placements)).Info("set placements succeeded")

	setPlacementDuration := time.Since(setPlacementStart)
	s.metrics.SetPlacementDuration.Record(setPlacementDuration)
	s.metrics.SetPlacementSuccess.Inc(int64(len(placements)))
}

// Enqueue calls resource manager to return those tasks which could not be placed
func (s *service) Enqueue(
	ctx context.Context,
	assignments []*models.Assignment,
	reason string) {
	if len(assignments) == 0 {
		return
	}

	ctx, cancelFunc := context.WithTimeout(ctx, _timeout)
	defer cancelFunc()

	gangs := make([]*resmgrsvc.Gang, 0, len(assignments))
	for _, assignment := range assignments {
		gangs = append(gangs, &resmgrsvc.Gang{
			Tasks: []*resmgr.Task{assignment.GetTask().GetTask()},
		})
	}

	var request = &resmgrsvc.EnqueueGangsRequest{
		Gangs:  gangs,
		Reason: reason,
	}
	response, err := s.resourceManager.EnqueueGangs(ctx, request)
	if err != nil {
		log.WithFields(log.Fields{
			"gangs": len(gangs),
			"error": err.Error(),
		}).WithError(errors.New("failed to return tasks"))
		return
	}

	log.WithFields(log.Fields{
		"request":  request,
		"response": response,
	}).Warn("enqueue gangs back to resmgr called")

	if response.GetError() != nil {
		log.WithFields(log.Fields{
			"gangs": len(gangs),
			"error": response.Error.String(),
		}).Error("failed to place tasks")
		return
	}
}

func (s *service) createTasks(gang *resmgrsvc.Gang, now time.Time) []*models.Task {
	var tasks []*models.Task
	resTasks := gang.GetTasks()
	if len(resTasks) == 0 {
		return tasks
	}
	// A value for maxRounds of <= 0 means there is no limit
	maxRounds := s.config.MaxRounds.Value(resTasks[0].Type)
	duration := s.config.MaxDurations.Value(resTasks[0].Type)
	deadline := now.Add(duration)
	for _, task := range resTasks {
		tasks = append(tasks, models.NewTask(gang, task, deadline, maxRounds))
	}
	return tasks
}
