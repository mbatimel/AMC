package postgres

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type fakeRow struct {
	value uuid.UUID
	err   error
}

func (r fakeRow) Scan(dest ...interface{}) error {
	if r.err != nil {
		return r.err
	}
	*(dest[0].(*uuid.UUID)) = r.value
	return nil
}

type fakeTransaction struct {
	rows       []fakeRow
	execCalls  []string
	execErrAt  int
	committed  bool
	rolledBack bool
}

func (tx *fakeTransaction) QueryRow(context.Context, string, ...interface{}) pgx.Row {
	row := tx.rows[0]
	tx.rows = tx.rows[1:]
	return row
}

func (tx *fakeTransaction) Exec(_ context.Context, sql string, _ ...interface{}) (pgconn.CommandTag, error) {
	tx.execCalls = append(tx.execCalls, sql)
	if tx.execErrAt > 0 && len(tx.execCalls) == tx.execErrAt {
		return nil, errors.New("write failed")
	}
	return pgconn.CommandTag("INSERT 0 1"), nil
}

func (tx *fakeTransaction) Commit(context.Context) error {
	tx.committed = true
	return nil
}

func (tx *fakeTransaction) Rollback(context.Context) error {
	tx.rolledBack = true
	return nil
}

func testIPArgs() []*string {
	values := make([]string, 18)
	result := make([]*string, len(values))
	for i := range values {
		result[i] = &values[i]
	}
	return result
}

func TestCreateIPUserLinksAndActivatesClientInTransaction(t *testing.T) {
	counterpartyID, userID, roleID := uuid.New(), uuid.New(), uuid.New()
	tx := &fakeTransaction{rows: []fakeRow{{value: counterpartyID}, {value: userID}, {value: roleID}}}
	storage := &Storage{beginTx: func(context.Context) (transaction, error) { return tx, nil }}
	args := testIPArgs()

	createdID, err := storage.CreateIPUser(
		context.Background(), "user@example.com", "hash", "Иванов", "Иван", "Иванович",
		args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8],
		args[9], args[10], args[11], args[12], args[13], args[14], args[15], args[16], args[17], 3,
	)
	if err != nil {
		t.Fatalf("CreateIPUser() error = %v", err)
	}
	if createdID != userID || !tx.committed {
		t.Fatalf("createdID=%s committed=%v", createdID, tx.committed)
	}
	if len(tx.execCalls) != 2 || tx.execCalls[0] != sqlInsertUserClient || tx.execCalls[1] != sqlInsertUserRole {
		t.Fatalf("transaction writes = %v", tx.execCalls)
	}
}

func TestCreateIPUserRollsBackWhenClientLinkFails(t *testing.T) {
	tx := &fakeTransaction{
		rows:      []fakeRow{{value: uuid.New()}, {value: uuid.New()}},
		execErrAt: 1,
	}
	storage := &Storage{beginTx: func(context.Context) (transaction, error) { return tx, nil }}
	args := testIPArgs()

	_, err := storage.CreateIPUser(
		context.Background(), "user@example.com", "hash", "Иванов", "Иван", "",
		args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8],
		args[9], args[10], args[11], args[12], args[13], args[14], args[15], args[16], args[17], 3,
	)
	if err == nil {
		t.Fatal("CreateIPUser() error = nil")
	}
	if tx.committed || !tx.rolledBack {
		t.Fatalf("committed=%v rolledBack=%v", tx.committed, tx.rolledBack)
	}
}

func TestCreateIPUserIntegration(t *testing.T) {
	dsn := os.Getenv("AUTH_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("AUTH_TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	t.Cleanup(pool.Close)

	storage := New(pool)
	email := "auth-registration-" + uuid.NewString() + "@example.com"
	companyName := "ООО Интеграция"
	inn := uuid.NewString()
	director := "Иванов Иван Иванович"
	phone := "+79990000000"
	args := testIPArgs()
	args[0], args[2], args[9], args[11] = &companyName, &inn, &director, &phone

	userID, err := storage.CreateIPUser(
		ctx, email, "hash", "Иванов", "Иван", "Иванович",
		args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8],
		args[9], args[10], args[11], args[12], args[13], args[14], args[15], args[16], args[17], 3,
	)
	if err != nil {
		t.Fatalf("CreateIPUser() error = %v", err)
	}
	var counterpartyID uuid.UUID
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM user_roles WHERE user_id = $1`, userID)
		_, _ = pool.Exec(ctx, `DELETE FROM user_clients WHERE user_id = $1`, userID)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
		_, _ = pool.Exec(ctx, `DELETE FROM counterparties WHERE id = $1`, counterpartyID)
	})

	var activeClientID uuid.UUID
	var surename, name, middleName, storedPhone string
	if err = pool.QueryRow(ctx, `
		SELECT counterparty_id, active_client_id, surename, name, middle_name, phone
		FROM users WHERE id = $1
	`, userID).Scan(&counterpartyID, &activeClientID, &surename, &name, &middleName, &storedPhone); err != nil {
		t.Fatalf("load registered user: %v", err)
	}
	if activeClientID != counterpartyID || surename != "Иванов" || name != "Иван" || middleName != "Иванович" || storedPhone != phone {
		t.Fatalf("active=%s counterparty=%s fio=%q %q %q phone=%q", activeClientID, counterpartyID, surename, name, middleName, storedPhone)
	}
	var linked, isDefault bool
	if err = pool.QueryRow(ctx, `
		SELECT TRUE, is_default FROM user_clients WHERE user_id = $1 AND client_id = $2
	`, userID, counterpartyID).Scan(&linked, &isDefault); err != nil {
		t.Fatalf("load user_clients link: %v", err)
	}
	if !linked || !isDefault {
		t.Fatalf("linked=%v default=%v", linked, isDefault)
	}

	duplicateCompany := "ООО Не должно сохраниться"
	duplicateINN := uuid.NewString()
	duplicateArgs := testIPArgs()
	duplicateArgs[0], duplicateArgs[2] = &duplicateCompany, &duplicateINN
	_, err = storage.CreateIPUser(
		ctx, email, "other-hash", "Петров", "Пётр", "",
		duplicateArgs[0], duplicateArgs[1], duplicateArgs[2], duplicateArgs[3], duplicateArgs[4], duplicateArgs[5], duplicateArgs[6], duplicateArgs[7], duplicateArgs[8],
		duplicateArgs[9], duplicateArgs[10], duplicateArgs[11], duplicateArgs[12], duplicateArgs[13], duplicateArgs[14], duplicateArgs[15], duplicateArgs[16], duplicateArgs[17], 3,
	)
	if !errors.Is(err, ErrEmailTaken) {
		t.Fatalf("duplicate CreateIPUser() error = %v, want ErrEmailTaken", err)
	}
	var partialCounterparties int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM counterparties WHERE inn = $1`, duplicateINN).Scan(&partialCounterparties); err != nil {
		t.Fatalf("count partial counterparties: %v", err)
	}
	if partialCounterparties != 0 {
		t.Fatalf("partial counterparties = %d, want 0", partialCounterparties)
	}

	if _, err = pool.Exec(ctx, `UPDATE users SET is_active = FALSE, deleted_at = now() WHERE id = $1`, userID); err != nil {
		t.Fatalf("soft delete user: %v", err)
	}
	if _, err = storage.GetUserByEmail(ctx, email); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("GetUserByEmail() after soft delete error = %v, want ErrUserNotFound", err)
	}
}
