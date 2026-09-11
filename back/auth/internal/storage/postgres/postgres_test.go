package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type fakeRow struct {
	value uuid.UUID
	err   error
	scan  func(dest ...interface{}) error
}

func (r fakeRow) Scan(dest ...interface{}) error {
	if r.err != nil {
		return r.err
	}
	if r.scan != nil {
		return r.scan(dest...)
	}
	*(dest[0].(*uuid.UUID)) = r.value
	return nil
}

func counterpartyRow(id uuid.UUID, matchCount int) fakeRow {
	return fakeRow{scan: func(dest ...interface{}) error {
		*(dest[0].(*uuid.UUID)) = id
		*(dest[1].(*int)) = matchCount
		return nil
	}}
}

func boolRow(value bool) fakeRow {
	return fakeRow{scan: func(dest ...interface{}) error {
		*(dest[0].(*bool)) = value
		return nil
	}}
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

func testCanonicalINN() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "")
}

func TestCreateIPUserLinksAndActivatesClientInTransaction(t *testing.T) {
	counterpartyID, userID, roleID := uuid.New(), uuid.New(), uuid.New()
	tx := &fakeTransaction{rows: []fakeRow{{err: pgx.ErrNoRows}, {value: counterpartyID}, {value: userID}, {value: roleID}}}
	storage := &Storage{beginTx: func(context.Context) (transaction, error) { return tx, nil }}
	args := testIPArgs()

	createdID, err := storage.CreateIPUser(
		context.Background(), "user@example.com", "hash", "Иванов", "Иван", "Иванович",
		args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8],
		args[9], args[10], args[11], args[12], args[13], args[14], args[15], args[16], args[17],
		"", "", 3,
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
		rows:      []fakeRow{{err: pgx.ErrNoRows}, {value: uuid.New()}, {value: uuid.New()}},
		execErrAt: 1,
	}
	storage := &Storage{beginTx: func(context.Context) (transaction, error) { return tx, nil }}
	args := testIPArgs()

	_, err := storage.CreateIPUser(
		context.Background(), "user@example.com", "hash", "Иванов", "Иван", "",
		args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8],
		args[9], args[10], args[11], args[12], args[13], args[14], args[15], args[16], args[17],
		"", "", 3,
	)
	if err == nil {
		t.Fatal("CreateIPUser() error = nil")
	}
	if tx.committed || !tx.rolledBack {
		t.Fatalf("committed=%v rolledBack=%v", tx.committed, tx.rolledBack)
	}
}

func TestCreateIPUserReusesOrphanCounterparty(t *testing.T) {
	counterpartyID, userID, roleID := uuid.New(), uuid.New(), uuid.New()
	tx := &fakeTransaction{rows: []fakeRow{
		counterpartyRow(counterpartyID, 1),
		boolRow(false),
		{value: userID},
		{value: roleID},
	}}
	storage := &Storage{beginTx: func(context.Context) (transaction, error) { return tx, nil }}
	args := testIPArgs()

	createdID, err := storage.CreateIPUser(
		context.Background(), "user@example.com", "hash", "Иванов", "Иван", "",
		args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8],
		args[9], args[10], args[11], args[12], args[13], args[14], args[15], args[16], args[17],
		"https://s3.example.com/requisites.pdf", "requisites.pdf", 3,
	)
	if err != nil {
		t.Fatalf("CreateIPUser() error = %v", err)
	}
	if createdID != userID || !tx.committed {
		t.Fatalf("createdID=%s committed=%v", createdID, tx.committed)
	}
	if len(tx.execCalls) != 3 || tx.execCalls[0] != sqlUpdateCounterpartyForRegistration || tx.execCalls[1] != sqlInsertUserClient || tx.execCalls[2] != sqlInsertUserRole {
		t.Fatalf("transaction writes = %v", tx.execCalls)
	}
}

func TestCreateIPUserRejectsCounterpartyWithActiveUser(t *testing.T) {
	tx := &fakeTransaction{rows: []fakeRow{counterpartyRow(uuid.New(), 1), boolRow(true)}}
	storage := &Storage{beginTx: func(context.Context) (transaction, error) { return tx, nil }}
	args := testIPArgs()

	_, err := storage.CreateIPUser(
		context.Background(), "user@example.com", "hash", "Иванов", "Иван", "",
		args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8],
		args[9], args[10], args[11], args[12], args[13], args[14], args[15], args[16], args[17],
		"", "", 3,
	)
	if !errors.Is(err, ErrInnTaken) {
		t.Fatalf("CreateIPUser() error = %v, want ErrInnTaken", err)
	}
	if tx.committed || !tx.rolledBack || len(tx.execCalls) != 0 {
		t.Fatalf("committed=%v rolledBack=%v writes=%d", tx.committed, tx.rolledBack, len(tx.execCalls))
	}
}

func TestCreateIPUserMapsKnownUniqueConstraints(t *testing.T) {
	tests := []struct {
		name       string
		constraint string
		want       error
	}{
		{name: "email expression index", constraint: "idx_users_email_lower_unique", want: ErrEmailTaken},
		{name: "legacy email constraint", constraint: "users_email_key", want: ErrEmailTaken},
		{name: "phone", constraint: "idx_users_phone_unique", want: ErrPhoneTaken},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &fakeTransaction{rows: []fakeRow{
				{err: pgx.ErrNoRows},
				{value: uuid.New()},
				{err: &pgconn.PgError{Code: uniqueViolationCode, ConstraintName: tt.constraint}},
			}}
			storage := &Storage{beginTx: func(context.Context) (transaction, error) { return tx, nil }}
			args := testIPArgs()

			_, err := storage.CreateIPUser(
				context.Background(), "user@example.com", "hash", "Иванов", "Иван", "",
				args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8],
				args[9], args[10], args[11], args[12], args[13], args[14], args[15], args[16], args[17],
				"", "", 3,
			)
			if !errors.Is(err, tt.want) {
				t.Fatalf("CreateIPUser() error = %v, want %v", err, tt.want)
			}
			if tx.committed || !tx.rolledBack {
				t.Fatalf("committed=%v rolledBack=%v", tx.committed, tx.rolledBack)
			}
		})
	}
}

func TestCreateIPUserUnknownUniqueConstraintIsNotEmail(t *testing.T) {
	tx := &fakeTransaction{rows: []fakeRow{
		{err: pgx.ErrNoRows},
		{value: uuid.New()},
		{err: &pgconn.PgError{Code: uniqueViolationCode, ConstraintName: "users_unknown_key"}},
	}}
	storage := &Storage{beginTx: func(context.Context) (transaction, error) { return tx, nil }}
	args := testIPArgs()

	_, err := storage.CreateIPUser(
		context.Background(), "user@example.com", "hash", "Иванов", "Иван", "",
		args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8],
		args[9], args[10], args[11], args[12], args[13], args[14], args[15], args[16], args[17],
		"", "", 3,
	)
	if err == nil || errors.Is(err, ErrEmailTaken) || errors.Is(err, ErrPhoneTaken) {
		t.Fatalf("CreateIPUser() error = %v, want unclassified database error", err)
	}
}

func TestCreateIPUserMapsCounterpartyINNConstraint(t *testing.T) {
	tx := &fakeTransaction{rows: []fakeRow{
		{err: pgx.ErrNoRows},
		{err: &pgconn.PgError{Code: uniqueViolationCode, ConstraintName: "uq_counterparties_inn"}},
	}}
	storage := &Storage{beginTx: func(context.Context) (transaction, error) { return tx, nil }}
	args := testIPArgs()

	_, err := storage.CreateIPUser(
		context.Background(), "user@example.com", "hash", "Иванов", "Иван", "",
		args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8],
		args[9], args[10], args[11], args[12], args[13], args[14], args[15], args[16], args[17],
		"", "", 3,
	)
	if !errors.Is(err, ErrInnTaken) {
		t.Fatalf("CreateIPUser() error = %v, want ErrInnTaken", err)
	}
}

func TestCreateIPUserRollsBackWhenRoleInsertFails(t *testing.T) {
	tx := &fakeTransaction{
		rows:      []fakeRow{{err: pgx.ErrNoRows}, {value: uuid.New()}, {value: uuid.New()}, {value: uuid.New()}},
		execErrAt: 2,
	}
	storage := &Storage{beginTx: func(context.Context) (transaction, error) { return tx, nil }}
	args := testIPArgs()

	_, err := storage.CreateIPUser(
		context.Background(), "user@example.com", "hash", "Иванов", "Иван", "",
		args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8],
		args[9], args[10], args[11], args[12], args[13], args[14], args[15], args[16], args[17],
		"", "", 3,
	)
	if err == nil || tx.committed || !tx.rolledBack {
		t.Fatalf("error=%v committed=%v rolledBack=%v", err, tx.committed, tx.rolledBack)
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
	inn := testCanonicalINN()
	director := "Иванов Иван Иванович"
	phone := "+79990000000"
	args := testIPArgs()
	args[0], args[2], args[9], args[11] = &companyName, &inn, &director, &phone

	userID, err := storage.CreateIPUser(
		ctx, email, "hash", "Иванов", "Иван", "Иванович",
		args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8],
		args[9], args[10], args[11], args[12], args[13], args[14], args[15], args[16], args[17],
		"https://s3.example.com/requisites.pdf", "requisites.pdf", 3,
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
	if inUse, checkErr := storage.CounterpartyINNInUse(ctx, inn); checkErr != nil || !inUse {
		t.Fatalf("CounterpartyINNInUse() inUse=%v error=%v, want true", inUse, checkErr)
	}
	if loaded, loadErr := storage.GetUserByEmail(ctx, strings.ToUpper(email)); loadErr != nil || loaded.ID != userID {
		t.Fatalf("case-insensitive GetUserByEmail() user=%s error=%v", loaded.ID, loadErr)
	}

	activeDuplicateArgs := testIPArgs()
	activeDuplicateArgs[2] = &inn
	_, err = storage.CreateIPUser(
		ctx, "other-"+email, "other-hash", "Петров", "Пётр", "",
		activeDuplicateArgs[0], activeDuplicateArgs[1], activeDuplicateArgs[2], activeDuplicateArgs[3], activeDuplicateArgs[4], activeDuplicateArgs[5], activeDuplicateArgs[6], activeDuplicateArgs[7], activeDuplicateArgs[8],
		activeDuplicateArgs[9], activeDuplicateArgs[10], activeDuplicateArgs[11], activeDuplicateArgs[12], activeDuplicateArgs[13], activeDuplicateArgs[14], activeDuplicateArgs[15], activeDuplicateArgs[16], activeDuplicateArgs[17],
		"", "", 3,
	)
	if !errors.Is(err, ErrInnTaken) {
		t.Fatalf("active counterparty duplicate error = %v, want ErrInnTaken", err)
	}

	duplicateCompany := "ООО Не должно сохраниться"
	duplicateINN := testCanonicalINN()
	duplicateArgs := testIPArgs()
	duplicateArgs[0], duplicateArgs[2] = &duplicateCompany, &duplicateINN
	_, err = storage.CreateIPUser(
		ctx, email, "other-hash", "Петров", "Пётр", "",
		duplicateArgs[0], duplicateArgs[1], duplicateArgs[2], duplicateArgs[3], duplicateArgs[4], duplicateArgs[5], duplicateArgs[6], duplicateArgs[7], duplicateArgs[8],
		duplicateArgs[9], duplicateArgs[10], duplicateArgs[11], duplicateArgs[12], duplicateArgs[13], duplicateArgs[14], duplicateArgs[15], duplicateArgs[16], duplicateArgs[17],
		"", "", 3,
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

	caseDuplicateINN := testCanonicalINN()
	caseDuplicatePhone := "+79990000001"
	caseDuplicateArgs := testIPArgs()
	caseDuplicateArgs[2], caseDuplicateArgs[11] = &caseDuplicateINN, &caseDuplicatePhone
	_, err = storage.CreateIPUser(
		ctx, strings.ToUpper(email), "other-hash", "Петров", "Пётр", "",
		caseDuplicateArgs[0], caseDuplicateArgs[1], caseDuplicateArgs[2], caseDuplicateArgs[3], caseDuplicateArgs[4], caseDuplicateArgs[5], caseDuplicateArgs[6], caseDuplicateArgs[7], caseDuplicateArgs[8],
		caseDuplicateArgs[9], caseDuplicateArgs[10], caseDuplicateArgs[11], caseDuplicateArgs[12], caseDuplicateArgs[13], caseDuplicateArgs[14], caseDuplicateArgs[15], caseDuplicateArgs[16], caseDuplicateArgs[17],
		"", "", 3,
	)
	if !errors.Is(err, ErrEmailTaken) {
		t.Fatalf("case-insensitive email duplicate error = %v, want ErrEmailTaken", err)
	}

	phoneDuplicateINN := testCanonicalINN()
	phoneDuplicateArgs := testIPArgs()
	phoneDuplicateArgs[2], phoneDuplicateArgs[11] = &phoneDuplicateINN, &phone
	_, err = storage.CreateIPUser(
		ctx, "phone-"+email, "other-hash", "Петров", "Пётр", "",
		phoneDuplicateArgs[0], phoneDuplicateArgs[1], phoneDuplicateArgs[2], phoneDuplicateArgs[3], phoneDuplicateArgs[4], phoneDuplicateArgs[5], phoneDuplicateArgs[6], phoneDuplicateArgs[7], phoneDuplicateArgs[8],
		phoneDuplicateArgs[9], phoneDuplicateArgs[10], phoneDuplicateArgs[11], phoneDuplicateArgs[12], phoneDuplicateArgs[13], phoneDuplicateArgs[14], phoneDuplicateArgs[15], phoneDuplicateArgs[16], phoneDuplicateArgs[17],
		"", "", 3,
	)
	if !errors.Is(err, ErrPhoneTaken) {
		t.Fatalf("phone duplicate error = %v, want ErrPhoneTaken", err)
	}
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM counterparties WHERE inn = $1`, phoneDuplicateINN).Scan(&partialCounterparties); err != nil {
		t.Fatalf("count phone-conflict counterparty: %v", err)
	}
	if partialCounterparties != 0 {
		t.Fatalf("phone-conflict partial counterparties = %d, want 0", partialCounterparties)
	}

	if _, err = pool.Exec(ctx, `UPDATE users SET is_active = FALSE, deleted_at = now() WHERE id = $1`, userID); err != nil {
		t.Fatalf("soft delete user: %v", err)
	}
	if _, err = storage.GetUserByEmail(ctx, email); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("GetUserByEmail() after soft delete error = %v, want ErrUserNotFound", err)
	}
	if inUse, checkErr := storage.CounterpartyINNInUse(ctx, inn); checkErr != nil || inUse {
		t.Fatalf("CounterpartyINNInUse() after soft delete inUse=%v error=%v, want false", inUse, checkErr)
	}

	replacementUserID, err := storage.CreateIPUser(
		ctx, email, "replacement-hash", "Новый", "Пользователь", "",
		args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8],
		args[9], args[10], args[11], args[12], args[13], args[14], args[15], args[16], args[17],
		"https://s3.example.com/replacement.pdf", "replacement.pdf", 3,
	)
	if err != nil {
		t.Fatalf("reuse counterparty after soft delete: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM user_roles WHERE user_id = $1`, replacementUserID)
		_, _ = pool.Exec(ctx, `DELETE FROM user_clients WHERE user_id = $1`, replacementUserID)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, replacementUserID)
	})
	var reusedCounterpartyID uuid.UUID
	if err = pool.QueryRow(ctx, `SELECT counterparty_id FROM users WHERE id = $1`, replacementUserID).Scan(&reusedCounterpartyID); err != nil {
		t.Fatalf("load replacement user: %v", err)
	}
	if reusedCounterpartyID != counterpartyID {
		t.Fatalf("replacement counterparty = %s, want reused %s", reusedCounterpartyID, counterpartyID)
	}
}

func TestCreateIPUserConcurrentOrphanReuseIntegration(t *testing.T) {
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

	inn := testCanonicalINN()
	formattedINN := inn[:12] + " - " + inn[12:]
	var counterpartyID uuid.UUID
	if err = pool.QueryRow(ctx, `INSERT INTO counterparties (type, name, inn, status) VALUES ('ip', 'orphan race test', $1, 'new') RETURNING id`, formattedINN).Scan(&counterpartyID); err != nil {
		t.Fatalf("insert orphan counterparty: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM user_roles WHERE user_id IN (SELECT id FROM users WHERE counterparty_id = $1)`, counterpartyID)
		_, _ = pool.Exec(ctx, `DELETE FROM user_clients WHERE client_id = $1`, counterpartyID)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE counterparty_id = $1`, counterpartyID)
		_, _ = pool.Exec(ctx, `DELETE FROM counterparties WHERE id = $1`, counterpartyID)
	})

	storage := New(pool)
	start := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			<-start
			args := testIPArgs()
			phone := fmt.Sprintf("+7999000010%d", index)
			args[2], args[11] = &inn, &phone
			_, createErr := storage.CreateIPUser(
				ctx, fmt.Sprintf("orphan-race-%d-%s@example.com", index, uuid.NewString()), "hash", "Иванов", "Иван", "",
				args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8],
				args[9], args[10], args[11], args[12], args[13], args[14], args[15], args[16], args[17],
				"", "", 3,
			)
			results <- createErr
		}(i)
	}
	close(start)
	wg.Wait()
	close(results)

	var successes, innConflicts int
	for createErr := range results {
		switch {
		case createErr == nil:
			successes++
		case errors.Is(createErr, ErrInnTaken):
			innConflicts++
		default:
			t.Fatalf("unexpected concurrent CreateIPUser() error: %v", createErr)
		}
	}
	if successes != 1 || innConflicts != 1 {
		t.Fatalf("successes=%d INN conflicts=%d, want 1 and 1", successes, innConflicts)
	}

	var storedINN string
	if err = pool.QueryRow(ctx, `SELECT inn FROM counterparties WHERE id = $1`, counterpartyID).Scan(&storedINN); err != nil {
		t.Fatalf("load reused counterparty INN: %v", err)
	}
	if storedINN != inn {
		t.Fatalf("reused counterparty INN = %q, want canonical %q", storedINN, inn)
	}
}
