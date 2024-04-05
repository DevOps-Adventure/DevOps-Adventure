package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	logrusfluent "github.com/evalphobia/logrus_fluent"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shirou/gopsutil/cpu"
	"github.com/sirupsen/logrus"
)

// defining metrics -counter,cpu,responce time monitoring for Prometeus
var (
	cpuGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "minitwit_cpu_load_percent",
		Help: "Current load of the CPU in percent.",
	})
	responseCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "minitwit_http_responses_total",
		Help: "The count of HTTP responses sent.",
	})
	requestDurationSummary = promauto.NewSummary(prometheus.SummaryOpts{
		Name: "minitwit_request_duration_milliseconds",
		Help: "Request duration distribution.",
	})

	newSignupsCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "minitwit_new_signups_total",
		Help: "Total number of new user signups.",
	})

	activeUsers = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "active_users",
		Help: "Current number of active users.",
	})

	dbProcessDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "minitwit_db_process_duration_seconds",
			Help: "Time spent in database processes, by function.",
		},
		[]string{"function"},
	)
)
var logger *logrus.Logger

// connect lorus with tcp to fluent
func setupLogger(environment string) {
	logger = logrus.New()
	if environment != "CI" && environment != "LOCAL" {
		var hook *logrusfluent.FluentHook
		var err error
		retriesLimit := 3
		delayBase := time.Second

		for i := 0; i < retriesLimit; i++ {
			hook, err = logrusfluent.NewWithConfig(logrusfluent.Config{
				Port: 24224,
				Host: "fluentd",
			})
			if err != nil {
				logger.Warnf("Failed to create Fluentd hook (Attempt %d of %d): %v", i+1, retriesLimit, err)
				time.Sleep(delayBase)
				// exp backoff
				delayBase *= 2
			} else {
				break
			}
		}
		if err != nil {
			logger.Warnf("Unable to establish connection to Fluentd after %d attempts. Log set to os stdout", retriesLimit)
			hook = nil
			logger.Out = os.Stdout
		}
		if hook != nil {
			logger.AddHook(hook)
			hook.SetTag("minitwit.tag")
			hook.SetMessageField("message")
		}
	} else {
		// in CI enviroment
		logger.Out = os.Stdout
	}
	logger.SetLevel(logrus.DebugLevel)
}

// defining registation of Prometeus
func prometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// for CPU use monitoring
func getCPUPercent() float64 {
	percentages, err := cpu.Percent(0, false)
	if err != nil {
		log.Printf("Error getting CPU usage: %v", err)
		return 0
	}
	if len(percentages) > 0 {
		return percentages[0]
	}
	return 0
}
func beforeRequestHandler(c *gin.Context) {
	// Set CPU usage
	cpuUsage := getCPUPercent()
	cpuGauge.Set(cpuUsage)
}

// for responce time monitoring and counting requests
func AfterRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		// This function needs to be deferred at the start of each handler where it's needed
		// Assuming db is the database connection passed through the context
		// And start time is set at the beginning of the handler

		defer func() {
			// Close the database connection
			if db, exists := c.Get("db"); exists {
				if dbConn, ok := db.(*sql.DB); ok {
					dbConn.Close()
				}
			}

			println("AfterRequest")

			// Increment the response counter for Prometeus
			responseCounter.Inc()

			// Calculate the elapsed time in milliseconds
			if startTime, exists := c.Get("startTime"); exists {
				if start, ok := startTime.(time.Time); ok {
					elapsedTime := time.Since(start).Milliseconds()
					requestDurationSummary.Observe(float64(elapsedTime))
				}
			}
		}()
	}

}

// prometeus cannot decrement counter so I am using gauge instead
func userLogIn() {
	activeUsers.Inc()
}

func userLogout() {
	activeUsers.Dec()
}

// this registers the dbProcessDuration mwtric with Prometheus registry.
func init() {
	prometheus.MustRegister(dbProcessDuration)
}
