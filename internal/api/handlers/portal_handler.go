package handlers

import "github.com/gin-gonic/gin"

type PortalHandler struct{}

func NewPortalHandler() *PortalHandler { return &PortalHandler{} }

func (h *PortalHandler) CreateToken(c *gin.Context)          { NotImplemented(c) }
func (h *PortalHandler) Home(c *gin.Context)                 { NotImplemented(c) }
func (h *PortalHandler) Cancel(c *gin.Context)               { NotImplemented(c) }
func (h *PortalHandler) UpdatePaymentMethod(c *gin.Context)  { NotImplemented(c) }
