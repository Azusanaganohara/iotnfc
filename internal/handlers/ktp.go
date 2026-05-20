package handlers

import (
	"net/http"
	"strconv"

	"iot-ktp-api/internal/models"
	"iot-ktp-api/internal/services"
	"iot-ktp-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type KTPHandler struct {
	svc *services.KTPService
}

func NewKTPHandler(svc *services.KTPService) *KTPHandler {
	return &KTPHandler{svc: svc}
}

func (h *KTPHandler) Tap(c *gin.Context) {
	var input services.TapInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "Validation error", err.Error())
		return
	}

	deviceRaw, exists := c.Get("device")
	if !exists {
		utils.ResponseError(c, http.StatusInternalServerError, "Device context missing", "")
		return
	}
	device := deviceRaw.(*models.IotDevice)

	result, err := h.svc.ProcessTap(device, input)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "Tap processing failed", err.Error())
		return
	}

	statusCode := http.StatusOK
	if result.Action == string(models.ActionRegistered) {
		statusCode = http.StatusCreated
	}

	utils.ResponseOK(c, statusCode, result.Message, result)
}

func (h *KTPHandler) GetLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	nodeID := c.Query("node_id")
	unixID := c.Query("unix_id")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	logs, total, err := h.svc.GetLogs(nodeID, unixID, page, limit)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "Failed to fetch logs", err.Error())
		return
	}

	utils.ResponseOK(c, http.StatusOK, "Access logs", gin.H{
		"logs": logs,
		"pagination": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

func (h *KTPHandler) GetLogsByDevice(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	logs, total, err := h.svc.GetLogs(c.Param("node_id"), "", page, limit)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "Failed to fetch logs", err.Error())
		return
	}

	utils.ResponseOK(c, http.StatusOK, "Device access logs", gin.H{
		"logs":  logs,
		"total": total,
	})
}

func (h *KTPHandler) GetLogsByUnixID(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	unixID := c.Param("unix_id")

	logs, total, err := h.svc.GetLogs("", unixID, page, limit)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "Failed to fetch logs", err.Error())
		return
	}

	utils.ResponseOK(c, http.StatusOK, "Member access logs", gin.H{
		"logs":  logs,
		"total": total,
	})
}
