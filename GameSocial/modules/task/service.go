// task 模块负责任务定义、进度与打卡等业务能力。
package task

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// TaskDef 对应数据库 task_def 表的数据结构。
type TaskDef struct {
	ID           uint64          `json:"id"`
	TaskCode     string          `json:"taskCode"`
	Name         string          `json:"name"`
	PeriodType   string          `json:"periodType"`
	TargetCount  int             `json:"targetCount"`
	RewardPoints int64           `json:"rewardPoints"`
	Status       int             `json:"status"`
	RuleJSON     json.RawMessage `json:"ruleJson,omitempty"`
	CreatedAt    time.Time       `json:"createdAt"`
}

// CreateTaskDefRequest 创建任务定义入参。
type CreateTaskDefRequest struct {
	TaskCode     string          `json:"taskCode"`
	Name         string          `json:"name"`
	PeriodType   string          `json:"periodType"`
	TargetCount  int             `json:"targetCount"`
	RewardPoints int64           `json:"rewardPoints"`
	Status       int             `json:"status"`
	RuleJSON     json.RawMessage `json:"ruleJson"`
}

// UpdateTaskDefRequest 更新任务定义入参。
type UpdateTaskDefRequest struct {
	Name         string          `json:"name"`
	PeriodType   string          `json:"periodType"`
	TargetCount  int             `json:"targetCount"`
	RewardPoints int64           `json:"rewardPoints"`
	Status       int             `json:"status"`
	RuleJSON     json.RawMessage `json:"ruleJson"`
}

// ListTaskDefRequest 列表查询入参。
type ListTaskDefRequest struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
	Status int `json:"status"`
}

// Service 定义 task 模块对外提供的业务接口（任务定义 CRUD）。
type Service interface {
	CreateTaskDef(ctx context.Context, req CreateTaskDefRequest) (TaskDef, error)
	UpdateTaskDef(ctx context.Context, id uint64, req UpdateTaskDefRequest) (TaskDef, error)
	DeleteTaskDef(ctx context.Context, id uint64) error
	GetTaskDef(ctx context.Context, id uint64) (TaskDef, error)
	ListTaskDef(ctx context.Context, req ListTaskDefRequest) ([]TaskDef, error)
}

type service struct {
	db *sql.DB
}

// NewService 创建 task 模块服务。
func NewService(db *sql.DB) Service {
	return &service{db: db}
}

func (s *service) CreateTaskDef(ctx context.Context, req CreateTaskDefRequest) (TaskDef, error) {
	// 1) 基础校验。
	if s.db == nil {
		return TaskDef{}, errors.New("database disabled")
	}
	if req.TaskCode == "" {
		return TaskDef{}, errors.New("task_code is empty")
	}
	if req.Name == "" {
		return TaskDef{}, errors.New("name is empty")
	}
	if req.PeriodType == "" {
		return TaskDef{}, errors.New("period_type is empty")
	}
	if req.TargetCount <= 0 {
		return TaskDef{}, errors.New("target_count must be > 0")
	}
	if req.RewardPoints < 0 {
		return TaskDef{}, errors.New("reward_points must be >= 0")
	}
	if req.Status == 0 {
		req.Status = 1
	}

	// 2) 写入 task_def 表：rule_json 允许为空。
	var rule any
	if len(req.RuleJSON) != 0 {
		rule = string(req.RuleJSON)
	}
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO task_def (task_code, name, period_type, target_count, reward_points, status, rule_json, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, NOW())
	`, req.TaskCode, req.Name, req.PeriodType, req.TargetCount, req.RewardPoints, req.Status, rule)
	if err != nil {
		return TaskDef{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return TaskDef{}, err
	}
	return s.GetTaskDef(ctx, uint64(id))
}

func (s *service) UpdateTaskDef(ctx context.Context, id uint64, req UpdateTaskDefRequest) (TaskDef, error) {
	// 1) 基础校验。
	if s.db == nil {
		return TaskDef{}, errors.New("database disabled")
	}
	if id == 0 {
		return TaskDef{}, errors.New("invalid id")
	}
	if req.Name == "" {
		return TaskDef{}, errors.New("name is empty")
	}
	if req.PeriodType == "" {
		return TaskDef{}, errors.New("period_type is empty")
	}
	if req.TargetCount <= 0 {
		return TaskDef{}, errors.New("target_count must be > 0")
	}
	if req.RewardPoints < 0 {
		return TaskDef{}, errors.New("reward_points must be >= 0")
	}
	if req.Status == 0 {
		req.Status = 1
	}

	// 2) 更新 task_def 表：rule_json 同样允许为空。
	var rule any
	if len(req.RuleJSON) != 0 {
		rule = string(req.RuleJSON)
	}
	result, err := s.db.ExecContext(ctx, `
		UPDATE task_def
		SET name = ?, period_type = ?, target_count = ?, reward_points = ?, status = ?, rule_json = ?
		WHERE id = ?
	`, req.Name, req.PeriodType, req.TargetCount, req.RewardPoints, req.Status, rule, id)
	if err != nil {
		return TaskDef{}, err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return TaskDef{}, fmt.Errorf("task_def not found")
	}
	return s.GetTaskDef(ctx, id)
}

func (s *service) DeleteTaskDef(ctx context.Context, id uint64) error {
	// 1) 基础校验。
	if s.db == nil {
		return errors.New("database disabled")
	}
	if id == 0 {
		return errors.New("invalid id")
	}

	// 2) 软删除：把 status 置为 0，保留历史进度引用。
	result, err := s.db.ExecContext(ctx, `
		UPDATE task_def
		SET status = 0
		WHERE id = ? AND status <> 0
	`, id)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("task_def not found")
	}
	return nil
}

func (s *service) GetTaskDef(ctx context.Context, id uint64) (TaskDef, error) {
	// 1) 基础校验。
	if s.db == nil {
		return TaskDef{}, errors.New("database disabled")
	}
	if id == 0 {
		return TaskDef{}, errors.New("invalid id")
	}

	// 2) 查询单条：rule_json 是 JSON 列，扫描为 []byte 再转 RawMessage。
	var td TaskDef
	var ruleBytes []byte
	row := s.db.QueryRowContext(ctx, `
		SELECT id, task_code, name, period_type, target_count, reward_points, status, rule_json, created_at
		FROM task_def
		WHERE id = ?
		LIMIT 1
	`, id)
	if err := row.Scan(&td.ID, &td.TaskCode, &td.Name, &td.PeriodType, &td.TargetCount, &td.RewardPoints, &td.Status, &ruleBytes, &td.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return TaskDef{}, fmt.Errorf("task_def not found")
		}
		return TaskDef{}, err
	}
	if len(ruleBytes) != 0 {
		td.RuleJSON = json.RawMessage(ruleBytes)
	}
	return td, nil
}

func (s *service) ListTaskDef(ctx context.Context, req ListTaskDefRequest) ([]TaskDef, error) {
	// 1) 基础校验与分页兜底。
	if s.db == nil {
		return nil, errors.New("database disabled")
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 200 {
		req.Limit = 200
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// 2) 组装筛选条件：status=0 代表不过滤，但默认排除已删除。
	where := ""
	args := make([]any, 0, 3)
	if req.Status != 0 {
		where = "WHERE status = ?"
		args = append(args, req.Status)
	} else {
		where = "WHERE status <> 0"
	}
	args = append(args, req.Limit, req.Offset)

	// 3) 查询列表：按 id 倒序。
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, task_code, name, period_type, target_count, reward_points, status, IFNULL(rule_json, JSON_OBJECT()), created_at
		FROM task_def
		`+where+`
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]TaskDef, 0, req.Limit)
	for rows.Next() {
		var td TaskDef
		var ruleBytes []byte
		if err := rows.Scan(&td.ID, &td.TaskCode, &td.Name, &td.PeriodType, &td.TargetCount, &td.RewardPoints, &td.Status, &ruleBytes, &td.CreatedAt); err != nil {
			return nil, err
		}
		if len(ruleBytes) != 0 {
			td.RuleJSON = json.RawMessage(ruleBytes)
		}
		out = append(out, td)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
