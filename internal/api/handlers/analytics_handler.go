package handlers

import "github.com/gin-gonic/gin"

type AnalyticsHandler struct{}

func NewAnalyticsHandler() *AnalyticsHandler { return &AnalyticsHandler{} }

func (h *AnalyticsHandler) Register(rg *gin.RouterGroup) {
	rg.GET("/analytics/mrr", NotImplemented)
	rg.GET("/analytics/churn", NotImplemented)
	rg.GET("/analytics/dunning", NotImplemented)
	rg.GET("/analytics/revenue", NotImplemented)
}
