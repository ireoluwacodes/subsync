package handlers

import "github.com/gin-gonic/gin"

type WebhookHandler struct{}

func NewWebhookHandler() *WebhookHandler { return &WebhookHandler{} }

func (h *WebhookHandler) CreateEndpoint(c *gin.Context)  { NotImplemented(c) }
func (h *WebhookHandler) ListEndpoints(c *gin.Context)   { NotImplemented(c) }
func (h *WebhookHandler) GetEndpoint(c *gin.Context)     { NotImplemented(c) }
func (h *WebhookHandler) UpdateEndpoint(c *gin.Context)  { NotImplemented(c) }
func (h *WebhookHandler) DeleteEndpoint(c *gin.Context)  { NotImplemented(c) }
