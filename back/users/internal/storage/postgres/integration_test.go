package postgres

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"

	internalModels "github.com/mbatimel/AMC/users/internal/models"
)

func TestStorageUserClientsAndFavoritesIntegration(t *testing.T) {
	dsn := os.Getenv("USERS_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("USERS_TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	t.Cleanup(pool.Close)

	var firstClientID, secondClientID uuid.UUID
	if err = pool.QueryRow(ctx, `
		INSERT INTO counterparties (type, name, inn, status)
		VALUES ('organization', 'Integration client 1', $1, 'active')
		RETURNING id
	`, uuid.NewString()).Scan(&firstClientID); err != nil {
		t.Fatalf("insert first client: %v", err)
	}
	if err = pool.QueryRow(ctx, `
		INSERT INTO counterparties (type, name, inn, status)
		VALUES ('ip', 'Integration client 2', $1, 'active')
		RETURNING id
	`, uuid.NewString()).Scan(&secondClientID); err != nil {
		t.Fatalf("insert second client: %v", err)
	}

	productIDs := []uuid.UUID{uuid.New(), uuid.New()}
	for index, productID := range productIDs {
		if _, err = pool.Exec(ctx, `
			INSERT INTO products (id, sku, name, slug)
			VALUES ($1, $2, $3, $4)
		`, productID, uuid.NewString(), "Integration product", uuid.NewString()); err != nil {
			t.Fatalf("insert product %d: %v", index, err)
		}
	}

	var userID uuid.UUID
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM favorites WHERE user_id = $1`, userID)
		_, _ = pool.Exec(ctx, `DELETE FROM user_clients WHERE user_id = $1`, userID)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
		_, _ = pool.Exec(ctx, `DELETE FROM products WHERE id = ANY($1::uuid[])`, productIDs)
		_, _ = pool.Exec(ctx, `DELETE FROM counterparties WHERE id = ANY($1::uuid[])`, []uuid.UUID{firstClientID, secondClientID})
	})

	storage := New(pool)
	email := "integration-" + uuid.NewString() + "@example.com"
	created, err := storage.CreateUser(ctx, internalModels.CreateUserParams{
		Email: email, FirstName: "Integration", LastName: "User",
		Status: "active", IsActive: true, ClientID: &firstClientID,
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	userID = created.ID
	if !created.ActiveClientID.Valid || created.ActiveClientID.UUID != firstClientID {
		t.Fatalf("unexpected initial active client: %+v", created.ActiveClientID)
	}

	updated, err := storage.UpdateUser(ctx, internalModels.UpdateUserParams{
		UserID: userID, ClientID: &secondClientID,
	})
	if err != nil {
		t.Fatalf("UpdateUser() error = %v", err)
	}
	if updated.ID != userID {
		t.Fatalf("unexpected updated user: %+v", updated)
	}

	clients, err := storage.ListUserClients(ctx, userID)
	if err != nil {
		t.Fatalf("ListUserClients() error = %v", err)
	}
	if len(clients) != 2 {
		t.Fatalf("client count = %d, want 2", len(clients))
	}

	activeClient, err := storage.SetActiveClient(ctx, userID, secondClientID)
	if err != nil {
		t.Fatalf("SetActiveClient() error = %v", err)
	}
	if activeClient.ID != secondClientID || !activeClient.IsActive {
		t.Fatalf("unexpected active client: %+v", activeClient)
	}
	storedActiveClientID, err := storage.GetActiveClient(ctx, userID)
	if err != nil || storedActiveClientID != secondClientID {
		t.Fatalf("GetActiveClient() id=%s error=%v", storedActiveClientID, err)
	}

	for _, productID := range productIDs {
		if _, _, err = storage.AddFavorite(ctx, userID, secondClientID, productID); err != nil {
			t.Fatalf("AddFavorite(%s) error = %v", productID, err)
		}
	}
	if _, createdAgain, err := storage.AddFavorite(ctx, userID, secondClientID, productIDs[0]); err != nil {
		t.Fatalf("idempotent AddFavorite() error = %v", err)
	} else if createdAgain {
		t.Fatal("duplicate favorite was inserted")
	}

	favorites, err := storage.ListFavorites(ctx, userID, secondClientID)
	if err != nil {
		t.Fatalf("ListFavorites() error = %v", err)
	}
	if len(favorites) != 2 {
		t.Fatalf("favorite count = %d, want 2", len(favorites))
	}
	deleted, err := storage.DeleteFavorites(ctx, userID, secondClientID, productIDs)
	if err != nil {
		t.Fatalf("DeleteFavorites() error = %v", err)
	}
	if deleted != 2 {
		t.Fatalf("deleted = %d, want 2", deleted)
	}

	active := true
	listParams := internalModels.ListUsersParams{
		Q: email, ClientID: &secondClientID, IsActive: &active,
		Limit: 10, Sort: "-email",
	}
	users, err := storage.ListUsers(ctx, listParams)
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}
	total, err := storage.CountUsers(ctx, listParams)
	if err != nil {
		t.Fatalf("CountUsers() error = %v", err)
	}
	if len(users) != 1 || total != 1 || users[0].ID != userID {
		t.Fatalf("unexpected filtered users=%+v total=%d", users, total)
	}
}
