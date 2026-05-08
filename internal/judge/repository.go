package judge

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository handles database operations for jobs.
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new judge repository.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// CreateJob inserts a new JobRecord.
func (r *Repository) CreateJob(job *JobRecord) error {
	return r.db.Create(job).Error
}

// UpdateJobResult updates the status and result fields of an existing job.
func (r *Repository) UpdateJobResult(jobID uuid.UUID, dto UpdateJobDTO) error {
	return r.db.Model(&JobRecord{}).Where("id = ?", jobID).Updates(dto).Error
}

// GetJobByID retrieves a job by its ID.
func (r *Repository) GetJobByID(jobID uuid.UUID) (*JobRecord, error) {
	var job JobRecord
	if err := r.db.First(&job, "id = ?", jobID).Error; err != nil {
		return nil, err
	}
	return &job, nil
}
