package trading

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/AlexGalhardo/vukem/database"
)

type (
	Store  struct{ DB *database.DB }
	Record struct {
		ID        string          `json:"id"`
		Payload   json.RawMessage `json:"payload"`
		Version   int64           `json:"version"`
		UpdatedAt string          `json:"updated_at"`
		Status    string          `json:"status"`
	}
)

func validTable(t string) bool {
	return t == "trading_accounts" || t == "datasets" || t == "strategies" || t == "backtests" || t == "artifacts"
}
func (s Store) q(query string) string { return database.Rebind(s.DB.Dialect(), query) }
func (s Store) Initialize(ctx context.Context) error {
	for _, a := range []State{
		NewState("b3", "BRL", "simulation"), NewState("us", "USD", "paper_broker"), NewState("research", "USD", "backtest"),
	} {
		if a.ID == "us" {
			a.Cash = Zero
			a.Initial = Zero
			a.DayStart = Zero
			a.Peak = Zero
		}
		b, _ := json.Marshal(a)
		_, err := s.DB.ExecContext(
			ctx,
			s.q(
				"INSERT INTO trading_accounts (id,payload,version,updated_at,status) VALUES (?,?,0,?,'halted') ON CONFLICT (id) DO NOTHING",
			),
			a.ID,
			string(b),
			time.Now().UTC().Format(time.RFC3339),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s Store) List(ctx context.Context, table string, limit int) ([]Record, error) {
	if !validTable(table) {
		return nil, errors.New("invalid table")
	}
	if limit < 1 || limit > 200 {
		limit = 100
	}
	order := "updated_at DESC"
	if table == "strategies" {
		order = "CASE WHEN status='enabled' THEN 0 ELSE 1 END, updated_at DESC"
	}
	rows, err := s.DB.QueryContext(
		ctx,
		fmt.Sprintf("SELECT id,payload,version,updated_at,status FROM %s ORDER BY %s LIMIT %d", table, order, limit),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Record{}
	for rows.Next() {
		var r Record
		var body string
		if err := rows.Scan(&r.ID, &body, &r.Version, &r.UpdatedAt, &r.Status); err != nil {
			return nil, err
		}
		r.Payload = json.RawMessage(body)
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s Store) Get(ctx context.Context, table, id string, out any) error {
	if !validTable(table) {
		return errors.New("invalid table")
	}
	var body string
	if err := s.DB.QueryRowContext(ctx, s.q("SELECT payload FROM "+table+" WHERE id=?"), id).Scan(&body); err != nil {
		return err
	}
	return json.Unmarshal([]byte(body), out)
}

func (s Store) Create(ctx context.Context, table, id, status, actor string, value any) error {
	if !validTable(table) {
		return errors.New("invalid table")
	}
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.DB.Tx(ctx, func(tx *database.Tx) error {
		query := "INSERT INTO " + table + " (id,payload,version,updated_at,status) VALUES (?,?,0,?,?)"
		if table == "artifacts" {
			query += " ON CONFLICT (id) DO NOTHING"
		}
		_, err := tx.ExecContext(ctx, s.q(query), id, string(b), time.Now().UTC().Format(time.RFC3339), status)
		if err != nil {
			return err
		}
		return audit(ctx, tx, actor, "create."+table, id, "Created immutable record")
	})
}

func audit(ctx context.Context, tx *database.Tx, actor, action, account, details string) error {
	_, err := tx.ExecContext(
		ctx,
		database.Rebind(
			tx.Dialect(),
			"INSERT INTO audit_events (actor,action,account_id,details,created_at) VALUES (?,?,?,?,?)",
		),
		actor,
		action,
		account,
		details,
		time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

// Mutate serializes account decisions and persists the reserve and intent together.
func (s Store) Mutate(ctx context.Context, id, actor, action string, fn func(*State) error) error {
	return s.DB.Tx(ctx, func(tx *database.Tx) error {
		var raw string
		var version int64
		q := s.q("SELECT payload,version FROM trading_accounts WHERE id=?")
		if tx.Dialect().Name() == "postgres" {
			q += " FOR UPDATE"
		}
		if err := tx.QueryRowContext(ctx, q, id).Scan(&raw, &version); err != nil {
			return err
		}
		var a State
		if err := json.Unmarshal([]byte(raw), &a); err != nil {
			return err
		}
		oldPolicy := a.Policy
		oldDecisions := len(a.RiskDecisions)
		wasHalted, oldReason := a.Halted, a.Reason
		if err := fn(&a); err != nil {
			return err
		}
		for _, decision := range a.RiskDecisions[oldDecisions:] {
			raw, _ := json.Marshal(decision)
			if err := audit(ctx, tx, "engine", "risk.rejected", id, string(raw)); err != nil {
				return err
			}
		}
		if a.Halted && (!wasHalted || a.Reason != oldReason) {
			if err := audit(ctx, tx, "engine", "risk.halt", id, a.Reason); err != nil {
				return err
			}
		}
		if wasHalted && !a.Halted {
			if err := audit(ctx, tx, "engine", "risk.resume", id, "Halt lifted: "+oldReason); err != nil {
				return err
			}
		}
		b, err := json.Marshal(a)
		if err != nil {
			return err
		}
		status := "ready"
		if a.Halted {
			status = "halted"
		}
		res, err := tx.ExecContext(
			ctx,
			s.q(
				"UPDATE trading_accounts SET payload=?,version=version+1,updated_at=?,status=? WHERE id=? AND version=?",
			),
			string(b),
			time.Now().UTC().Format(time.RFC3339),
			status,
			id,
			version,
		)
		if err != nil {
			return err
		}
		n, _ := res.RowsAffected()
		if n != 1 {
			return errors.New("concurrent update; retry")
		}
		if actor != "engine" || action != "tick" {
			details := "Generation " + strconv.FormatInt(a.Generation, 10)
			if action == "risk.policy" {
				b, _ := json.Marshal(map[string]any{"before": oldPolicy, "after": a.Policy, "generation": a.Generation})
				details = string(b)
			}
			return audit(ctx, tx, actor, action, id, details)
		}
		return nil
	})
}

func (s Store) Status(ctx context.Context, table, id, status, actor string) error {
	if !validTable(table) {
		return errors.New("invalid table")
	}
	if table != "backtests" && table != "strategies" {
		return errors.New("status not mutable")
	}
	return s.DB.Tx(ctx, func(tx *database.Tx) error {
		query := "UPDATE " + table + " SET status=?,updated_at=? WHERE id=?"
		if table == "backtests" {
			query += " AND status IN ('queued','running')"
		}
		res, err := tx.ExecContext(ctx, s.q(query), status, time.Now().UTC().Format(time.RFC3339), id)
		if err != nil {
			return err
		}
		n, _ := res.RowsAffected()
		if n != 1 {
			return sql.ErrNoRows
		}
		return audit(ctx, tx, actor, table+"."+status, id, "Status changed")
	})
}

type AuditEvent struct {
	ID      int64  `json:"id"`
	Actor   string `json:"actor"`
	Action  string `json:"action"`
	Account string `json:"account"`
	Details string `json:"details"`
	Time    string `json:"time"`
}

func (s Store) Audit(ctx context.Context) ([]AuditEvent, error) {
	rows, err := s.DB.QueryContext(
		ctx,
		"SELECT id,actor,action,account_id,details,created_at FROM audit_events ORDER BY id DESC LIMIT 100",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []AuditEvent{}
	for rows.Next() {
		var a AuditEvent
		if err := rows.Scan(&a.ID, &a.Actor, &a.Action, &a.Account, &a.Details, &a.Time); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s Store) ActiveStrategies(ctx context.Context) ([]Strategy, error) {
	rows, err := s.DB.QueryContext(ctx, "SELECT payload FROM strategies WHERE status='enabled' ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Strategy{}
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		var st Strategy
		if err := json.Unmarshal([]byte(raw), &st); err != nil {
			return nil, err
		}
		st.Enabled = true
		out = append(out, st)
	}
	return out, rows.Err()
}

type Job struct {
	ID          string    `json:"id"`
	DatasetID   string    `json:"dataset_id"`
	Strategy    Strategy  `json:"strategy"`
	Policy      Policy    `json:"policy"`
	Partition   string    `json:"partition"`
	Result      *Result   `json:"result,omitempty"`
	Error       string    `json:"error,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	FinishedAt  time.Time `json:"finished_at"`
	CodeVersion string    `json:"code_version"`
}

func (s Store) ProcessJob(ctx context.Context) error {
	var id, raw string
	err := s.DB.QueryRowContext(ctx, "SELECT id,payload FROM backtests WHERE status='queued' ORDER BY updated_at LIMIT 1").
		Scan(&id, &raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	res, err := s.DB.ExecContext(
		ctx,
		s.q("UPDATE backtests SET status='running',updated_at=? WHERE id=? AND status='queued'"),
		time.Now().UTC().Format(time.RFC3339),
		id,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil
	}
	var job Job
	if err := json.Unmarshal([]byte(raw), &job); err != nil {
		return err
	}
	var ds Dataset
	err = s.Get(ctx, "datasets", job.DatasetID, &ds)
	runCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	done := make(chan struct{})
	defer close(done)
	go func() {
		tick := time.NewTicker(200 * time.Millisecond)
		defer tick.Stop()
		for {
			select {
			case <-done:
				return
			case <-runCtx.Done():
				return
			case <-tick.C:
				var status string
				e := s.DB.QueryRowContext(runCtx, s.q("SELECT status FROM backtests WHERE id=?"), id).Scan(&status)
				if e != nil || status != "running" {
					cancel()
					return
				}
			}
		}
	}()
	var result Result
	if err == nil {
		trials, _ := s.Trials(runCtx, job.Strategy)
		result, err = BacktestTrials(runCtx, ds, job.Strategy, job.Policy, job.Partition, trials)
	}
	status := "completed"
	if err != nil {
		status = "failed"
		job.Error = err.Error()
	} else {
		if err := s.Create(ctx, "artifacts", job.ID, "backtest", "engine", result); err != nil {
			return err
		}
		result.Curve = nil
		result.Fills = nil
		result.Final = State{}
		job.Result = &result
	}
	job.FinishedAt = time.Now().UTC()
	b, e := json.Marshal(job)
	if e != nil {
		return e
	}
	_, e = s.DB.ExecContext(
		ctx,
		s.q("UPDATE backtests SET payload=?,status=?,updated_at=? WHERE id=? AND status='running'"),
		string(b),
		status,
		time.Now().UTC().Format(time.RFC3339),
		id,
	)
	return e
}

// ToggleStrategy serializes activation with the account lock. B3 uses one replay
// clock per account; US permits one active configuration per symbol.
func (s Store) ToggleStrategy(ctx context.Context, id string, enabled bool) error {
	return s.DB.Tx(ctx, func(tx *database.Tx) error {
		var raw string
		if err := tx.QueryRowContext(ctx, s.q("SELECT payload FROM strategies WHERE id=?"), id).Scan(&raw); err != nil {
			return err
		}
		var strategy Strategy
		if err := json.Unmarshal([]byte(raw), &strategy); err != nil {
			return err
		}
		q := s.q("SELECT payload FROM trading_accounts WHERE id=?")
		if tx.Dialect().Name() == "postgres" {
			q += " FOR UPDATE"
		}
		if err := tx.QueryRowContext(ctx, q, strategy.AccountID).Scan(&raw); err != nil {
			return err
		}
		var a State
		if err := json.Unmarshal([]byte(raw), &a); err != nil {
			return err
		}
		if !enabled {
			if p := a.Positions[strategy.Symbol]; p.Qty.IsPositive() {
				return errors.New("open position: keep the protection strategy enabled")
			}
			for _, o := range a.Orders {
				if o.Open() && o.Symbol == strategy.Symbol {
					return errors.New("pending order on symbol")
				}
			}
		}
		if enabled {
			conflict := false
			err := func() error {
				rows, err := tx.QueryContext(ctx, "SELECT id,payload FROM strategies WHERE status='enabled'")
				if err != nil {
					return err
				}
				defer rows.Close()
				for rows.Next() {
					var otherID, body string
					if err := rows.Scan(&otherID, &body); err != nil {
						return err
					}
					var other Strategy
					if err := json.Unmarshal([]byte(body), &other); err != nil {
						return err
					}
					if otherID != id && other.AccountID == strategy.AccountID &&
						(a.ID == "b3" || other.Symbol == strategy.Symbol) {
						conflict = true
					}
				}
				return rows.Err()
			}()
			if err != nil {
				return err
			}
			if conflict {
				return errors.New(
					"an active strategy already exists for this account/symbol; B3 allows one replay clock at a time",
				)
			}
		}
		status := "disabled"
		if enabled {
			status = "enabled"
		}
		_, err := tx.ExecContext(
			ctx,
			s.q("UPDATE strategies SET status=?,version=version+1,updated_at=? WHERE id=?"),
			status,
			time.Now().UTC().Format(time.RFC3339),
			id,
		)
		if err != nil {
			return err
		}
		return audit(ctx, tx, "owner", "strategy."+status, strategy.AccountID, id)
	})
}

func (s Store) CreateStrategy(ctx context.Context, st *Strategy) error {
	return s.DB.Tx(ctx, func(tx *database.Tx) error {
		var account string
		q := s.q("SELECT payload FROM trading_accounts WHERE id=?")
		if tx.Dialect().Name() == "postgres" {
			q += " FOR UPDATE"
		}
		if err := tx.QueryRowContext(ctx, q, st.AccountID).Scan(&account); err != nil {
			return err
		}
		st.Version = 1
		err := func() error {
			rows, err := tx.QueryContext(ctx, "SELECT payload FROM strategies")
			if err != nil {
				return err
			}
			defer rows.Close()
			for rows.Next() {
				var body string
				if err := rows.Scan(&body); err != nil {
					return err
				}
				var other Strategy
				if err := json.Unmarshal([]byte(body), &other); err != nil {
					return err
				}
				if other.Name == st.Name && other.Symbol == st.Symbol && other.AccountID == st.AccountID &&
					other.Version >= st.Version {
					st.Version = other.Version + 1
				}
			}
			return rows.Err()
		}()
		if err != nil {
			return err
		}
		body, err := json.Marshal(st)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(
			ctx,
			s.q("INSERT INTO strategies(id,payload,version,updated_at,status) VALUES (?,?,0,?,'disabled')"),
			st.ID,
			string(body),
			time.Now().UTC().Format(time.RFC3339),
		)
		if err != nil {
			return err
		}
		return audit(ctx, tx, "owner", "strategy.create", st.AccountID, st.ID+" version "+strconv.Itoa(st.Version))
	})
}

func (s Store) ResetReplay(ctx context.Context) error {
	return s.DB.Tx(ctx, func(tx *database.Tx) error {
		var raw string
		var version int64
		q := "SELECT payload,version FROM trading_accounts WHERE id='b3'"
		if tx.Dialect().Name() == "postgres" {
			q += " FOR UPDATE"
		}
		if err := tx.QueryRowContext(ctx, q).Scan(&raw, &version); err != nil {
			return err
		}
		var a State
		if err := json.Unmarshal([]byte(raw), &a); err != nil {
			return err
		}
		if !a.Halted || len(a.Positions) > 0 {
			return errors.New("block and close positions before starting a new replay")
		}
		for _, o := range a.Orders {
			if o.Open() {
				return errors.New("open orders block a new replay")
			}
		}
		id := "replay-" + ID()
		now := time.Now().UTC().Format(time.RFC3339)
		if _, err := tx.ExecContext(
			ctx,
			s.q("INSERT INTO artifacts(id,payload,version,updated_at,status) VALUES (?,?,0,?,'replay')"),
			id,
			raw,
			now,
		); err != nil {
			return err
		}
		next := NewState("b3", "BRL", "simulation")
		next.LeaderEpoch = a.LeaderEpoch
		next.Generation = a.Generation + 1
		next.Heartbeat = a.Heartbeat
		next.Policy = a.Policy
		body, _ := json.Marshal(next)
		res, err := tx.ExecContext(
			ctx,
			s.q(
				"UPDATE trading_accounts SET payload=?,version=version+1,updated_at=?,status='halted' WHERE id='b3' AND version=?",
			),
			string(body),
			now,
			version,
		)
		if err != nil {
			return err
		}
		n, _ := res.RowsAffected()
		if n != 1 {
			return errors.New("concurrent reset")
		}
		return audit(ctx, tx, "owner", "simulation.reset", "b3", id)
	})
}

// ValidationUsed reports whether the held-out validation partition was already run for this exact
// strategy version. Validation is the final check on a candidate; running it again to tune the
// candidate turns it into a training set.
func (s Store) ValidationUsed(ctx context.Context, strategyID string) (bool, error) {
	rows, err := s.DB.QueryContext(
		ctx,
		"SELECT payload FROM backtests WHERE status IN ('queued','running','completed')",
	)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var body string
		if err := rows.Scan(&body); err != nil {
			return false, err
		}
		var job Job
		if json.Unmarshal([]byte(body), &job) == nil && job.Partition == "validation" && job.Strategy.ID == strategyID {
			return true, nil
		}
	}
	return false, rows.Err()
}

// Trials counts the versions of a strategy configuration that were tried: same name, symbol and
// account. The deflated Sharpe ratio uses it to discount selection among them.
func (s Store) Trials(ctx context.Context, st Strategy) (int, error) {
	rows, err := s.DB.QueryContext(ctx, "SELECT payload FROM strategies")
	if err != nil {
		return 1, err
	}
	defer rows.Close()
	n := 0
	for rows.Next() {
		var body string
		if err := rows.Scan(&body); err != nil {
			return 1, err
		}
		var other Strategy
		if json.Unmarshal([]byte(body), &other) == nil && other.Name == st.Name && other.Symbol == st.Symbol &&
			other.AccountID == st.AccountID {
			n++
		}
	}
	return max(n, 1), rows.Err()
}
