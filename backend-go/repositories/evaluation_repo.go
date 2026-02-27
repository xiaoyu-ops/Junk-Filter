package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/junkfilter/backend-go/models"
)

type EvaluationRepository struct {
	db *sql.DB
}

func NewEvaluationRepository(db *sql.DB) *EvaluationRepository {
	return &EvaluationRepository{db: db}
}

// Create inserts a new evaluation
func (er *EvaluationRepository) Create(ctx context.Context, req *models.EvaluationRequest) (*models.Evaluation, error) {
	evaluation := &models.Evaluation{
		ContentID:        req.ContentID,
		InnovationScore:  req.InnovationScore,
		DepthScore:       req.DepthScore,
		Decision:         req.Decision,
		Reasoning:        req.Reasoning,
		TLDR:             req.TLDR,
		KeyConcepts:      pq.StringArray(req.KeyConcepts),
		EvaluatedAt:      time.Now(),
		EvaluatorVersion: req.EvaluatorVersion,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Get the task_id from content
	var taskID uuid.UUID
	err := er.db.QueryRowContext(ctx,
		"SELECT task_id FROM content WHERE id = $1",
		req.ContentID,
	).Scan(&taskID)
	if err != nil {
		return nil, err
	}

	evaluation.TaskID = taskID

	err = er.db.QueryRowContext(ctx,
		`INSERT INTO evaluation (content_id, task_id, innovation_score, depth_score, decision,
		                         reasoning, tldr, key_concepts, evaluated_at, evaluator_version, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		 RETURNING id`,
		evaluation.ContentID, evaluation.TaskID, evaluation.InnovationScore, evaluation.DepthScore,
		evaluation.Decision, evaluation.Reasoning, evaluation.TLDR, evaluation.KeyConcepts,
		evaluation.EvaluatedAt, evaluation.EvaluatorVersion, evaluation.CreatedAt, evaluation.UpdatedAt,
	).Scan(&evaluation.ID)

	if err != nil {
		return nil, err
	}
	return evaluation, nil
}

// GetByContentID retrieves evaluation by content ID
func (er *EvaluationRepository) GetByContentID(ctx context.Context, contentID int64) (*models.Evaluation, error) {
	evaluation := &models.Evaluation{}
	var keyConcepts pq.StringArray

	err := er.db.QueryRowContext(ctx,
		`SELECT id, content_id, task_id, innovation_score, depth_score, decision, reasoning,
		        tldr, key_concepts, evaluated_at, evaluator_version, created_at, updated_at
		 FROM evaluation WHERE content_id = $1`,
		contentID,
	).Scan(&evaluation.ID, &evaluation.ContentID, &evaluation.TaskID, &evaluation.InnovationScore,
		&evaluation.DepthScore, &evaluation.Decision, &evaluation.Reasoning, &evaluation.TLDR,
		&keyConcepts, &evaluation.EvaluatedAt, &evaluation.EvaluatorVersion, &evaluation.CreatedAt, &evaluation.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	evaluation.KeyConcepts = keyConcepts
	return evaluation, nil
}

// GetByTaskID retrieves evaluation by task ID
func (er *EvaluationRepository) GetByTaskID(ctx context.Context, taskID uuid.UUID) (*models.Evaluation, error) {
	evaluation := &models.Evaluation{}
	var keyConcepts pq.StringArray

	err := er.db.QueryRowContext(ctx,
		`SELECT id, content_id, task_id, innovation_score, depth_score, decision, reasoning,
		        tldr, key_concepts, evaluated_at, evaluator_version, created_at, updated_at
		 FROM evaluation WHERE task_id = $1`,
		taskID,
	).Scan(&evaluation.ID, &evaluation.ContentID, &evaluation.TaskID, &evaluation.InnovationScore,
		&evaluation.DepthScore, &evaluation.Decision, &evaluation.Reasoning, &evaluation.TLDR,
		&keyConcepts, &evaluation.EvaluatedAt, &evaluation.EvaluatorVersion, &evaluation.CreatedAt, &evaluation.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	evaluation.KeyConcepts = keyConcepts
	return evaluation, nil
}

// ListByDecision retrieves evaluations by decision
func (er *EvaluationRepository) ListByDecision(ctx context.Context, decision string, limit, offset int) ([]*models.Evaluation, error) {
	rows, err := er.db.QueryContext(ctx,
		`SELECT id, content_id, task_id, innovation_score, depth_score, decision, reasoning,
		        tldr, key_concepts, evaluated_at, evaluator_version, created_at, updated_at
		 FROM evaluation WHERE decision = $1
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		decision, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var evaluations []*models.Evaluation
	for rows.Next() {
		evaluation := &models.Evaluation{}
		var keyConcepts pq.StringArray

		err := rows.Scan(&evaluation.ID, &evaluation.ContentID, &evaluation.TaskID, &evaluation.InnovationScore,
			&evaluation.DepthScore, &evaluation.Decision, &evaluation.Reasoning, &evaluation.TLDR,
			&keyConcepts, &evaluation.EvaluatedAt, &evaluation.EvaluatorVersion, &evaluation.CreatedAt, &evaluation.UpdatedAt)
		if err != nil {
			return nil, err
		}

		evaluation.KeyConcepts = keyConcepts
		evaluations = append(evaluations, evaluation)
	}

	return evaluations, rows.Err()
}

// ListHighScores retrieves evaluations with high scores
func (er *EvaluationRepository) ListHighScores(ctx context.Context, minInnovation, minDepth, limit, offset int) ([]*models.Evaluation, error) {
	rows, err := er.db.QueryContext(ctx,
		`SELECT id, content_id, task_id, innovation_score, depth_score, decision, reasoning,
		        tldr, key_concepts, evaluated_at, evaluator_version, created_at, updated_at
		 FROM evaluation WHERE innovation_score >= $1 AND depth_score >= $2
		 ORDER BY (innovation_score + depth_score) DESC LIMIT $3 OFFSET $4`,
		minInnovation, minDepth, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var evaluations []*models.Evaluation
	for rows.Next() {
		evaluation := &models.Evaluation{}
		var keyConcepts pq.StringArray

		err := rows.Scan(&evaluation.ID, &evaluation.ContentID, &evaluation.TaskID, &evaluation.InnovationScore,
			&evaluation.DepthScore, &evaluation.Decision, &evaluation.Reasoning, &evaluation.TLDR,
			&keyConcepts, &evaluation.EvaluatedAt, &evaluation.EvaluatorVersion, &evaluation.CreatedAt, &evaluation.UpdatedAt)
		if err != nil {
			return nil, err
		}

		evaluation.KeyConcepts = keyConcepts
		evaluations = append(evaluations, evaluation)
	}

	return evaluations, rows.Err()
}
