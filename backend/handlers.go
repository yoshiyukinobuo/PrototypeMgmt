package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
)

// ---- helpers ----

// queryMaps always uses a prepared statement so MySQL returns typed values
// (numbers stay numbers in the JSON).
func queryMaps(db *sql.DB, q string, args ...any) ([]map[string]any, error) {
	stmt, err := db.Prepare(q)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	out := []map[string]any{}
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		m := make(map[string]any, len(cols))
		for i, c := range cols {
			if b, ok := vals[i].([]byte); ok {
				m[c] = string(b)
			} else {
				m[c] = vals[i]
			}
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Server) list(c *gin.Context, q string, args ...any) {
	out, err := queryMaps(s.db, q, args...)
	if err != nil {
		serverErr(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func serverErr(c *gin.Context, err error) {
	log.Printf("error: %v", err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
}

func badReq(c *gin.Context, msg string) { c.JSON(http.StatusBadRequest, gin.H{"error": msg}) }

// dbErr maps common MySQL errors to useful HTTP statuses.
func dbErr(c *gin.Context, err error) {
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		switch me.Number {
		case 1062:
			c.JSON(http.StatusConflict, gin.H{"error": "already exists"})
			return
		case 1452:
			badReq(c, "referenced record does not exist")
			return
		}
	}
	serverErr(c, err)
}

func idParam(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		badReq(c, "invalid id")
		return 0, false
	}
	return id, true
}

func (s *Server) logSync(dir, entity, status, msg string) {
	if _, err := s.db.Exec(`INSERT INTO erp_sync_log (direction, entity, status, message) VALUES (?,?,?,?)`, dir, entity, status, msg); err != nil {
		log.Printf("sync log: %v", err)
	}
}

// ---- dashboard ----

func (s *Server) dashboard(c *gin.Context) {
	res := gin.H{}
	count := func(key, q string) {
		var n int
		if err := s.db.QueryRow(q).Scan(&n); err != nil {
			log.Printf("dashboard %s: %v", key, err)
		}
		res[key] = n
	}
	count("open_requests", `SELECT COUNT(*) FROM prototype_requests WHERE status IN ('REQUESTED','APPROVED')`)
	count("active_lots", `SELECT COUNT(*) FROM prototype_lots WHERE stage <> 'CLOSED'`)
	count("open_issues", `SELECT COUNT(*) FROM issues WHERE status IN ('OPEN','ANALYZING')`)
	count("failed_units", `SELECT COUNT(*) FROM units WHERE status = 'FAIL'`)
	stages, err := queryMaps(s.db, `SELECT stage, COUNT(*) AS n FROM prototype_lots GROUP BY stage`)
	if err != nil {
		serverErr(c, err)
		return
	}
	res["lots_by_stage"] = stages
	c.JSON(http.StatusOK, res)
}

// ---- requests ----

func (s *Server) listRequests(c *gin.Context) {
	s.list(c, `SELECT * FROM prototype_requests ORDER BY id DESC`)
}

func (s *Server) createRequest(c *gin.Context) {
	var in struct {
		Title       string  `json:"title" binding:"required,max=200"`
		ProductType string  `json:"product_type" binding:"required,oneof=GPS ANTENNA OTHER"`
		Customer    *string `json:"customer"`
		Quantity    int     `json:"quantity" binding:"required,min=1,max=100000"`
		DueDate     *string `json:"due_date"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		badReq(c, err.Error())
		return
	}
	if in.DueDate != nil && *in.DueDate == "" {
		in.DueDate = nil
	}
	res, err := s.db.Exec(`INSERT INTO prototype_requests (title, product_type, customer, quantity, due_date) VALUES (?,?,?,?,?)`,
		in.Title, in.ProductType, in.Customer, in.Quantity, in.DueDate)
	if err != nil {
		dbErr(c, err)
		return
	}
	id, _ := res.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (s *Server) setRequestStatus(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	var in struct {
		Status string `json:"status" binding:"required,oneof=REQUESTED APPROVED IN_BUILD EVALUATING DONE REJECTED"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		badReq(c, err.Error())
		return
	}
	if _, err := s.db.Exec(`UPDATE prototype_requests SET status=? WHERE id=?`, in.Status, id); err != nil {
		dbErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ---- lots & units ----

func (s *Server) listLots(c *gin.Context) {
	s.list(c, `SELECT l.*, r.title AS request_title,
	  (SELECT COUNT(*) FROM units u WHERE u.lot_id = l.id) AS unit_count,
	  (SELECT COUNT(*) FROM units u WHERE u.lot_id = l.id AND u.status = 'FAIL') AS fail_count
	  FROM prototype_lots l LEFT JOIN prototype_requests r ON r.id = l.request_id
	  ORDER BY l.id DESC`)
}

func (s *Server) createLot(c *gin.Context) {
	var in struct {
		LotNo       string  `json:"lot_no" binding:"required,max=40"`
		RequestID   *int64  `json:"request_id"`
		ItemCode    *string `json:"item_code"`
		BOMRevision string  `json:"bom_revision" binding:"max=20"`
		Quantity    int     `json:"quantity" binding:"required,min=1,max=1000"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		badReq(c, err.Error())
		return
	}
	if in.ItemCode != nil && *in.ItemCode == "" {
		in.ItemCode = nil
	}
	tx, err := s.db.Begin()
	if err != nil {
		serverErr(c, err)
		return
	}
	defer tx.Rollback()
	res, err := tx.Exec(`INSERT INTO prototype_lots (lot_no, request_id, item_code, bom_revision, quantity) VALUES (?,?,?,?,?)`,
		in.LotNo, in.RequestID, in.ItemCode, in.BOMRevision, in.Quantity)
	if err != nil {
		dbErr(c, err)
		return
	}
	lotID, _ := res.LastInsertId()
	for i := 1; i <= in.Quantity; i++ { // one serial number per unit
		if _, err := tx.Exec(`INSERT INTO units (serial_no, lot_id) VALUES (?,?)`, fmt.Sprintf("%s-%03d", in.LotNo, i), lotID); err != nil {
			dbErr(c, err)
			return
		}
	}
	if err := tx.Commit(); err != nil {
		serverErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": lotID})
}

func (s *Server) setLotStage(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	var in struct {
		Stage string `json:"stage" binding:"required,oneof=PLANNED BUILD INSPECTION SHIPPED EVALUATION CLOSED"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		badReq(c, err.Error())
		return
	}
	if _, err := s.db.Exec(`UPDATE prototype_lots SET stage=? WHERE id=?`, in.Stage, id); err != nil {
		dbErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Server) listUnits(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	s.list(c, `SELECT * FROM units WHERE lot_id=? ORDER BY serial_no`, id)
}

func (s *Server) listTests(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	s.list(c, `SELECT * FROM test_results WHERE unit_id=? ORDER BY id DESC`, id)
}

func (s *Server) addTest(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	var in struct {
		TestName string   `json:"test_name" binding:"required,max=100"`
		Value    *float64 `json:"value"`
		Unit     string   `json:"unit" binding:"max=20"`
		SpecMin  *float64 `json:"spec_min"`
		SpecMax  *float64 `json:"spec_max"`
		Passed   *bool    `json:"passed"`
		FileURL  string   `json:"file_url" binding:"max=500"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		badReq(c, err.Error())
		return
	}
	var passed bool
	switch {
	case in.Value != nil && (in.SpecMin != nil || in.SpecMax != nil):
		passed = (in.SpecMin == nil || *in.Value >= *in.SpecMin) && (in.SpecMax == nil || *in.Value <= *in.SpecMax)
	case in.Passed != nil:
		passed = *in.Passed
	default:
		badReq(c, "give a value with spec limits, or mark pass/fail")
		return
	}
	tx, err := s.db.Begin()
	if err != nil {
		serverErr(c, err)
		return
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`INSERT INTO test_results (unit_id, test_name, value, unit, spec_min, spec_max, passed, file_url) VALUES (?,?,?,?,?,?,?,?)`,
		id, in.TestName, in.Value, in.Unit, in.SpecMin, in.SpecMax, passed, in.FileURL); err != nil {
		dbErr(c, err)
		return
	}
	q := `UPDATE units SET status='FAIL' WHERE id=?`
	if passed {
		q = `UPDATE units SET status='PASS' WHERE id=? AND status IN ('BUILT','PASS')`
	}
	if _, err := tx.Exec(q, id); err != nil {
		serverErr(c, err)
		return
	}
	if err := tx.Commit(); err != nil {
		serverErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"passed": passed})
}

// ---- issues ----

func (s *Server) listIssues(c *gin.Context) {
	s.list(c, `SELECT i.*, l.lot_no FROM issues i LEFT JOIN prototype_lots l ON l.id = i.lot_id ORDER BY i.id DESC`)
}

func (s *Server) createIssue(c *gin.Context) {
	var in struct {
		LotID          *int64  `json:"lot_id"`
		Title          string  `json:"title" binding:"required,max=200"`
		Description    *string `json:"description"`
		Severity       string  `json:"severity" binding:"required,oneof=LOW MEDIUM HIGH"`
		Countermeasure *string `json:"countermeasure"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		badReq(c, err.Error())
		return
	}
	res, err := s.db.Exec(`INSERT INTO issues (lot_id, title, description, severity, countermeasure) VALUES (?,?,?,?,?)`,
		in.LotID, in.Title, in.Description, in.Severity, in.Countermeasure)
	if err != nil {
		dbErr(c, err)
		return
	}
	id, _ := res.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (s *Server) setIssueStatus(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	var in struct {
		Status string `json:"status" binding:"required,oneof=OPEN ANALYZING FIXED CLOSED"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		badReq(c, err.Error())
		return
	}
	if _, err := s.db.Exec(`UPDATE issues SET status=? WHERE id=?`, in.Status, id); err != nil {
		dbErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ---- ERP ----

func (s *Server) listItems(c *gin.Context) {
	s.list(c, `SELECT * FROM erp_items ORDER BY item_code`)
}

func (s *Server) syncLog(c *gin.Context) {
	s.list(c, `SELECT * FROM erp_sync_log ORDER BY id DESC LIMIT 50`)
}

func (s *Server) erpSync(c *gin.Context) {
	items, err := s.erp.ListItems(c.Request.Context())
	if err != nil {
		s.logSync("IN", "Item", "ERROR", err.Error())
		c.JSON(http.StatusBadGateway, gin.H{"error": "ERP request failed; see the sync log"})
		return
	}
	for _, it := range items {
		if _, err := s.db.Exec(`INSERT INTO erp_items (item_code, item_name, item_group) VALUES (?,?,?)
		  ON DUPLICATE KEY UPDATE item_name=VALUES(item_name), item_group=VALUES(item_group), synced_at=CURRENT_TIMESTAMP`,
			it.Code, it.Name, it.Group); err != nil {
			s.logSync("IN", "Item", "ERROR", err.Error())
			serverErr(c, err)
			return
		}
	}
	s.logSync("IN", "Item", "OK", fmt.Sprintf("%d items synced", len(items)))
	c.JSON(http.StatusOK, gin.H{"synced": len(items)})
}

func (s *Server) erpWorkOrder(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	var lotNo string
	var item, wo sql.NullString
	var qty int
	err := s.db.QueryRow(`SELECT lot_no, item_code, quantity, erp_work_order FROM prototype_lots WHERE id=?`, id).Scan(&lotNo, &item, &qty, &wo)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "lot not found"})
		return
	}
	if err != nil {
		serverErr(c, err)
		return
	}
	if !item.Valid {
		badReq(c, "set an ERP item on the lot first")
		return
	}
	if wo.Valid {
		c.JSON(http.StatusConflict, gin.H{"error": "work order already created: " + wo.String})
		return
	}
	name, err := s.erp.CreateWorkOrder(c.Request.Context(), WorkOrder{ItemCode: item.String, Qty: qty, LotNo: lotNo})
	if err != nil {
		s.logSync("OUT", "Work Order", "ERROR", err.Error())
		c.JSON(http.StatusBadGateway, gin.H{"error": "ERP request failed; see the sync log"})
		return
	}
	if _, err := s.db.Exec(`UPDATE prototype_lots SET erp_work_order=? WHERE id=?`, name, id); err != nil {
		serverErr(c, err)
		return
	}
	s.logSync("OUT", "Work Order", "OK", lotNo+" -> "+name)
	c.JSON(http.StatusCreated, gin.H{"work_order": name})
}

// メモ 002: 申請から完了までの状態遷移を確認。画面とAPIのデータ項目を照合。

// メモ 004: ロット数量に応じた個体生成の前提を確認。関連テーブルとの整合性を確認。

// メモ 005: 連番の桁数とロット番号の組み合わせを確認。状態値の表記を統一。

// メモ 007: 試験名と測定値の登録項目を確認。登録後の再読み込み手順を確認。

// メモ 008: 上下限値と手動判定の扱いを確認。未登録データの表示を確認。

// メモ 012: ERP品目マスターの取得項目を確認。READMEの説明と実装を照合。

// メモ 022: 申請から完了までの状態遷移を確認。設定値の既定動作を確認。

// メモ 024: ロット数量に応じた個体生成の前提を確認。READMEの説明と実装を照合。

// メモ 025: 連番の桁数とロット番号の組み合わせを確認。入力値の境界条件を確認。

// メモ 027: 試験名と測定値の登録項目を確認。エラー時の利用者向け表示を確認。

// メモ 028: 上下限値と手動判定の扱いを確認。関連テーブルとの整合性を確認。

// メモ 032: ERP品目マスターの取得項目を確認。未登録データの表示を確認。

// メモ 042: 申請から完了までの状態遷移を確認。一覧表示の並び順を確認。

// メモ 044: ロット数量に応じた個体生成の前提を確認。未登録データの表示を確認。

// メモ 045: 連番の桁数とロット番号の組み合わせを確認。APIレスポンスとの対応を確認。

// メモ 047: 試験名と測定値の登録項目を確認。運用時の注意点を整理。

// メモ 048: 上下限値と手動判定の扱いを確認。READMEの説明と実装を照合。

// メモ 052: ERP品目マスターの取得項目を確認。関連テーブルとの整合性を確認。

// メモ 062: 申請から完了までの状態遷移を確認。画面とAPIのデータ項目を照合。

// メモ 064: ロット数量に応じた個体生成の前提を確認。関連テーブルとの整合性を確認。
