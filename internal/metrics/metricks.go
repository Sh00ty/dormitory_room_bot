package metric

import (
	"net/http"
	"time"

	"github.com/Sh00ty/dormitory_room_bot/internal/logger"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const prometheusPost = ":9000"

func init() {

	err := prometheus.Register(TotalSheduledTasks)
	if err != nil {
		logger.Errorf("cant register metric: %s", err)
	}
	err = prometheus.Register(TotalCreatedTasks)
	if err != nil {
		logger.Errorf("cant register metric: %s", err)
	}
	err = prometheus.Register(TotalSuccesfulPostgresRequests)
	if err != nil {
		logger.Errorf("cant register metric: %s", err)
	}

	err = prometheus.Register(TotalPostgresRequests)
	if err != nil {
		logger.Errorf("cant register metric: %s", err)
	}
	err = prometheus.Register(PutExecutingTime)
	if err != nil {
		logger.Errorf("cant register metric: %s", err)
	}
	err = prometheus.Register(TotalCredidSum)
	if err != nil {
		logger.Errorf("cant register metric: %s", err)
	}
	err = prometheus.Register(TotalCreditsCommands)
	if err != nil {
		logger.Errorf("cant register metric: %s", err)
	}
	err = prometheus.Register(TotalListCommands)
	if err != nil {
		logger.Errorf("cant register metric: %s", err)
	}

	go func() {
		router := mux.NewRouter()
		router.Path("/metrics").Handler(promhttp.Handler())
		srv := http.Server{
			ReadHeaderTimeout: time.Second,
			ReadTimeout:       time.Second,
			IdleTimeout:       2 * time.Minute,
			Addr:              prometheusPost,
			Handler:           router,
		}
		srv.SetKeepAlivesEnabled(true)
		err := srv.ListenAndServe()
		if err != nil {
			logger.Fataf(err.Error())
		}
	}()
}

var (
	TotalCredidSum = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "total_credit_sum", Help: "Sum of all credits provided with bot"}, []string{"channel_id"})

	TotalCreditsCommands = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "total_number_of_credit_command"}, []string{"command"})

	TotalListCommands = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "total_number_of_lists_commands"}, []string{"command"})

	TotalSheduledTasks = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "my_sheduled_tasks", Help: "Number of sheduled task"}, []string{"task_id"})

	TotalCreatedTasks = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "my_succsessfuly_created_tasks", Help: "Number of succsesfully created tasks"}, []string{"task_id"})

	TotalSuccesfulPostgresRequests = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "my_postgres_succses_count", Help: "Number of postgres requests withot any error"}, []string{"request"})

	TotalPostgresRequests = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "my_postgres_count", Help: "Number of total postgres requests"}, []string{"request"})

	UnsuccsessfulGetInRecaller = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "unsucsesful_get_from_redis", Help: "Number of unsecsesful get from redis in recaller"}, []string{""})

	PutExecutingTime = prometheus.NewHistogram(prometheus.HistogramOpts{Name: "wp_put_executing_time", Help: "time of puting job to worker pool", Buckets: PutExecutingTimeBucket})
)

var (
	// nanoseconds
	PutExecutingTimeBucket = []float64{0, 10, 20, 30, 90, 100, 1000, 2000}
)
