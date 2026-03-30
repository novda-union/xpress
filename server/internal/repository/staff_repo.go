package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xpressgo/server/internal/model"
)

type StaffRepo struct {
	db *pgxpool.Pool
}

type StaffListItem struct {
	model.Staff
	BranchName *string `json:"branch_name,omitempty"`
}

type StaffGroup struct {
	BranchID   *string         `json:"branch_id,omitempty"`
	BranchName string          `json:"branch_name"`
	Staff      []StaffListItem `json:"staff"`
}

func NewStaffRepo(db *pgxpool.Pool) *StaffRepo {
	return &StaffRepo{db: db}
}

func (r *StaffRepo) GetByStoreAndCode(ctx context.Context, storeID, staffCode string) (*model.Staff, error) {
	s := &model.Staff{}
	err := r.db.QueryRow(ctx, `
		SELECT id, store_id, branch_id, staff_code, name, password_hash, role, is_active, created_at
		FROM store_staff WHERE store_id = $1 AND staff_code = $2 AND is_active = true
	`, storeID, staffCode).Scan(
		&s.ID, &s.StoreID, &s.BranchID, &s.StaffCode, &s.Name, &s.PasswordHash, &s.Role, &s.IsActive, &s.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *StaffRepo) GetByID(ctx context.Context, id, storeID string) (*model.Staff, error) {
	s := &model.Staff{}
	err := r.db.QueryRow(ctx, `
		SELECT id, store_id, branch_id, staff_code, name, password_hash, role, is_active, created_at
		FROM store_staff
		WHERE id = $1 AND store_id = $2
	`, id, storeID).Scan(
		&s.ID, &s.StoreID, &s.BranchID, &s.StaffCode, &s.Name, &s.PasswordHash, &s.Role, &s.IsActive, &s.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *StaffRepo) ListByScope(ctx context.Context, storeID string, branchID *string) ([]StaffGroup, error) {
	query := `
		SELECT
			ss.id, ss.store_id, ss.branch_id, ss.staff_code, ss.name, ss.password_hash, ss.role, ss.is_active, ss.created_at,
			b.name
		FROM store_staff ss
		LEFT JOIN branches b ON b.id = ss.branch_id
		WHERE ss.store_id = $1 AND ss.is_active = true
	`
	args := []any{storeID}
	if branchID != nil && *branchID != "" {
		query += ` AND ss.branch_id = $2 AND ss.role = 'barista'`
		args = append(args, *branchID)
	}
	query += ` ORDER BY COALESCE(b.name, 'Store'), ss.role, ss.name`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groupOrder := make([]string, 0)
	groups := make(map[string]*StaffGroup)

	for rows.Next() {
		var item StaffListItem
		if err := rows.Scan(
			&item.ID, &item.StoreID, &item.BranchID, &item.StaffCode, &item.Name, &item.PasswordHash, &item.Role, &item.IsActive, &item.CreatedAt,
			&item.BranchName,
		); err != nil {
			return nil, err
		}

		groupKey := "store"
		groupName := "Store"
		var groupBranchID *string
		if item.BranchID != nil {
			groupKey = *item.BranchID
			groupBranchID = item.BranchID
			if item.BranchName != nil && *item.BranchName != "" {
				groupName = *item.BranchName
			}
		}

		group, ok := groups[groupKey]
		if !ok {
			group = &StaffGroup{
				BranchID:   groupBranchID,
				BranchName: groupName,
				Staff:      []StaffListItem{},
			}
			groups[groupKey] = group
			groupOrder = append(groupOrder, groupKey)
		}
		group.Staff = append(group.Staff, item)
	}

	result := make([]StaffGroup, 0, len(groupOrder))
	for _, key := range groupOrder {
		result = append(result, *groups[key])
	}
	if result == nil {
		result = []StaffGroup{}
	}
	return result, nil
}

func (r *StaffRepo) Create(ctx context.Context, staff *model.Staff) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO store_staff (store_id, branch_id, staff_code, name, password_hash, role, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`, staff.StoreID, staff.BranchID, staff.StaffCode, staff.Name, staff.PasswordHash, staff.Role, staff.IsActive).Scan(
		&staff.ID, &staff.CreatedAt,
	)
}

func (r *StaffRepo) Update(ctx context.Context, staff *model.Staff) error {
	_, err := r.db.Exec(ctx, `
		UPDATE store_staff
		SET
			branch_id = $3,
			staff_code = $4,
			name = $5,
			password_hash = CASE WHEN $6 = '' THEN password_hash ELSE $6 END,
			role = $7,
			is_active = $8
		WHERE id = $1 AND store_id = $2
	`, staff.ID, staff.StoreID, staff.BranchID, staff.StaffCode, staff.Name, staff.PasswordHash, staff.Role, staff.IsActive)
	return err
}

func (r *StaffRepo) Deactivate(ctx context.Context, id, storeID string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE store_staff
		SET is_active = false
		WHERE id = $1 AND store_id = $2
	`, id, storeID)
	return err
}
