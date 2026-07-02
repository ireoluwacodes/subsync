package handlers

import "github.com/gin-gonic/gin"

type AnalyticsHandler struct{}

func NewAnalyticsHandler() *AnalyticsHandler { return &AnalyticsHandler{} }

func (h *AnalyticsHandler) MRR(c *gin.Context)     { NotImplemented(c) }
func (h *AnalyticsHandler) Churn(c *gin.Context)   { NotImplemented(c) }
func (h *AnalyticsHandler) Dunning(c *gin.Context) { NotImplemented(c) }
func (h *AnalyticsHandler) Revenue(c *gin.Context) { NotImplemented(c) }
