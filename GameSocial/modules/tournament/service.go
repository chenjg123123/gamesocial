// tournament 模块负责赛事发布、维护与查询等业务能力。
package tournament

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Tournament 对应数据库 tournament 表的数据结构。
type Tournament struct {
	ID             uint64    `json:"id"`
	Title          string    `json:"title"`
	Content        string    `json:"content,omitempty"`
	CoverURL       string    `json:"coverUrl,omitempty"`
	ImageURLs      []string  `json:"imageUrls,omitempty"`
	StartAt        time.Time `json:"startAt"`
	EndAt          time.Time `json:"endAt"`
	Status         string    `json:"status"`
	CreatedByAdmin uint64    `json:"createdByAdminId"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// CreateTournamentRequest 创建赛事入参。
type CreateTournamentRequest struct {
	Title          string    `json:"title"`
	Content        string    `json:"content"`
	CoverURL       string    `json:"coverUrl"`
	ImageURLs      []string  `json:"imageUrls,omitempty"`
	StartAt        time.Time `json:"startAt"`
	EndAt          time.Time `json:"endAt"`
	Status         string    `json:"status"`
	CreatedByAdmin uint64    `json:"createdByAdminId"`
}

// UpdateTournamentRequest 更新赛事入参。
type UpdateTournamentRequest struct {
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CoverURL  string    `json:"coverUrl"`
	ImageURLs []string  `json:"imageUrls,omitempty"`
	StartAt   time.Time `json:"startAt"`
	EndAt     time.Time `json:"endAt"`
	Status    string    `json:"status"`
}

// ListTournamentRequest 列表查询入参。
type ListTournamentRequest struct {
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
	Status string `json:"status"`
}

// ListJoinedTournamentRequest 查询“我已报名赛事”列表的入参。
// 说明：
// - Status 用于按赛事状态过滤（如 PUBLISHED/FINISHED）；为空时默认排除 CANCELED
// - Keyword 用于按赛事标题模糊搜索（LIKE %keyword%）
// - Offset/Limit 用于分页
type ListJoinedTournamentRequest struct {
	Offset  int    `json:"offset"`
	Limit   int    `json:"limit"`
	Status  string `json:"status"`
	Keyword string `json:"keyword"`
}

// JoinedTournament 是“我已报名赛事”列表的返回项。
// 在 Tournament 基础字段上，额外携带报名信息（JoinStatus/JoinedAt）。
type JoinedTournament struct {
	Tournament
	JoinStatus string    `json:"joinStatus"`
	JoinedAt   time.Time `json:"joinedAt"`
}

// TournamentParticipant 对应 tournament_participant 表的数据结构。
type TournamentParticipant struct {
	ID           uint64    `json:"id"`
	TournamentID uint64    `json:"tournamentId"`
	UserID       uint64    `json:"userId"`
	JoinStatus   string    `json:"joinStatus"`
	JoinedAt     time.Time `json:"joinedAt"`
}

// TournamentResultItem 是赛事成绩列表中的单条数据（排行榜项）。
type TournamentResultItem struct {
	UserID    uint64 `json:"userId"`
	RankNo    int    `json:"rankNo"`
	Score     int    `json:"score"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatarUrl"`
}

// TournamentResults 是赛事成绩查询返回结构。
type TournamentResults struct {
	Items []TournamentResultItem `json:"items"`
	My    *TournamentResultItem  `json:"my,omitempty"`
}

// Service 定义 tournament 模块对外提供的业务接口（赛事 CRUD）。
type Service interface {
	Create(ctx context.Context, req CreateTournamentRequest) (Tournament, error)
	Update(ctx context.Context, id uint64, req UpdateTournamentRequest) (Tournament, error)
	Delete(ctx context.Context, id uint64) error
	Get(ctx context.Context, id uint64) (Tournament, error)
	List(ctx context.Context, req ListTournamentRequest) ([]Tournament, error)
	ListJoined(ctx context.Context, userID uint64, req ListJoinedTournamentRequest) ([]JoinedTournament, error)
	Join(ctx context.Context, tournamentID, userID uint64) error
	Cancel(ctx context.Context, tournamentID, userID uint64) error
	GetResults(ctx context.Context, tournamentID, userID uint64, offset, limit int) (TournamentResults, error)
}

type service struct {
	db *sql.DB
}

// NewService 创建 tournament 模块服务。
func NewService(db *sql.DB) Service {
	return &service{db: db}
}

// Create 创建赛事并返回创建后的赛事详情。
func (s *service) Create(ctx context.Context, req CreateTournamentRequest) (Tournament, error) {
	// 1) 基础校验。
	if s.db == nil {
		return Tournament{}, errors.New("database disabled")
	}
	if req.Title == "" {
		return Tournament{}, errors.New("title is empty")
	}
	if req.StartAt.IsZero() || req.EndAt.IsZero() {
		return Tournament{}, errors.New("start_at/end_at is empty")
	}
	if req.EndAt.Before(req.StartAt) {
		return Tournament{}, errors.New("end_at must be >= start_at")
	}
	if req.Status == "" {
		req.Status = "DRAFT"
	}
	if req.CreatedByAdmin == 0 {
		req.CreatedByAdmin = 1
	}

	if len(req.ImageURLs) == 0 && req.CoverURL != "" {
		req.ImageURLs = []string{req.CoverURL}
	}
	if len(req.ImageURLs) > 0 {
		req.CoverURL = req.ImageURLs[0]
	}

	imageURLsJSON := ""
	if len(req.ImageURLs) > 0 {
		b, err := json.Marshal(req.ImageURLs)
		if err != nil {
			return Tournament{}, errors.New("invalid imageUrls")
		}
		imageURLsJSON = string(b)
	}

	// 2) 写入 tournament 表，并返回创建后的详情。
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO tournament (title, content, cover_url, image_urls_json, start_at, end_at, status, created_by_admin_id, created_at, updated_at)
		VALUES (?, NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), ?, ?, ?, ?, NOW(), NOW())
	`, req.Title, req.Content, req.CoverURL, imageURLsJSON, req.StartAt, req.EndAt, req.Status, req.CreatedByAdmin)
	if err != nil && isUnknownColumn(err, "image_urls_json") {
		res, err = s.db.ExecContext(ctx, `
			INSERT INTO tournament (title, content, cover_url, start_at, end_at, status, created_by_admin_id, created_at, updated_at)
			VALUES (?, NULLIF(?, ''), NULLIF(?, ''), ?, ?, ?, ?, NOW(), NOW())
		`, req.Title, req.Content, req.CoverURL, req.StartAt, req.EndAt, req.Status, req.CreatedByAdmin)
	}
	if err != nil {
		return Tournament{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Tournament{}, err
	}
	return s.Get(ctx, uint64(id))
}

// Update 更新赛事并返回更新后的赛事详情。
func (s *service) Update(ctx context.Context, id uint64, req UpdateTournamentRequest) (Tournament, error) {
	// 1) 基础校验。
	if s.db == nil {
		return Tournament{}, errors.New("database disabled")
	}
	if id == 0 {
		return Tournament{}, errors.New("invalid id")
	}
	if req.Title == "" {
		return Tournament{}, errors.New("title is empty")
	}
	if req.StartAt.IsZero() || req.EndAt.IsZero() {
		return Tournament{}, errors.New("start_at/end_at is empty")
	}
	if req.EndAt.Before(req.StartAt) {
		return Tournament{}, errors.New("end_at must be >= start_at")
	}
	if req.Status == "" {
		req.Status = "DRAFT"
	}

	if len(req.ImageURLs) == 0 && req.CoverURL != "" {
		req.ImageURLs = []string{req.CoverURL}
	}
	if len(req.ImageURLs) > 0 {
		req.CoverURL = req.ImageURLs[0]
	}

	imageURLsJSON := ""
	if len(req.ImageURLs) > 0 {
		b, err := json.Marshal(req.ImageURLs)
		if err != nil {
			return Tournament{}, errors.New("invalid imageUrls")
		}
		imageURLsJSON = string(b)
	}

	// 2) 更新可变字段，并刷新 updated_at。
	result, err := s.db.ExecContext(ctx, `
		UPDATE tournament
		SET title = ?, content = NULLIF(?, ''), cover_url = NULLIF(?, ''), image_urls_json = NULLIF(?, ''), start_at = ?, end_at = ?, status = ?, updated_at = NOW()
		WHERE id = ?
	`, req.Title, req.Content, req.CoverURL, imageURLsJSON, req.StartAt, req.EndAt, req.Status, id)
	if err != nil && isUnknownColumn(err, "image_urls_json") {
		result, err = s.db.ExecContext(ctx, `
			UPDATE tournament
			SET title = ?, content = NULLIF(?, ''), cover_url = NULLIF(?, ''), start_at = ?, end_at = ?, status = ?, updated_at = NOW()
			WHERE id = ?
		`, req.Title, req.Content, req.CoverURL, req.StartAt, req.EndAt, req.Status, id)
	}
	if err != nil {
		return Tournament{}, err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return Tournament{}, fmt.Errorf("tournament not found")
	}
	return s.Get(ctx, id)
}

// Delete 软删除赛事（status=CANCELED）。
func (s *service) Delete(ctx context.Context, id uint64) error {
	// 1) 基础校验。
	if s.db == nil {
		return errors.New("database disabled")
	}
	if id == 0 {
		return errors.New("invalid id")
	}

	// 2) 软删除：把 status 标记为 CANCELED，保留历史报名/成绩/发奖的引用。
	result, err := s.db.ExecContext(ctx, `
		UPDATE tournament
		SET status = 'CANCELED', updated_at = NOW()
		WHERE id = ? AND status <> 'CANCELED'
	`, id)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("tournament not found")
	}
	return nil
}

// Get 获取赛事详情。
func (s *service) Get(ctx context.Context, id uint64) (Tournament, error) {
	// 1) 基础校验。
	if s.db == nil {
		return Tournament{}, errors.New("database disabled")
	}
	if id == 0 {
		return Tournament{}, errors.New("invalid id")
	}

	// 2) 查询单条记录：content/cover_url 可空。
	var t Tournament
	var content, cover, imageURLs sql.NullString
	row := s.db.QueryRowContext(ctx, `
		SELECT id, title, content, cover_url, image_urls_json, start_at, end_at, status, created_by_admin_id, created_at, updated_at
		FROM tournament
		WHERE id = ?
		LIMIT 1
	`, id)
	if err := row.Scan(&t.ID, &t.Title, &content, &cover, &imageURLs, &t.StartAt, &t.EndAt, &t.Status, &t.CreatedByAdmin, &t.CreatedAt, &t.UpdatedAt); err != nil {
		if isUnknownColumn(err, "image_urls_json") {
			row2 := s.db.QueryRowContext(ctx, `
				SELECT id, title, content, cover_url, start_at, end_at, status, created_by_admin_id, created_at, updated_at
				FROM tournament
				WHERE id = ?
				LIMIT 1
			`, id)
			if err2 := row2.Scan(&t.ID, &t.Title, &content, &cover, &t.StartAt, &t.EndAt, &t.Status, &t.CreatedByAdmin, &t.CreatedAt, &t.UpdatedAt); err2 != nil {
				if err2 == sql.ErrNoRows {
					return Tournament{}, fmt.Errorf("tournament not found")
				}
				return Tournament{}, err2
			}
			t.Content = content.String
			t.CoverURL = cover.String
			if t.CoverURL != "" {
				t.ImageURLs = []string{t.CoverURL}
			}
			return t, nil
		}
		if err == sql.ErrNoRows {
			return Tournament{}, fmt.Errorf("tournament not found")
		}
		return Tournament{}, err
	}
	t.Content = content.String
	t.CoverURL = cover.String
	if imageURLs.Valid && strings.TrimSpace(imageURLs.String) != "" {
		var list []string
		if err := json.Unmarshal([]byte(imageURLs.String), &list); err == nil {
			t.ImageURLs = list
		}
	}
	if len(t.ImageURLs) == 0 && t.CoverURL != "" {
		t.ImageURLs = []string{t.CoverURL}
	}
	return t, nil
}

// List 获取赛事列表。
func (s *service) List(ctx context.Context, req ListTournamentRequest) ([]Tournament, error) {
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

	// 2) 组装筛选条件：未指定 status 则默认排除 CANCELED。
	where := ""
	args := make([]any, 0, 3)
	if req.Status != "" {
		where = "WHERE status = ?"
		args = append(args, req.Status)
	} else {
		where = "WHERE status <> 'CANCELED'"
	}
	args = append(args, req.Limit, req.Offset)

	// 3) 查询列表：按 start_at 倒序，便于后台优先看到最近赛事。
	withImageURLsJSON := true
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, title, IFNULL(content, ''), IFNULL(cover_url, ''), IFNULL(image_urls_json, ''), start_at, end_at, status, created_by_admin_id, created_at, updated_at
		FROM tournament
		`+where+`
		ORDER BY start_at DESC, id DESC
		LIMIT ? OFFSET ?
	`, args...)
	if err != nil && isUnknownColumn(err, "image_urls_json") {
		withImageURLsJSON = false
		rows, err = s.db.QueryContext(ctx, `
			SELECT id, title, IFNULL(content, ''), IFNULL(cover_url, ''), start_at, end_at, status, created_by_admin_id, created_at, updated_at
			FROM tournament
			`+where+`
			ORDER BY start_at DESC, id DESC
			LIMIT ? OFFSET ?
		`, args...)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Tournament, 0, req.Limit)
	for rows.Next() {
		var t Tournament
		var imageURLsJSON string
		if withImageURLsJSON {
			if err := rows.Scan(&t.ID, &t.Title, &t.Content, &t.CoverURL, &imageURLsJSON, &t.StartAt, &t.EndAt, &t.Status, &t.CreatedByAdmin, &t.CreatedAt, &t.UpdatedAt); err != nil {
				return nil, err
			}
		} else {
			if err := rows.Scan(&t.ID, &t.Title, &t.Content, &t.CoverURL, &t.StartAt, &t.EndAt, &t.Status, &t.CreatedByAdmin, &t.CreatedAt, &t.UpdatedAt); err != nil {
				return nil, err
			}
		}
		if strings.TrimSpace(imageURLsJSON) != "" {
			var list []string
			if err := json.Unmarshal([]byte(imageURLsJSON), &list); err == nil {
				t.ImageURLs = list
			}
		}
		if len(t.ImageURLs) == 0 && t.CoverURL != "" {
			t.ImageURLs = []string{t.CoverURL}
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// ListJoined 查询指定用户已报名（JOINED）的赛事列表。
// 排序：按报名时间 joined_at 倒序，其次按 participant.id 倒序。
func (s *service) ListJoined(ctx context.Context, userID uint64, req ListJoinedTournamentRequest) ([]JoinedTournament, error) {
	if s.db == nil {
		return nil, errors.New("database disabled")
	}
	if userID == 0 {
		return nil, errors.New("invalid user id")
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

	// 仅返回当前用户仍处于 JOINED 的报名记录。
	// 这里 join_status 的过滤放在 participant 上，status 的过滤放在 tournament 上。
	where := "WHERE p.user_id = ? AND p.join_status = 'JOINED'"
	args := make([]any, 0, 5)
	args = append(args, userID)
	if req.Status != "" {
		where += " AND t.status = ?"
		args = append(args, req.Status)
	} else {
		// 默认排除已取消的赛事，避免列表出现管理员删除/取消的内容。
		where += " AND t.status <> 'CANCELED'"
	}
	if req.Keyword != "" {
		where += " AND t.title LIKE ?"
		args = append(args, "%"+req.Keyword+"%")
	}
	args = append(args, req.Limit, req.Offset)

	withImageURLsJSON := true
	rows, err := s.db.QueryContext(ctx, `
		SELECT
			t.id, t.title, IFNULL(t.content, ''), IFNULL(t.cover_url, ''), IFNULL(t.image_urls_json, ''), t.start_at, t.end_at, t.status, t.created_by_admin_id, t.created_at, t.updated_at,
			p.join_status, p.joined_at
		FROM tournament_participant p
		INNER JOIN tournament t ON t.id = p.tournament_id
		`+where+`
		ORDER BY p.joined_at DESC, p.id DESC
		LIMIT ? OFFSET ?
	`, args...)
	if err != nil && isUnknownColumn(err, "image_urls_json") {
		withImageURLsJSON = false
		rows, err = s.db.QueryContext(ctx, `
			SELECT
				t.id, t.title, IFNULL(t.content, ''), IFNULL(t.cover_url, ''), t.start_at, t.end_at, t.status, t.created_by_admin_id, t.created_at, t.updated_at,
				p.join_status, p.joined_at
			FROM tournament_participant p
			INNER JOIN tournament t ON t.id = p.tournament_id
			`+where+`
			ORDER BY p.joined_at DESC, p.id DESC
			LIMIT ? OFFSET ?
		`, args...)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]JoinedTournament, 0, req.Limit)
	for rows.Next() {
		var it JoinedTournament
		var imageURLsJSON string
		if withImageURLsJSON {
			if err := rows.Scan(
				&it.ID, &it.Title, &it.Content, &it.CoverURL, &imageURLsJSON, &it.StartAt, &it.EndAt, &it.Status, &it.CreatedByAdmin, &it.CreatedAt, &it.UpdatedAt,
				&it.JoinStatus, &it.JoinedAt,
			); err != nil {
				return nil, err
			}
		} else {
			if err := rows.Scan(
				&it.ID, &it.Title, &it.Content, &it.CoverURL, &it.StartAt, &it.EndAt, &it.Status, &it.CreatedByAdmin, &it.CreatedAt, &it.UpdatedAt,
				&it.JoinStatus, &it.JoinedAt,
			); err != nil {
				return nil, err
			}
		}
		if strings.TrimSpace(imageURLsJSON) != "" {
			var list []string
			if err := json.Unmarshal([]byte(imageURLsJSON), &list); err == nil {
				it.ImageURLs = list
			}
		}
		if len(it.ImageURLs) == 0 && it.CoverURL != "" {
			it.ImageURLs = []string{it.CoverURL}
		}
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func isUnknownColumn(err error, column string) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "Error 1054") && strings.Contains(s, "Unknown column") && strings.Contains(s, column)
}

func (s *service) Join(ctx context.Context, tournamentID, userID uint64) error {
	if s.db == nil {
		return errors.New("database disabled")
	}
	if tournamentID == 0 {
		return errors.New("invalid tournament id")
	}
	if userID == 0 {
		return errors.New("invalid user id")
	}

	t, err := s.Get(ctx, tournamentID)
	if err != nil {
		return err
	}
	if t.Status != "PUBLISHED" {
		return fmt.Errorf("tournament not published")
	}
	if !t.EndAt.IsZero() && time.Now().After(t.EndAt) {
		return fmt.Errorf("tournament ended")
	}
	// 先检查用户是否已经参加当前赛事。
	var count int
	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM tournament_participant WHERE tournament_id = ? AND user_id = ? AND join_status <> 'CANCELED'
	`, tournamentID, userID).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("请勿重复报名")
	}

	// 2) 插入或更新报名记录：join_status=JOINED。
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO tournament_participant (tournament_id, user_id, join_status, joined_at)
		VALUES (?, ?, 'JOINED', NOW())
		ON DUPLICATE KEY UPDATE join_status = 'JOINED', joined_at = NOW()
	`, tournamentID, userID)
	return err
}

func (s *service) Cancel(ctx context.Context, tournamentID, userID uint64) error {
	if s.db == nil {
		return errors.New("database disabled")
	}
	if tournamentID == 0 {
		return errors.New("invalid tournament id")
	}
	if userID == 0 {
		return errors.New("invalid user id")
	}

	_, err := s.Get(ctx, tournamentID)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, `
		UPDATE tournament_participant
		SET join_status = 'CANCELED'
		WHERE tournament_id = ? AND user_id = ? AND join_status <> 'CANCELED'
	`, tournamentID, userID)
	if err != nil {
		return err
	}
	return nil
}

func (s *service) GetResults(ctx context.Context, tournamentID, userID uint64, offset, limit int) (TournamentResults, error) {
	if s.db == nil {
		return TournamentResults{}, errors.New("database disabled")
	}
	if tournamentID == 0 {
		return TournamentResults{}, errors.New("invalid tournament id")
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	_, err := s.Get(ctx, tournamentID)
	if err != nil {
		return TournamentResults{}, err
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT r.user_id, r.rank_no, r.score, IFNULL(u.nickname, ''), IFNULL(u.avatar_url, '')
		FROM tournament_result r
		LEFT JOIN user u ON u.id = r.user_id
		WHERE r.tournament_id = ?
		ORDER BY r.rank_no ASC, r.id ASC
		LIMIT ? OFFSET ?
	`, tournamentID, limit, offset)
	if err != nil {
		return TournamentResults{}, err
	}
	defer rows.Close()

	items := make([]TournamentResultItem, 0, limit)
	for rows.Next() {
		var it TournamentResultItem
		var score sql.NullInt64
		if err := rows.Scan(&it.UserID, &it.RankNo, &score, &it.Nickname, &it.AvatarURL); err != nil {
			return TournamentResults{}, err
		}
		if score.Valid {
			it.Score = int(score.Int64)
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return TournamentResults{}, err
	}

	var my *TournamentResultItem
	if userID != 0 {
		var it TournamentResultItem
		var score sql.NullInt64
		err := s.db.QueryRowContext(ctx, `
			SELECT r.user_id, r.rank_no, r.score, IFNULL(u.nickname, ''), IFNULL(u.avatar_url, '')
			FROM tournament_result r
			LEFT JOIN user u ON u.id = r.user_id
			WHERE r.tournament_id = ? AND r.user_id = ?
			LIMIT 1
		`, tournamentID, userID).Scan(&it.UserID, &it.RankNo, &score, &it.Nickname, &it.AvatarURL)
		if err != nil && err != sql.ErrNoRows {
			return TournamentResults{}, err
		}
		if err == nil {
			if score.Valid {
				it.Score = int(score.Int64)
			}
			my = &it
		}
	}

	return TournamentResults{
		Items: items,
		My:    my,
	}, nil
}
