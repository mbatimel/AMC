package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"

	internalModels "github.com/mbatimel/AMC/products/internal/models"
)

var ErrPromotionNotFound = errors.New("promotion not found")

func insertPromotionProducts(ctx context.Context, tx pgx.Tx, promotionID uuid.UUID, products []internalModels.PromotionProduct) error {
	for _, product := range products {
		if _, err := tx.Exec(ctx,
			`INSERT INTO promotion_products (promotion_id, product_id, min_qty) VALUES ($1, $2, $3)`,
			promotionID, product.ProductID, product.MinQty,
		); err != nil {
			return classifyWriteError(fmt.Errorf("insert promotion product: %w", err))
		}
	}
	return nil
}

func scanPromotion(row rowScanner) (internalModels.Promotion, error) {
	var promotion internalModels.Promotion
	err := row.Scan(
		&promotion.ID,
		&promotion.Name,
		&promotion.DiscountPercent,
		&promotion.StartsAt,
		&promotion.EndsAt,
		&promotion.CreatedAt,
		&promotion.UpdatedAt,
	)
	return promotion, err
}

func (s *Storage) promotionProducts(ctx context.Context, promotionIDs []uuid.UUID) (map[uuid.UUID][]internalModels.PromotionProduct, error) {
	result := make(map[uuid.UUID][]internalModels.PromotionProduct)
	if len(promotionIDs) == 0 {
		return result, nil
	}
	rows, err := s.pool.Query(ctx,
		`SELECT promotion_id, product_id, min_qty FROM promotion_products WHERE promotion_id = ANY($1) ORDER BY product_id`,
		promotionIDs,
	)
	if err != nil {
		return nil, fmt.Errorf("list promotion products: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var promotionID uuid.UUID
		var product internalModels.PromotionProduct
		if err = rows.Scan(&promotionID, &product.ProductID, &product.MinQty); err != nil {
			return nil, fmt.Errorf("scan promotion product: %w", err)
		}
		result[promotionID] = append(result[promotionID], product)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate promotion products: %w", err)
	}
	return result, nil
}

func (s *Storage) CreatePromotion(ctx context.Context, params internalModels.CreatePromotionParams) (internalModels.Promotion, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return internalModels.Promotion{}, fmt.Errorf("begin create promotion: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	promotion := internalModels.Promotion{
		Name:            params.Name,
		DiscountPercent: params.DiscountPercent,
		StartsAt:        params.StartsAt,
		EndsAt:          params.EndsAt,
		Products:        params.Products,
	}
	err = tx.QueryRow(ctx,
		`INSERT INTO promotions (name, discount_percent, starts_at, ends_at) VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at, updated_at`,
		params.Name, params.DiscountPercent, params.StartsAt, params.EndsAt,
	).Scan(&promotion.ID, &promotion.CreatedAt, &promotion.UpdatedAt)
	if err != nil {
		return internalModels.Promotion{}, fmt.Errorf("insert promotion: %w", err)
	}
	if err = insertPromotionProducts(ctx, tx, promotion.ID, params.Products); err != nil {
		return internalModels.Promotion{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return internalModels.Promotion{}, fmt.Errorf("commit create promotion: %w", err)
	}
	return promotion, nil
}

func (s *Storage) GetPromotionByID(ctx context.Context, promotionID uuid.UUID) (internalModels.Promotion, error) {
	promotion, err := scanPromotion(s.pool.QueryRow(ctx,
		`SELECT id, name, discount_percent, starts_at, ends_at, created_at, updated_at FROM promotions WHERE id = $1`,
		promotionID,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return internalModels.Promotion{}, ErrPromotionNotFound
	}
	if err != nil {
		return internalModels.Promotion{}, fmt.Errorf("get promotion by id: %w", err)
	}
	products, err := s.promotionProducts(ctx, []uuid.UUID{promotionID})
	if err != nil {
		return internalModels.Promotion{}, err
	}
	promotion.Products = products[promotionID]
	return promotion, nil
}

func (s *Storage) ListPromotions(ctx context.Context, limit int, offset int) ([]internalModels.Promotion, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, name, discount_percent, starts_at, ends_at, created_at, updated_at
		 FROM promotions ORDER BY starts_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list promotions: %w", err)
	}
	defer rows.Close()

	promotions := make([]internalModels.Promotion, 0)
	ids := make([]uuid.UUID, 0)
	for rows.Next() {
		promotion, scanErr := scanPromotion(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan promotion: %w", scanErr)
		}
		promotions = append(promotions, promotion)
		ids = append(ids, promotion.ID)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate promotions: %w", err)
	}

	products, err := s.promotionProducts(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range promotions {
		promotions[i].Products = products[promotions[i].ID]
	}
	return promotions, nil
}

func (s *Storage) CountPromotions(ctx context.Context) (int, error) {
	var total int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM promotions`).Scan(&total); err != nil {
		return 0, fmt.Errorf("count promotions: %w", err)
	}
	return total, nil
}

func (s *Storage) UpdatePromotion(ctx context.Context, params internalModels.UpdatePromotionParams) (internalModels.Promotion, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return internalModels.Promotion{}, fmt.Errorf("begin update promotion: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	promotion := internalModels.Promotion{
		ID:              params.PromotionID,
		Name:            params.Name,
		DiscountPercent: params.DiscountPercent,
		StartsAt:        params.StartsAt,
		EndsAt:          params.EndsAt,
		Products:        params.Products,
	}
	err = tx.QueryRow(ctx,
		`UPDATE promotions SET name = $2, discount_percent = $3, starts_at = $4, ends_at = $5, updated_at = now()
		 WHERE id = $1 RETURNING created_at, updated_at`,
		params.PromotionID, params.Name, params.DiscountPercent, params.StartsAt, params.EndsAt,
	).Scan(&promotion.CreatedAt, &promotion.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return internalModels.Promotion{}, ErrPromotionNotFound
	}
	if err != nil {
		return internalModels.Promotion{}, fmt.Errorf("update promotion: %w", err)
	}
	if _, err = tx.Exec(ctx, `DELETE FROM promotion_products WHERE promotion_id = $1`, params.PromotionID); err != nil {
		return internalModels.Promotion{}, fmt.Errorf("clear promotion products: %w", err)
	}
	if err = insertPromotionProducts(ctx, tx, params.PromotionID, params.Products); err != nil {
		return internalModels.Promotion{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return internalModels.Promotion{}, fmt.Errorf("commit update promotion: %w", err)
	}
	return promotion, nil
}

func (s *Storage) DeletePromotion(ctx context.Context, promotionID uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM promotions WHERE id = $1`, promotionID)
	if err != nil {
		return fmt.Errorf("delete promotion: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrPromotionNotFound
	}
	return nil
}
