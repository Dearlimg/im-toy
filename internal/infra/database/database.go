package database

import (
	"database/sql"
	"im-chat/internal/infra/config"
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shirou/gopsutil/mem"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB
var (
	dbPoolStats     *prometheus.GaugeVec
	poolRegistry    *prometheus.Registry
	initMetricsOnce sync.Once
)

func InitDB(cfg *config.DatabaseConfig) error {
	var err error

	DB, err = gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	maxIdle := cfg.MaxIdleConns
	if maxIdle <= 0 {
		maxIdle = 10
	}
	maxOpen := cfg.MaxOpenConns
	if maxOpen <= 0 {
		maxOpen = 100
	}

	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	initMetricsOnce.Do(initMetrics)

	go startPoolMonitor(sqlDB)

	logrus.Info("Database connected successfully")
	return nil
}

func GetDB() *gorm.DB {
	return DB
}

func initMetrics() {
	poolRegistry = prometheus.NewRegistry()

	dbPoolStats = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "db_pool_stats",
			Help: "Database connection pool statistics",
		},
		[]string{"type"},
	)

	poolRegistry.MustRegister(dbPoolStats)
	poolRegistry.MustRegister(prometheus.NewGoCollector())
	poolRegistry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
}

func startPoolMonitor(sqlDB *sql.DB) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		stats := sqlDB.Stats()
		dbPoolStats.WithLabelValues("in_use").Set(float64(stats.InUse))
		dbPoolStats.WithLabelValues("idle").Set(float64(stats.Idle))
		dbPoolStats.WithLabelValues("open_connections").Set(float64(stats.OpenConnections))
		dbPoolStats.WithLabelValues("wait_count").Set(float64(stats.WaitCount))
		dbPoolStats.WithLabelValues("wait_duration_ms").Set(float64(stats.WaitDuration.Milliseconds()))

		v, _ := mem.VirtualMemory()
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		logrus.WithFields(logrus.Fields{
			"db_pool": logrus.Fields{
				"open_connections": stats.OpenConnections,
				"in_use":           stats.InUse,
				"idle":             stats.Idle,
				"wait_count":       stats.WaitCount,
				"wait_duration_ms": stats.WaitDuration.Milliseconds(),
			},
			"memory": logrus.Fields{
				"sys_mb":      v.Total / 1024 / 1024 / 1024,
				"used_mb":     v.Used / 1024 / 1024 / 1024,
				"go_alloc_mb": m.Alloc / 1024 / 1024,
				"go_sys_mb":   m.Sys / 1024 / 1024,
				"go_gc_count": m.NumGC,
			},
		}).Info("system_stats")
	}
}

func PoolStatsHandler() gin.HandlerFunc {
	h := promhttp.HandlerFor(poolRegistry, promhttp.HandlerOpts{})
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
