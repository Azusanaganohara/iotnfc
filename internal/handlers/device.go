package handlers

import (
	"net/http"

	"iot-ktp-api/internal/services"
	"iot-ktp-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type DeviceHandler struct {
	svc *services.DeviceService
}

func NewDeviceHandler(svc *services.DeviceService) *DeviceHandler {
	return &DeviceHandler{svc: svc}
}

func (h *DeviceHandler) Provision(c *gin.Context) {
	var input services.ProvisionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "Validation error", err.Error())
		return
	}

	result, err := h.svc.Provision(input)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "Provisioning failed", err.Error())
		return
	}

	utils.ResponseOK(c, http.StatusCreated, result.Message, result)
}

func (h *DeviceHandler) GetMyStatus(c *gin.Context) {
	nodeID, _ := c.Get("device_node_id")
	device, err := h.svc.GetByNodeID(nodeID.(string))
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, "Device not found", err.Error())
		return
	}

	utils.ResponseOK(c, http.StatusOK, "Device status", gin.H{
		"node_id":      device.NodeID,
		"device_name":  device.DeviceName,
		"mode":         device.Mode,
		"location":     device.Location,
		"last_seen_at": device.LastSeenAt,
	})
}

func (h *DeviceHandler) GetAll(c *gin.Context) {
	devices, err := h.svc.GetAll()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "Failed to fetch devices", err.Error())
		return
	}
	utils.ResponseOK(c, http.StatusOK, "Device list", gin.H{
		"devices": devices,
		"total":   len(devices),
	})
}

func (h *DeviceHandler) GetOne(c *gin.Context) {
	nodeID := c.Param("node_id")
	device, err := h.svc.GetByNodeID(nodeID)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, "Device not found", err.Error())
		return
	}
	utils.ResponseOK(c, http.StatusOK, "Device detail", device)
}

func (h *DeviceHandler) Update(c *gin.Context) {
	nodeID := c.Param("node_id")
	var body struct {
		DeviceName string `json:"device_name"`
		Location   string `json:"location"`
	}
	c.ShouldBindJSON(&body)

	device, err := h.svc.UpdateDevice(nodeID, body.DeviceName, body.Location)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, "Update failed", err.Error())
		return
	}
	utils.ResponseOK(c, http.StatusOK, "Device updated", device)
}

func (h *DeviceHandler) SetMode(c *gin.Context) {
	nodeID := c.Param("node_id")
	userID, _ := c.Get("user_id")

	var input services.SetModeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "Validation error", err.Error())
		return
	}

	device, err := h.svc.SetMode(nodeID, userID.(string), input)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "Failed to set mode", err.Error())
		return
	}

	utils.ResponseOK(c, http.StatusOK, "Device mode updated to "+input.Mode, device)
}

func (h *DeviceHandler) Delete(c *gin.Context) {
	nodeID := c.Param("node_id")
	if err := h.svc.Delete(nodeID); err != nil {
		utils.ResponseError(c, http.StatusNotFound, "Delete failed", err.Error())
		return
	}
	utils.ResponseOK(c, http.StatusOK, "Device deleted", nil)
}
