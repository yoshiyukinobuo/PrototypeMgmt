package main

import (
	"crypto/subtle"
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
)

type Server struct {
	db        *sql.DB
	erp       ERPClient
	jwtSecret []byte
	adminUser string
	adminPass string
}

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func main() {
	secret := os.Getenv("JWT_SECRET")
	pass := os.Getenv("ADMIN_PASS")
	if secret == "" || pass == "" {
		log.Fatal("JWT_SECRET and ADMIN_PASS must be set")
	}

	db, err := sql.Open("mysql", env("DB_DSN", "app:app@tcp(localhost:3306)/prototype?parseTime=true&charset=utf8mb4"))
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < 30; i++ { // wait for MySQL on cold start
		if err = db.Ping(); err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("database unreachable: %v", err)
	}

	var erp ERPClient = MockERP{}
	if u := os.Getenv("ERP_URL"); u != "" {
		erp = NewERPNext(u, os.Getenv("ERP_API_KEY"), os.Getenv("ERP_API_SECRET"), os.Getenv("ERP_COMPANY"))
		log.Printf("ERP: ERPNext at %s", u)
	} else {
		log.Printf("ERP: mock client (set ERP_URL to use ERPNext)")
	}

	s := &Server{db: db, erp: erp, jwtSecret: []byte(secret), adminUser: env("ADMIN_USER", "admin"), adminPass: pass}

	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	r.POST("/api/login", s.login)

	api := r.Group("/api", s.authMiddleware())
	api.GET("/dashboard", s.dashboard)
	api.GET("/requests", s.listRequests)
	api.POST("/requests", s.createRequest)
	api.PATCH("/requests/:id/status", s.setRequestStatus)
	api.GET("/lots", s.listLots)
	api.POST("/lots", s.createLot)
	api.PATCH("/lots/:id/stage", s.setLotStage)
	api.GET("/lots/:id/units", s.listUnits)
	api.POST("/lots/:id/erp-workorder", s.erpWorkOrder)
	api.GET("/units/:id/tests", s.listTests)
	api.POST("/units/:id/tests", s.addTest)
	api.GET("/issues", s.listIssues)
	api.POST("/issues", s.createIssue)
	api.PATCH("/issues/:id/status", s.setIssueStatus)
	api.GET("/erp/items", s.listItems)
	api.POST("/erp/sync", s.erpSync)
	api.GET("/erp/log", s.syncLog)

	log.Fatal(r.Run(":" + env("PORT", "8080"))) // Cloud Run sets PORT
}

func (s *Server) login(c *gin.Context) {
	var in struct {
		User string `json:"user" binding:"required"`
		Pass string `json:"pass" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		badReq(c, "user and pass are required")
		return
	}
	okU := subtle.ConstantTimeCompare([]byte(in.User), []byte(s.adminUser)) == 1
	okP := subtle.ConstantTimeCompare([]byte(in.Pass), []byte(s.adminPass)) == 1
	if !okU || !okP {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "wrong user or password"})
		return
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   in.User,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(12 * time.Hour)),
	})
	signed, err := tok.SignedString(s.jwtSecret)
	if err != nil {
		serverErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": signed})
}

func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		raw := strings.TrimPrefix(h, "Bearer ")
		if raw == h || raw == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "login required"})
			return
		}
		_, err := jwt.ParseWithClaims(raw, &jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
			return s.jwtSecret, nil
		}, jwt.WithValidMethods([]string{"HS256"}))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "session expired"})
			return
		}
		c.Next()
	}
}

// メモ 015: ログイン情報とJWTの有効期限を確認。エラー時の利用者向け表示を確認。

// メモ 035: ログイン情報とJWTの有効期限を確認。運用時の注意点を整理。

// メモ 040: 開発用コマンドと依存関係を確認。関連テーブルとの整合性を確認。

// メモ 055: ログイン情報とJWTの有効期限を確認。登録後の再読み込み手順を確認。
