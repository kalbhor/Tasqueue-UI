package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/kalbhor/tasqueue/v2"
	inmemorybroker "github.com/kalbhor/tasqueue/v2/brokers/in-memory"
	redisbroker "github.com/kalbhor/tasqueue/v2/brokers/redis"
	inmemoryresults "github.com/kalbhor/tasqueue/v2/results/in-memory"
	redisresults "github.com/kalbhor/tasqueue/v2/results/redis"

	"github.com/kalbhor/tasqueue-ui/internal/config"
)

// Service provides access to Tasqueue data
type Service struct {
	server *tasqueue.Server
	broker tasqueue.Broker
	config config.Config
}

// DashboardStats holds overview statistics
type DashboardStats struct {
	TotalPending    int            `json:"total_pending"`
	TotalSuccess    int            `json:"total_success"`
	TotalFailed     int            `json:"total_failed"`
	QueueStats      map[string]int `json:"queue_stats"`
	RegisteredTasks []string       `json:"registered_tasks"`
}

// JobDetail extends JobMessage with additional computed fields
type JobDetail struct {
	tasqueue.JobMessage
	ResultData []byte `json:"result_data,omitempty"`
}

// ChainDetail extends ChainMessage with job details
type ChainDetail struct {
	tasqueue.ChainMessage
	Jobs []tasqueue.JobMessage `json:"jobs"`
}

// GroupDetail extends GroupMessage with job details
type GroupDetail struct {
	tasqueue.GroupMessage
	Jobs []tasqueue.JobMessage `json:"jobs"`
}

// NewService creates a new Tasqueue service instance
func NewService(cfg config.Config) (*Service, error) {
	var (
		broker  tasqueue.Broker
		results tasqueue.Results
		err     error
	)

	switch cfg.Broker.Type {
	case "redis":
		broker = redisbroker.New(redisbroker.Options{
			Addrs:    []string{cfg.Broker.Redis.Addr},
			Password: cfg.Broker.Redis.Password,
			DB:       cfg.Broker.Redis.DB,
		}, slog.Default())

		results = redisresults.New(redisresults.Options{
			Addrs:    []string{cfg.Broker.Redis.Addr},
			Password: cfg.Broker.Redis.Password,
			DB:       cfg.Broker.Redis.DB,
		}, slog.Default())

	case "in-memory":
		broker = inmemorybroker.New()
		results = inmemoryresults.New()

	// TODO: Add NATS JetStream support
	// case "nats-js":
	//     ...

	default:
		return nil, fmt.Errorf("unsupported broker type: %s", cfg.Broker.Type)
	}

	// Create server instance (read-only, no task handlers registered)
	srv, err := tasqueue.NewServer(tasqueue.ServerOpts{
		Broker:  broker,
		Results: results,
		Logger:  slog.Default().Handler(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create tasqueue server: %w", err)
	}

	return &Service{
		server: srv,
		broker: broker,
		config: cfg,
	}, nil
}

// GetDashboardStats returns overview statistics
func (s *Service) GetDashboardStats(ctx context.Context) (DashboardStats, error) {
	stats := DashboardStats{
		QueueStats: make(map[string]int),
	}

	// Get successful jobs count
	successJobs, err := s.server.GetSuccess(ctx)
	if err != nil {
		return stats, fmt.Errorf("failed to get success jobs: %w", err)
	}
	stats.TotalSuccess = len(successJobs)

	// Get failed jobs count
	failedJobs, err := s.server.GetFailed(ctx)
	if err != nil {
		return stats, fmt.Errorf("failed to get failed jobs: %w", err)
	}
	stats.TotalFailed = len(failedJobs)

	// Get pending jobs count for default queue using the new GetPendingCount method
	pendingCount, err := s.server.GetPendingCount(ctx, tasqueue.DefaultQueue)
	if err == nil {
		stats.TotalPending = int(pendingCount)
		stats.QueueStats[tasqueue.DefaultQueue] = int(pendingCount)
	}

	// Get registered tasks
	registeredTasks, err := s.server.GetTasks()
	if err != nil {
		return stats, fmt.Errorf("failed to get registered tasks: %w", err)
	}
	// Extract task names from TaskInfo structs
	taskNames := make([]string, len(registeredTasks))
	for i, task := range registeredTasks {
		taskNames[i] = task.Name
	}
	stats.RegisteredTasks = taskNames

	return stats, nil
}

// GetJob returns a specific job by ID with its result data
func (s *Service) GetJob(ctx context.Context, id string) (JobDetail, error) {
	job, err := s.server.GetJob(ctx, id)
	if err != nil {
		return JobDetail{}, fmt.Errorf("failed to get job: %w", err)
	}

	detail := JobDetail{
		JobMessage: job,
	}

	// Try to fetch result data if available
	resultData, err := s.server.GetResult(ctx, id)
	if err == nil {
		detail.ResultData = resultData
	}

	return detail, nil
}

// PendingJobsResult holds paginated pending jobs with metadata
type PendingJobsResult struct {
	Jobs   []tasqueue.JobMessage `json:"jobs"`
	Total  int64                 `json:"total"`
	Offset int                   `json:"offset"`
	Limit  int                   `json:"limit"`
}

// GetPendingJobs returns pending jobs for a specific queue
// Deprecated: Use GetPendingJobsWithPagination for better performance
func (s *Service) GetPendingJobs(ctx context.Context, queue string) ([]tasqueue.JobMessage, error) {
	if queue == "" {
		queue = tasqueue.DefaultQueue
	}

	jobs, err := s.server.GetPending(ctx, queue)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending jobs: %w", err)
	}

	return jobs, nil
}

// GetPendingJobsWithPagination returns paginated pending jobs for a specific queue
func (s *Service) GetPendingJobsWithPagination(ctx context.Context, queue string, offset, limit int) (PendingJobsResult, error) {
	if queue == "" {
		queue = tasqueue.DefaultQueue
	}

	// Set default pagination values
	if limit <= 0 {
		limit = 20 // Default limit
	}
	if offset < 0 {
		offset = 0
	}

	jobs, total, err := s.server.GetPendingWithPagination(ctx, queue, offset, limit)
	if err != nil {
		return PendingJobsResult{}, fmt.Errorf("failed to get pending jobs: %w", err)
	}

	return PendingJobsResult{
		Jobs:   jobs,
		Total:  total,
		Offset: offset,
		Limit:  limit,
	}, nil
}

// GetPendingCount returns the count of pending jobs in a specific queue
func (s *Service) GetPendingCount(ctx context.Context, queue string) (int64, error) {
	if queue == "" {
		queue = tasqueue.DefaultQueue
	}

	count, err := s.server.GetPendingCount(ctx, queue)
	if err != nil {
		return 0, fmt.Errorf("failed to get pending count: %w", err)
	}

	return count, nil
}

// GetJobsByStatus returns jobs filtered by status
func (s *Service) GetJobsByStatus(ctx context.Context, status string) ([]string, error) {
	switch status {
	case "successful":
		return s.server.GetSuccess(ctx)
	case "failed":
		return s.server.GetFailed(ctx)
	default:
		return nil, fmt.Errorf("unsupported status filter: %s", status)
	}
}

// GetChain returns chain details with all job information
func (s *Service) GetChain(ctx context.Context, id string) (ChainDetail, error) {
	chain, err := s.server.GetChain(ctx, id)
	if err != nil {
		return ChainDetail{}, fmt.Errorf("failed to get chain: %w", err)
	}

	detail := ChainDetail{
		ChainMessage: chain,
		Jobs:         make([]tasqueue.JobMessage, 0),
	}

	// Fetch all jobs in the chain
	allJobIDs := append(chain.PrevJobs, chain.JobID)
	for _, jobID := range allJobIDs {
		if jobID == "" {
			continue
		}
		job, err := s.server.GetJob(ctx, jobID)
		if err == nil {
			detail.Jobs = append(detail.Jobs, job)
		}
	}

	return detail, nil
}

// GetGroup returns group details with all job information
func (s *Service) GetGroup(ctx context.Context, id string) (GroupDetail, error) {
	group, err := s.server.GetGroup(ctx, id)
	if err != nil {
		return GroupDetail{}, fmt.Errorf("failed to get group: %w", err)
	}

	detail := GroupDetail{
		GroupMessage: group,
		Jobs:         make([]tasqueue.JobMessage, 0),
	}

	// Fetch all jobs in the group
	for jobID := range group.JobStatus {
		job, err := s.server.GetJob(ctx, jobID)
		if err == nil {
			detail.Jobs = append(detail.Jobs, job)
		}
	}

	return detail, nil
}

// ListChains returns all chain IDs from the results store
// Note: This requires scanning the results store with the chain prefix
func (s *Service) ListChains(ctx context.Context) ([]string, error) {
	// This is a placeholder - actual implementation depends on results backend
	// For Redis, we'd use SCAN with pattern "chain:msg:*"
	// For now, return empty list
	return []string{}, nil
}

// ListGroups returns all group IDs from the results store
func (s *Service) ListGroups(ctx context.Context) ([]string, error) {
	// This is a placeholder - actual implementation depends on results backend
	// For Redis, we'd use SCAN with pattern "group:msg:*"
	// For now, return empty list
	return []string{}, nil
}

// DeleteJob removes a job's metadata from the results store
func (s *Service) DeleteJob(ctx context.Context, id string) error {
	return s.server.DeleteJob(ctx, id)
}

// ScanKeys is a helper to scan keys with a prefix (Redis-specific)
// This is used to list all chains/groups
func (s *Service) ScanKeys(ctx context.Context, pattern string) ([]string, error) {
	// Only works with Redis backend
	if s.config.Broker.Type != "redis" {
		return nil, fmt.Errorf("scan operation only supported with Redis broker")
	}

	// Access the Redis client through the broker
	// This requires type assertion - handle carefully
	// For now, return empty list as placeholder
	return []string{}, nil
}

// ExtractIDFromKey extracts the ID from a prefixed key
func ExtractIDFromKey(key, prefix string) string {
	return strings.TrimPrefix(key, prefix)
}

// SearchResult holds the search results for jobs, chains, and groups
type SearchResult struct {
	Job   *JobDetail   `json:"job,omitempty"`
	Chain *ChainDetail `json:"chain,omitempty"`
	Group *GroupDetail `json:"group,omitempty"`
	Type  string       `json:"type"` // "job", "chain", "group", or "not_found"
}

// Search searches for a job, chain, or group by ID
// It tries to find the ID in all three types and returns the first match
func (s *Service) Search(ctx context.Context, id string) (SearchResult, error) {
	result := SearchResult{
		Type: "not_found",
	}

	// Try to get as job first
	job, err := s.GetJob(ctx, id)
	if err == nil {
		result.Job = &job
		result.Type = "job"
		return result, nil
	}

	// Try to get as chain
	chain, err := s.GetChain(ctx, id)
	if err == nil {
		result.Chain = &chain
		result.Type = "chain"
		return result, nil
	}

	// Try to get as group
	group, err := s.GetGroup(ctx, id)
	if err == nil {
		result.Group = &group
		result.Type = "group"
		return result, nil
	}

	return result, fmt.Errorf("no job, chain, or group found with ID: %s", id)
}
