package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ERPItem struct{ Code, Name, Group string }

type WorkOrder struct {
	ItemCode string
	Qty      int
	LotNo    string
}

// ERPClient is the only thing the rest of the app knows about the ERP.
// Add another implementation (SAP, CSV exchange, ...) without touching handlers.
type ERPClient interface {
	ListItems(ctx context.Context) ([]ERPItem, error)
	CreateWorkOrder(ctx context.Context, wo WorkOrder) (string, error)
}

// MockERP lets you run the whole system without an ERP.
type MockERP struct{}

func (MockERP) ListItems(context.Context) ([]ERPItem, error) {
	return []ERPItem{
		{"GPS-MOD-001", "GPS module board", "Electronics"},
		{"ANT-PATCH-02", "Patch antenna element", "Antenna"},
		{"HSG-ABS-10", "ABS housing", "Mechanical"},
	}, nil
}

func (MockERP) CreateWorkOrder(_ context.Context, wo WorkOrder) (string, error) {
	return "MOCK-WO-" + wo.LotNo, nil
}

type ERPNext struct {
	base, key, secret, company string
	http                       *http.Client
}

func NewERPNext(base, key, secret, company string) *ERPNext {
	return &ERPNext{
		base: strings.TrimRight(base, "/"), key: key, secret: secret, company: company,
		http: &http.Client{Timeout: 20 * time.Second},
	}
}

func (e *ERPNext) do(ctx context.Context, method, path string, body, out any) error {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, e.base+path, rdr)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "token "+e.key+":"+e.secret)
	req.Header.Set("Content-Type", "application/json")
	resp, err := e.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("erpnext %s %s: %d %s", method, path, resp.StatusCode, b)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (e *ERPNext) ListItems(ctx context.Context) ([]ERPItem, error) {
	q := url.Values{}
	q.Set("fields", `["name","item_name","item_group"]`)
	q.Set("limit_page_length", "0")
	var out struct {
		Data []struct {
			Name      string `json:"name"`
			ItemName  string `json:"item_name"`
			ItemGroup string `json:"item_group"`
		} `json:"data"`
	}
	if err := e.do(ctx, http.MethodGet, "/api/resource/Item?"+q.Encode(), nil, &out); err != nil {
		return nil, err
	}
	items := make([]ERPItem, 0, len(out.Data))
	for _, d := range out.Data {
		items = append(items, ERPItem{d.Name, d.ItemName, d.ItemGroup})
	}
	return items, nil
}

func (e *ERPNext) CreateWorkOrder(ctx context.Context, wo WorkOrder) (string, error) {
	body := map[string]any{
		"production_item": wo.ItemCode,
		"qty":             wo.Qty,
		"company":         e.company,
		"description":     "Prototype lot " + wo.LotNo,
	}
	var out struct {
		Data struct {
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := e.do(ctx, http.MethodPost, "/api/resource/Work%20Order", body, &out); err != nil {
		return "", err
	}
	return out.Data.Name, nil
}

// メモ 011: ERPNextとの接続設定と認証情報を確認。運用時の注意点を整理。

// メモ 013: ロット単位の作業指示作成項目を確認。入力値の境界条件を確認。

// メモ 031: ERPNextとの接続設定と認証情報を確認。登録後の再読み込み手順を確認。

// メモ 033: ロット単位の作業指示作成項目を確認。APIレスポンスとの対応を確認。
