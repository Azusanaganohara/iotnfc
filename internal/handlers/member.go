package handlers

import (
	"net/http"
	"strconv"

	"iot-ktp-api/internal/services"
	"iot-ktp-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type MemberHandler struct {
	svc *services.MemberService
}

func NewMemberHandler(svc *services.MemberService) *MemberHandler {
	return &MemberHandler{svc: svc}
}

func (h *MemberHandler) GetAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	members, total, err := h.svc.GetAll(page, limit, search)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "Failed to fetch members", err.Error())
		return
	}

	utils.ResponseOK(c, http.StatusOK, "Member list", gin.H{
		"members": members,
		"pagination": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

func (h *MemberHandler) Create(c *gin.Context) {
	var input services.CreateMemberInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "Validation error", err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	member, err := h.svc.Create(input, userID.(string))
	if err != nil {
		utils.ResponseError(c, http.StatusConflict, "Failed to create member", err.Error())
		return
	}

	utils.ResponseOK(c, http.StatusCreated, "Member registered successfully", member)
}

func (h *MemberHandler) GetByID(c *gin.Context) {
	member, err := h.svc.GetByID(c.Param("id"))
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, "Member not found", err.Error())
		return
	}
	utils.ResponseOK(c, http.StatusOK, "Member detail", member)
}

func (h *MemberHandler) GetByUnixID(c *gin.Context) {
	unixID := c.Param("unix_id")
	if unixID == "" {
		unixID = c.Param("nik")
	}
	member, err := h.svc.GetByUnixID(unixID)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, "Member not found", err.Error())
		return
	}
	utils.ResponseOK(c, http.StatusOK, "Member detail", member)
}

func (h *MemberHandler) Update(c *gin.Context) {
	var input services.UpdateMemberInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "Validation error", err.Error())
		return
	}

	member, err := h.svc.Update(c.Param("id"), input)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, "Update failed", err.Error())
		return
	}
	utils.ResponseOK(c, http.StatusOK, "Member updated", member)
}

func (h *MemberHandler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Param("id")); err != nil {
		utils.ResponseError(c, http.StatusNotFound, "Delete failed", err.Error())
		return
	}
	utils.ResponseOK(c, http.StatusOK, "Member deleted", nil)
}
