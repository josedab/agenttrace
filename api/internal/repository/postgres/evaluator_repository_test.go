package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// createTestEvaluator creates an evaluator with test data
func createTestEvaluator(name string, projectID, createdBy uuid.UUID) *domain.Evaluator {
	now := time.Now()
	return &domain.Evaluator{
		ID:              uuid.New(),
		ProjectID:       projectID,
		Name:            name,
		Description:     "Test evaluator description",
		Type:            domain.EvaluatorTypeLLM,
		Config:          `{"model": "gpt-4"}`,
		PromptTemplate:  "Evaluate this: {{input}}",
		Variables:       []string{"input"},
		TargetFilter:    `{"type": "generation"}`,
		SamplingRate:    1.0,
		ScoreName:       "quality",
		ScoreDataType:   domain.ScoreDataTypeNumeric,
		ScoreCategories: []string{},
		Enabled:         true,
		CreatedBy:       &createdBy,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// createTestAnnotationQueue creates an annotation queue with test data
func createTestAnnotationQueue(name string, projectID uuid.UUID) *domain.AnnotationQueue {
	now := time.Now()
	return &domain.AnnotationQueue{
		ID:          uuid.New(),
		ProjectID:   projectID,
		Name:        name,
		Description: "Test annotation queue",
		ScoreName:   "human-rating",
		ScoreConfig: `{"min": 1, "max": 5}`,
		Filters:     `{"type": "generation"}`,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func cleanupEvaluators(t *testing.T, db *database.PostgresDB, names ...string) {
	ctx := context.Background()
	for _, name := range names {
		// Clean up in dependency order
		_, _ = db.Pool.Exec(ctx, "DELETE FROM evaluation_jobs WHERE evaluator_id IN (SELECT id FROM evaluators WHERE name = $1)", name)
		_, _ = db.Pool.Exec(ctx, "DELETE FROM evaluators WHERE name = $1", name)
	}
}

func cleanupAnnotationQueues(t *testing.T, db *database.PostgresDB, names ...string) {
	ctx := context.Background()
	for _, name := range names {
		_, _ = db.Pool.Exec(ctx, "DELETE FROM annotation_queue_items WHERE queue_id IN (SELECT id FROM annotation_queues WHERE name = $1)", name)
		_, _ = db.Pool.Exec(ctx, "DELETE FROM annotation_queues WHERE name = $1", name)
	}
}

func TestEvaluatorRepository_Create(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	userRepo := NewUserRepository(db)
	evalRepo := NewEvaluatorRepository(db)
	ctx := context.Background()

	orgName := "Test Org for Evaluator Create"
	projectName := "Test Project for Evaluator Create"
	userEmail := "test-eval-create@example.com"
	evalName := "Test Evaluator Create"

	cleanupEvaluators(t, db, evalName)
	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	cleanupUsers(t, db, userEmail)
	defer cleanupEvaluators(t, db, evalName)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)
	defer cleanupUsers(t, db, userEmail)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	user := createTestUser(userEmail)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	eval := createTestEvaluator(evalName, project.ID, user.ID)
	err = evalRepo.Create(ctx, eval)
	require.NoError(t, err)

	// Verify by fetching
	fetched, err := evalRepo.GetByID(ctx, eval.ID)
	require.NoError(t, err)
	assert.Equal(t, eval.ID, fetched.ID)
	assert.Equal(t, eval.Name, fetched.Name)
}

func TestEvaluatorRepository_GetByID(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	userRepo := NewUserRepository(db)
	evalRepo := NewEvaluatorRepository(db)
	ctx := context.Background()

	orgName := "Test Org for Evaluator GetByID"
	projectName := "Test Project for Evaluator GetByID"
	userEmail := "test-eval-getbyid@example.com"
	evalName := "Test Evaluator GetByID"

	cleanupEvaluators(t, db, evalName)
	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	cleanupUsers(t, db, userEmail)
	defer cleanupEvaluators(t, db, evalName)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)
	defer cleanupUsers(t, db, userEmail)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	user := createTestUser(userEmail)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	eval := createTestEvaluator(evalName, project.ID, user.ID)
	err = evalRepo.Create(ctx, eval)
	require.NoError(t, err)

	t.Run("existing evaluator", func(t *testing.T) {
		fetched, err := evalRepo.GetByID(ctx, eval.ID)
		require.NoError(t, err)
		assert.Equal(t, eval.ID, fetched.ID)
	})

	t.Run("non-existent evaluator", func(t *testing.T) {
		_, err := evalRepo.GetByID(ctx, uuid.New())
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestEvaluatorRepository_GetByName(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	userRepo := NewUserRepository(db)
	evalRepo := NewEvaluatorRepository(db)
	ctx := context.Background()

	orgName := "Test Org for Evaluator GetByName"
	projectName := "Test Project for Evaluator GetByName"
	userEmail := "test-eval-getbyname@example.com"
	evalName := "Test Evaluator GetByName"

	cleanupEvaluators(t, db, evalName)
	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	cleanupUsers(t, db, userEmail)
	defer cleanupEvaluators(t, db, evalName)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)
	defer cleanupUsers(t, db, userEmail)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	user := createTestUser(userEmail)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	eval := createTestEvaluator(evalName, project.ID, user.ID)
	err = evalRepo.Create(ctx, eval)
	require.NoError(t, err)

	t.Run("existing name", func(t *testing.T) {
		fetched, err := evalRepo.GetByName(ctx, project.ID, evalName)
		require.NoError(t, err)
		assert.Equal(t, eval.ID, fetched.ID)
	})

	t.Run("non-existent name", func(t *testing.T) {
		_, err := evalRepo.GetByName(ctx, project.ID, "nonexistent")
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestEvaluatorRepository_Update(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	userRepo := NewUserRepository(db)
	evalRepo := NewEvaluatorRepository(db)
	ctx := context.Background()

	orgName := "Test Org for Evaluator Update"
	projectName := "Test Project for Evaluator Update"
	userEmail := "test-eval-update@example.com"
	evalName := "Test Evaluator Update"

	cleanupEvaluators(t, db, evalName)
	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	cleanupUsers(t, db, userEmail)
	defer cleanupEvaluators(t, db, evalName)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)
	defer cleanupUsers(t, db, userEmail)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	user := createTestUser(userEmail)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	eval := createTestEvaluator(evalName, project.ID, user.ID)
	err = evalRepo.Create(ctx, eval)
	require.NoError(t, err)

	// Update
	eval.Description = "Updated description"
	eval.Enabled = false
	eval.SamplingRate = 0.5
	err = evalRepo.Update(ctx, eval)
	require.NoError(t, err)

	// Verify
	fetched, err := evalRepo.GetByID(ctx, eval.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated description", fetched.Description)
	assert.False(t, fetched.Enabled)
	assert.Equal(t, 0.5, fetched.SamplingRate)
}

func TestEvaluatorRepository_Delete(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	userRepo := NewUserRepository(db)
	evalRepo := NewEvaluatorRepository(db)
	ctx := context.Background()

	orgName := "Test Org for Evaluator Delete"
	projectName := "Test Project for Evaluator Delete"
	userEmail := "test-eval-delete@example.com"
	evalName := "Test Evaluator Delete"

	cleanupEvaluators(t, db, evalName)
	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	cleanupUsers(t, db, userEmail)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)
	defer cleanupUsers(t, db, userEmail)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	user := createTestUser(userEmail)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	eval := createTestEvaluator(evalName, project.ID, user.ID)
	err = evalRepo.Create(ctx, eval)
	require.NoError(t, err)

	// Delete
	err = evalRepo.Delete(ctx, eval.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = evalRepo.GetByID(ctx, eval.ID)
	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestEvaluatorRepository_List(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	userRepo := NewUserRepository(db)
	evalRepo := NewEvaluatorRepository(db)
	ctx := context.Background()

	orgName := "Test Org for Evaluator List"
	projectName := "Test Project for Evaluator List"
	userEmail := "test-eval-list@example.com"
	evalName1 := "Test Evaluator List A"
	evalName2 := "Test Evaluator List B"

	cleanupEvaluators(t, db, evalName1, evalName2)
	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	cleanupUsers(t, db, userEmail)
	defer cleanupEvaluators(t, db, evalName1, evalName2)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)
	defer cleanupUsers(t, db, userEmail)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	user := createTestUser(userEmail)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create evaluators
	eval1 := createTestEvaluator(evalName1, project.ID, user.ID)
	err = evalRepo.Create(ctx, eval1)
	require.NoError(t, err)

	eval2 := createTestEvaluator(evalName2, project.ID, user.ID)
	eval2.Enabled = false
	err = evalRepo.Create(ctx, eval2)
	require.NoError(t, err)

	t.Run("basic list", func(t *testing.T) {
		filter := &domain.EvaluatorFilter{ProjectID: project.ID}
		list, err := evalRepo.List(ctx, filter, 10, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(2), list.TotalCount)
	})

	t.Run("filter by enabled", func(t *testing.T) {
		enabled := true
		filter := &domain.EvaluatorFilter{
			ProjectID: project.ID,
			Enabled:   &enabled,
		}
		list, err := evalRepo.List(ctx, filter, 10, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(1), list.TotalCount)
	})

	t.Run("filter by name", func(t *testing.T) {
		nameFilter := "List A"
		filter := &domain.EvaluatorFilter{
			ProjectID: project.ID,
			Name:      &nameFilter,
		}
		list, err := evalRepo.List(ctx, filter, 10, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(1), list.TotalCount)
	})
}

func TestEvaluatorRepository_ListEnabled(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	userRepo := NewUserRepository(db)
	evalRepo := NewEvaluatorRepository(db)
	ctx := context.Background()

	orgName := "Test Org for Evaluator ListEnabled"
	projectName := "Test Project for Evaluator ListEnabled"
	userEmail := "test-eval-listenabled@example.com"
	evalName1 := "Test Evaluator ListEnabled A"
	evalName2 := "Test Evaluator ListEnabled B"

	cleanupEvaluators(t, db, evalName1, evalName2)
	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	cleanupUsers(t, db, userEmail)
	defer cleanupEvaluators(t, db, evalName1, evalName2)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)
	defer cleanupUsers(t, db, userEmail)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	user := createTestUser(userEmail)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	eval1 := createTestEvaluator(evalName1, project.ID, user.ID)
	eval1.Enabled = true
	err = evalRepo.Create(ctx, eval1)
	require.NoError(t, err)

	eval2 := createTestEvaluator(evalName2, project.ID, user.ID)
	eval2.Enabled = false
	err = evalRepo.Create(ctx, eval2)
	require.NoError(t, err)

	enabled, err := evalRepo.ListEnabled(ctx, project.ID)
	require.NoError(t, err)
	assert.Len(t, enabled, 1)
	assert.Equal(t, evalName1, enabled[0].Name)
}

func TestEvaluatorRepository_NameExists(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	userRepo := NewUserRepository(db)
	evalRepo := NewEvaluatorRepository(db)
	ctx := context.Background()

	orgName := "Test Org for Evaluator NameExists"
	projectName := "Test Project for Evaluator NameExists"
	userEmail := "test-eval-nameexists@example.com"
	evalName := "Test Evaluator NameExists"

	cleanupEvaluators(t, db, evalName)
	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	cleanupUsers(t, db, userEmail)
	defer cleanupEvaluators(t, db, evalName)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)
	defer cleanupUsers(t, db, userEmail)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	user := createTestUser(userEmail)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("name does not exist", func(t *testing.T) {
		exists, err := evalRepo.NameExists(ctx, project.ID, evalName)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("name exists", func(t *testing.T) {
		eval := createTestEvaluator(evalName, project.ID, user.ID)
		err := evalRepo.Create(ctx, eval)
		require.NoError(t, err)

		exists, err := evalRepo.NameExists(ctx, project.ID, evalName)
		require.NoError(t, err)
		assert.True(t, exists)
	})
}

func TestEvaluatorRepository_Jobs(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	userRepo := NewUserRepository(db)
	evalRepo := NewEvaluatorRepository(db)
	ctx := context.Background()

	orgName := "Test Org for Evaluator Jobs"
	projectName := "Test Project for Evaluator Jobs"
	userEmail := "test-eval-jobs@example.com"
	evalName := "Test Evaluator Jobs"

	cleanupEvaluators(t, db, evalName)
	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	cleanupUsers(t, db, userEmail)
	defer cleanupEvaluators(t, db, evalName)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)
	defer cleanupUsers(t, db, userEmail)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	user := createTestUser(userEmail)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	eval := createTestEvaluator(evalName, project.ID, user.ID)
	err = evalRepo.Create(ctx, eval)
	require.NoError(t, err)

	traceID := uuid.New().String()

	t.Run("create job", func(t *testing.T) {
		job := &domain.EvaluationJob{
			ID:          uuid.New(),
			EvaluatorID: eval.ID,
			TraceID:     traceID,
			Status:      "pending",
			ScheduledAt: time.Now(),
			CreatedAt:   time.Now(),
		}
		err := evalRepo.CreateJob(ctx, job)
		require.NoError(t, err)
	})

	t.Run("job exists", func(t *testing.T) {
		exists, err := evalRepo.JobExists(ctx, eval.ID, traceID, nil)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("job does not exist", func(t *testing.T) {
		exists, err := evalRepo.JobExists(ctx, eval.ID, "nonexistent", nil)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("list pending jobs", func(t *testing.T) {
		jobs, err := evalRepo.ListPendingJobs(ctx, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(jobs), 1)
	})

	t.Run("get and update job", func(t *testing.T) {
		jobs, err := evalRepo.ListPendingJobs(ctx, 10)
		require.NoError(t, err)

		var jobID uuid.UUID
		for _, j := range jobs {
			if j.TraceID == traceID {
				jobID = j.ID
				break
			}
		}

		job, err := evalRepo.GetJobByID(ctx, jobID)
		require.NoError(t, err)

		now := time.Now()
		job.Status = "completed"
		job.StartedAt = &now
		job.CompletedAt = &now
		resultStr := `{"score": 0.85}`
		job.Result = &resultStr
		err = evalRepo.UpdateJob(ctx, job)
		require.NoError(t, err)

		fetched, err := evalRepo.GetJobByID(ctx, jobID)
		require.NoError(t, err)
		assert.Equal(t, "completed", fetched.Status)
	})

	t.Run("get eval count", func(t *testing.T) {
		count, err := evalRepo.GetEvalCount(ctx, eval.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})
}

func TestEvaluatorRepository_AnnotationQueues(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	userRepo := NewUserRepository(db)
	evalRepo := NewEvaluatorRepository(db)
	ctx := context.Background()

	orgName := "Test Org for Annotation Queues"
	projectName := "Test Project for Annotation Queues"
	userEmail := "test-annotation-queue@example.com"
	queueName := "Test Annotation Queue"

	cleanupAnnotationQueues(t, db, queueName)
	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	cleanupUsers(t, db, userEmail)
	defer cleanupAnnotationQueues(t, db, queueName)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)
	defer cleanupUsers(t, db, userEmail)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	user := createTestUser(userEmail)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	var queueID uuid.UUID

	t.Run("create annotation queue", func(t *testing.T) {
		queue := createTestAnnotationQueue(queueName, project.ID)
		queueID = queue.ID
		err := evalRepo.CreateAnnotationQueue(ctx, queue)
		require.NoError(t, err)
	})

	t.Run("get annotation queue", func(t *testing.T) {
		fetched, err := evalRepo.GetAnnotationQueueByID(ctx, queueID)
		require.NoError(t, err)
		assert.Equal(t, queueName, fetched.Name)
	})

	t.Run("update annotation queue", func(t *testing.T) {
		queue, err := evalRepo.GetAnnotationQueueByID(ctx, queueID)
		require.NoError(t, err)

		queue.Description = "Updated description"
		err = evalRepo.UpdateAnnotationQueue(ctx, queue)
		require.NoError(t, err)

		fetched, err := evalRepo.GetAnnotationQueueByID(ctx, queueID)
		require.NoError(t, err)
		assert.Equal(t, "Updated description", fetched.Description)
	})

	t.Run("list annotation queues", func(t *testing.T) {
		queues, err := evalRepo.ListAnnotationQueues(ctx, project.ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(queues), 1)
	})

	t.Run("annotation queue items", func(t *testing.T) {
		item := &domain.AnnotationQueueItem{
			ID:        uuid.New(),
			QueueID:   queueID,
			TraceID:   uuid.New().String(),
			Status:    "pending",
			CreatedAt: time.Now(),
		}
		err := evalRepo.CreateAnnotationQueueItem(ctx, item)
		require.NoError(t, err)

		// Get next item
		nextItem, err := evalRepo.GetNextAnnotationItem(ctx, queueID)
		require.NoError(t, err)
		assert.Equal(t, item.ID, nextItem.ID)

		// Get stats
		pending, completed, err := evalRepo.GetAnnotationQueueStats(ctx, queueID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), pending)
		assert.Equal(t, int64(0), completed)

		// Complete item
		err = evalRepo.CompleteAnnotationItem(ctx, item.ID, user.ID)
		require.NoError(t, err)

		pending, completed, err = evalRepo.GetAnnotationQueueStats(ctx, queueID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), pending)
		assert.Equal(t, int64(1), completed)
	})

	t.Run("delete annotation queue", func(t *testing.T) {
		err := evalRepo.DeleteAnnotationQueue(ctx, queueID)
		require.NoError(t, err)

		_, err = evalRepo.GetAnnotationQueueByID(ctx, queueID)
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}
