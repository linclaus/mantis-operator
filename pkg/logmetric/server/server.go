package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.moebius.com/mantis/pkg/logmetric/db"
	"gitlab.moebius.com/mantis/pkg/logmetric/metric"
	"gitlab.moebius.com/mantis/pkg/logmetric/model"
)

type Server struct {
	r                *mux.Router
	elasticMetricMap *model.ElasticMetricMap
	db               db.Storer
	debug            bool
}

func New(debug bool, db db.Storer, elasticMetricMap *model.ElasticMetricMap) Server {
	r := mux.NewRouter()
	s := Server{
		debug:            debug,
		r:                r,
		db:               db,
		elasticMetricMap: elasticMetricMap,
	}
	r.Handle("/metrics", s.metricHandler(promhttp.Handler()))
	r.HandleFunc("/metric/{id}", s.CreateStrategyMetric).Methods("POST")
	r.HandleFunc("/metric/{id}", s.GetStrategyMetric).Methods("GET")
	r.HandleFunc("/metric/{id}", s.UpdateStrategyMetric).Methods("PUT")
	r.HandleFunc("/metric/{id}", s.DeleteStrategyMetric).Methods("DELETE")
	return s
}

// Start starts a new server on the given address
func (s Server) Start(address string) {
	log.Println("Starting listener on", address)
	log.Fatal(http.ListenAndServe(address, s.r))
}

func (s Server) metricHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("before metric handler")
		s.elasticMetricMap.Range(func(k string, v *model.StrategyMetric) bool {
			em := s.db.GetMetric(*v)
			metric.ElasticMetricCountVec.WithLabelValues(em.StrategyId).Set(em.Count)
			return true
		})
		next.ServeHTTP(w, r)
		log.Println("after metric handler")
	})

}

func (s Server) handleFuncInterceptor(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("before handlerFunc")
		h(w, r)
		log.Println("after handlerFunc")
	}
}

func (s Server) GetStrategyMetric(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sm := s.elasticMetricMap.Get(vars["id"])
	fmt.Println(sm)
}

func (s Server) CreateStrategyMetric(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	vars := mux.Vars(r)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read payload: %s\n", err)
		http.Error(w, fmt.Sprintf("Failed to read payload: %s", err), http.StatusBadRequest)
		return
	}

	if s.debug {
		log.Println("Received webhook payload", string(body))
	}
	smr := &model.StrategyMetricRequest{}
	json.Unmarshal([]byte(body), smr)
	smr.StrategyId = vars["id"]
	sm := s.elasticMetricMap.Get(smr.StrategyId)
	if sm == nil {
		sm = &model.StrategyMetric{
			StrategyId: smr.StrategyId,
			Dsl:        smr.Dsl,
		}
		s.elasticMetricMap.Set(smr.StrategyId, sm)
	}
}

func (s Server) UpdateStrategyMetric(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	vars := mux.Vars(r)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read payload: %s\n", err)
		http.Error(w, fmt.Sprintf("Failed to read payload: %s", err), http.StatusBadRequest)
		return
	}

	if s.debug {
		log.Println("Received webhook payload", string(body))
	}
	smr := &model.StrategyMetricRequest{}
	json.Unmarshal([]byte(body), smr)
	smr.StrategyId = vars["id"]
	sm := s.elasticMetricMap.Get(smr.StrategyId)
	if sm != nil {
		if sm.Dsl != smr.Dsl || sm.StrategyId != smr.StrategyId {
			metric.ElasticMetricCountVec.DeleteLabelValues(sm.StrategyId)
		}
	}
	sm = &model.StrategyMetric{
		StrategyId: smr.StrategyId,
		Dsl:        smr.Dsl,
	}
	s.elasticMetricMap.Set(smr.StrategyId, sm)
}

func (s Server) DeleteStrategyMetric(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	em := s.elasticMetricMap.Get(vars["id"])
	if em != nil {
		metric.ElasticMetricCountVec.DeleteLabelValues(em.StrategyId)
		s.elasticMetricMap.Delete(em.StrategyId)
	}
}
