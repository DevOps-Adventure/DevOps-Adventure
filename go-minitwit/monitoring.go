package main

import (
	"database/sql"
	"log"
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
)

var logger *logrus.Logger

// connect lorus with tcp to fluent
func setupLogger() {
	logger = logrus.New()

	// Configure the Fluentd hook.
	hook, err := logrusfluent.NewWithConfig(logrusfluent.Config{
		Port: 24224,
		Host: "fluentd",
	})
	if err != nil {
		logger.Fatalf("Failed to create Fluentd hook: %v", err)
	}

	logger.SetLevel(logrus.DebugLevel)
	logger.AddHook(hook)

	hook.SetTag("minitwit.tag")
	hook.SetMessageField("message")
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

func UserSignupMonitoring() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if the registration was successful
		if success, exists := c.Get("registrationSuccess"); exists && success.(bool) {
			newSignupsCounter.Inc()
		}
	}
}

////
