package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	_ "github.com/wall-e/squads-rest-api/docs" // swagger docs
)

// ---------- Models ----------
type Multisig struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	MultisigAddress string    `gorm:"uniqueIndex;not null" json:"multisig_address"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type Vault struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	VaultAddress    string    `gorm:"uniqueIndex;not null" json:"vault_address"`
	MultisigAddress string    `gorm:"index;not null" json:"multisig_address"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type Member struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	MemberAddress   string    `gorm:"uniqueIndex;not null" json:"member_address"`
	MultisigAddress string    `gorm:"index;not null" json:"multisig_address"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type Spend struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	SpendAddress    string    `gorm:"uniqueIndex;not null" json:"spend_address"`
	MultisigAddress string    `gorm:"index;not null" json:"multisig_address"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// ---------- Response ----------
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type PageResult struct {
	Total int64       `json:"total"`
	Items interface{} `json:"items"`
}

var db *gorm.DB

// ---------- DB Init ----------
func initDB() {
	dsn := "root:123456@tcp(127.0.0.1:3306)/squads?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}
	if err := db.AutoMigrate(&Multisig{}, &Vault{}, &Member{}, &Spend{}); err != nil {
		log.Fatal("migration failed: ", err)
	}
}

// ---------- Helpers ----------
func parseID(c *gin.Context, key string) (int, bool) {
	idStr := c.Param(key)
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, Response{Success: false, Message: "invalid id"})
		return 0, false
	}
	return id, true
}

func paginate(c *gin.Context, tx *gorm.DB, out interface{}) (PageResult, error) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	var total int64
	if err := tx.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return PageResult{}, err
	}

	if err := tx.Offset((page - 1) * limit).Limit(limit).Find(out).Error; err != nil {
		return PageResult{}, err
	}
	return PageResult{Total: total, Items: out}, nil
}

func applyQuery(c *gin.Context, tx *gorm.DB, searchableFields []string, sortableFields []string) *gorm.DB {
	search := c.Query("search")
	if search != "" {
		like := "%" + search + "%"
		for i, field := range searchableFields {
			if i == 0 {
				tx = tx.Where(field+" LIKE ?", like)
			} else {
				tx = tx.Or(field+" LIKE ?", like)
			}
		}
	}
	sort := c.DefaultQuery("sort", "id desc")
	parts := strings.Fields(sort)
	field, order := "id", "desc"
	if len(parts) == 2 {
		field, order = parts[0], strings.ToLower(parts[1])
	}
	allowed := map[string]bool{}
	for _, f := range sortableFields {
		allowed[f] = true
	}
	if allowed[field] && (order == "asc" || order == "desc") {
		tx = tx.Order(field + " " + order)
	}
	return tx
}

// ==================== Multisig Handlers ====================

// @Summary List Multisigs
// @Description Get multisigs list with pagination, search, and sort
// @Tags Multisigs
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Param search query string false "Search keyword"
// @Param sort query string false "Sort field and order, e.g., 'id desc'"
// @Success 200 {object} Response
// @Router /multisigs [get]
func listMultisigs(c *gin.Context) {
	var items []Multisig
	tx := applyQuery(c, db.Model(&Multisig{}), []string{"name", "description"}, []string{"id", "name", "created_at", "updated_at"})
	result, err := paginate(c, tx, &items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, Response{Success: true, Data: result})
}

// @Summary Get Multisig
// @Description Get a single multisig by address
// @Tags Multisigs
// @Param multisig_address path string true "Multisig Address"
// @Success 200 {object} Response
// @Router /multisigs/{multisig_address} [get]
func getMultisig(c *gin.Context) {
	addr := c.Param("multisig_address")
	var item Multisig
	if err := db.Where("multisig_address = ?", addr).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, Response{Success: false, Message: "not found"})
		return
	}
	c.JSON(http.StatusOK, Response{Success: true, Data: item})
}

// @Summary Create Multisig
// @Description Create a new multisig
// @Tags Multisigs
// @Param body body Multisig true "Multisig object"
// @Success 201 {object} Response
// @Router /multisigs [post]
func createMultisig(c *gin.Context) {
	var item Multisig
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, Response{Success: false, Message: err.Error()})
		return
	}
	if err := db.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, Response{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, Response{Success: true, Data: item})
}

// @Summary Update Multisig
// @Description Update an existing multisig by address
// @Tags Multisigs
// @Param multisig_address path string true "Multisig Address"
// @Param body body Multisig true "Updated multisig object"
// @Success 200 {object} Response
// @Router /multisigs/{multisig_address} [put]
func updateMultisig(c *gin.Context) {
	addr := c.Param("multisig_address")
	var ms Multisig
	if err := db.Where("multisig_address = ?", addr).First(&ms).Error; err != nil {
		c.JSON(http.StatusNotFound, Response{Success: false, Message: "not found"})
		return
	}
	var input Multisig
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, Response{Success: false, Message: err.Error()})
		return
	}
	ms.Name = input.Name
	ms.Description = input.Description
	if err := db.Save(&ms).Error; err != nil {
		c.JSON(http.StatusInternalServerError, Response{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, Response{Success: true, Data: ms})
}

// @Summary Delete Multisig
// @Description Delete a multisig and all its related Vaults, Members, and Spends
// @Tags Multisigs
// @Param multisig_address path string true "Multisig Address"
// @Success 200 {object} Response
// @Router /multisigs/{multisig_address} [delete]
func deleteMultisig(c *gin.Context) {
	addr := c.Param("multisig_address")
	var ms Multisig
	if err := db.Where("multisig_address = ?", addr).First(&ms).Error; err != nil {
		c.JSON(http.StatusNotFound, Response{Success: false, Message: "not found"})
		return
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("multisig_address = ?", addr).Delete(&Vault{}).Error; err != nil {
			return err
		}
		if err := tx.Where("multisig_address = ?", addr).Delete(&Member{}).Error; err != nil {
			return err
		}
		if err := tx.Where("multisig_address = ?", addr).Delete(&Spend{}).Error; err != nil {
			return err
		}
		return tx.Delete(&ms).Error
	}); err != nil {
		c.JSON(http.StatusInternalServerError, Response{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, Response{Success: true})
}

// ==================== Vault Handlers ====================

// @Summary List Vaults
// @Description Get vaults list of a multisig with pagination, search, sort
// @Tags Vaults
// @Param multisig_address path string true "Multisig Address"
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Param search query string false "Search keyword"
// @Param sort query string false "Sort field and order, e.g., 'id desc'"
// @Success 200 {object} Response
// @Router /multisigs/{multisig_address}/vaults [get]
func listVaults(c *gin.Context) {
	addr := c.Param("multisig_address")
	var items []Vault
	tx := applyQuery(c, db.Model(&Vault{}).Where("multisig_address = ?", addr), []string{"name", "description"}, []string{"id", "name", "created_at", "updated_at"})
	result, err := paginate(c, tx, &items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, Response{Success: true, Data: result})
}

// @Summary Get Vault
// @Description Get a single vault by ID
// @Tags Vaults
// @Param multisig_address path string true "Multisig Address"
// @Param vault_id path int true "Vault ID"
// @Success 200 {object} Response
// @Router /multisigs/{multisig_address}/vaults/{vault_id} [get]
func getVault(c *gin.Context) {
	addr := c.Param("multisig_address")
	id, ok := parseID(c, "vault_id")
	if !ok {
		return
	}
	var item Vault
	if err := db.Where("multisig_address = ? AND id = ?", addr, id).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, Response{Success: false, Message: "not found"})
		return
	}
	c.JSON(http.StatusOK, Response{Success: true, Data: item})
}

// @Summary Create Vault
// @Description Create a new vault under a multisig
// @Tags Vaults
// @Param multisig_address path string true "Multisig Address"
// @Param body body Vault true "Vault object"
// @Success 201 {object} Response
// @Router /multisigs/{multisig_address}/vaults [post]
func createVault(c *gin.Context) {
	addr := c.Param("multisig_address")
	var item Vault
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, Response{Success: false, Message: err.Error()})
		return
	}
	item.MultisigAddress = addr
	if err := db.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, Response{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, Response{Success: true, Data: item})
}

// @Summary Update Vault
// @Description Update a vault
// @Tags Vaults
// @Param multisig_address path string true "Multisig Address"
// @Param vault_id path int true "Vault ID"
// @Param body body Vault true "Updated vault object"
// @Success 200 {object} Response
// @Router /multisigs/{multisig_address}/vaults/{vault_id} [put]
func updateVault(c *gin.Context) {
	addr := c.Param("multisig_address")
	id, ok := parseID(c, "vault_id")
	if !ok {
		return
	}
	var item Vault
	if err := db.Where("multisig_address = ? AND id = ?", addr, id).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, Response{Success: false, Message: "not found"})
		return
	}
	var input Vault
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, Response{Success: false, Message: err.Error()})
		return
	}
	item.Name = input.Name
	item.Description = input.Description
	if err := db.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, Response{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, Response{Success: true, Data: item})
}

// @Summary Delete Vault
// @Description Delete a vault
// @Tags Vaults
// @Param multisig_address path string true "Multisig Address"
// @Param vault_id path int true "Vault ID"
// @Success 200 {object} Response
// @Router /multisigs/{multisig_address}/vaults/{vault_id} [delete]
func deleteVault(c *gin.Context) {
	addr := c.Param("multisig_address")
	id, ok := parseID(c, "vault_id")
	if !ok {
		return
	}
	if err := db.Where("multisig_address = ? AND id = ?", addr, id).Delete(&Vault{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, Response{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, Response{Success: true})
}

// ==================== Member Handlers ====================

// @Summary List Members
// @Description Get members list of a multisig with pagination, search, sort
// @Tags Members
// @Param multisig_address path string true "Multisig Address"
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Param search query string false "Search keyword"
// @Param sort query string false "Sort field and order, e.g., 'id desc'"
// @Success 200 {object} Response
// @Router /multisigs/{multisig_address}/members [get]
func listMembers(c *gin.Context) {
	addr := c.Param("multisig_address")
	var items []Member
	tx := applyQuery(c, db.Model(&Member{}).Where("multisig_address = ?", addr), []string{"name", "description"}, []string{"id", "name", "created_at", "updated_at"})
	result, err := paginate(c, tx, &items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, Response{Success: true, Data: result})
}

// @Summary Get Member
// @Description Get a single member by ID
// @Tags Members
// @Param multisig_address path string true "Multisig Address"
// @Param member_id path int true "Member ID"
// @Success 200 {object} Response
// @Router /multisigs/{multisig_address}/members/{member_id} [get]
func getMember(c *gin.Context) {
	addr := c.Param("multisig_address")
	id, ok := parseID(c, "member_id")
	if !ok {
		return
	}
	var item Member
	if err := db.Where("multisig_address = ? AND id = ?", addr, id).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, Response{Success: false, Message: "not found"})
		return
	}
	c.JSON(http.StatusOK, Response{Success: true, Data: item})
}

// @Summary Create Member
// @Description Create a new member under a multisig
// @Tags Members
// @Param multisig_address path string true "Multisig Address"
// @Param body body Member true "Member object"
// @Success 201 {object} Response
// @Router /multisigs/{multisig_address}/members [post]
func createMember(c *gin.Context) {
	addr := c.Param("multisig_address")
	var item Member
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, Response{Success: false, Message: err.Error()})
		return
	}
	item.MultisigAddress = addr
	if err := db.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, Response{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, Response{Success: true, Data: item})
}

// @Summary Update Member
// @Description Update a member
// @Tags Members
// @Param multisig_address path string true "Multisig Address"
// @Param member_id path int true "Member ID"
// @Param body body Member true "Updated member object"
// @Success 200 {object} Response
// @Router /multisigs/{multisig_address}/members/{member_id} [put]
func updateMember(c *gin.Context) {
	addr := c.Param("multisig_address")
	id, ok := parseID(c, "member_id")
	if !ok {
		return
	}
	var item Member
	if err := db.Where("multisig_address = ? AND id = ?", addr, id).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, Response{Success: false, Message: "not found"})
		return
	}
	var input Member
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, Response{Success: false, Message: err.Error()})
		return
	}
	item.Name = input.Name
	item.Description = input.Description
	if err := db.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, Response{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, Response{Success: true, Data: item})
}

// @Summary Delete Member
// @Description Delete a member
// @Tags Members
// @Param multisig_address path string true "Multisig Address"
// @Param member_id path int true "Member ID"
// @Success 200 {object} Response
// @Router /multisigs/{multisig_address}/members/{member_id} [delete]
func deleteMember(c *gin.Context) {
	addr := c.Param("multisig_address")
	id, ok := parseID(c, "member_id")
	if !ok {
		return
	}
	if err := db.Where("multisig_address = ? AND id = ?", addr, id).Delete(&Member{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, Response{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, Response{Success: true})
}

// ==================== Spend Handlers ====================

// @Summary List Spends
// @Description Get spends list of a multisig with pagination, search, sort
// @Tags Spends
// @Param multisig_address path string true "Multisig Address"
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Param search query string false "Search keyword"
// @Param sort query string false "Sort field and order, e.g., 'id desc'"
// @Success 200 {object} Response
// @Router /multisigs/{multisig_address}/spends [get]
func listSpends(c *gin.Context) {
	addr := c.Param("multisig_address")
	var items []Spend
	tx := applyQuery(c, db.Model(&Spend{}).Where("multisig_address = ?", addr), []string{"name", "description"}, []string{"id", "name", "created_at", "updated_at"})
	result, err := paginate(c, tx, &items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, Response{Success: true, Data: result})
}

// @Summary Get Spend
// @Description Get a single spend by ID
// @Tags Spends
// @Param multisig_address path string true "Multisig Address"
// @Param spend_id path int true "Spend ID"
// @Success 200 {object} Response
// @Router /multisigs/{multisig_address}/spends/{spend_id} [get]
func getSpend(c *gin.Context) {
	addr := c.Param("multisig_address")
	id, ok := parseID(c, "spend_id")
	if !ok {
		return
	}
	var item Spend
	if err := db.Where("multisig_address = ? AND id = ?", addr, id).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, Response{Success: false, Message: "not found"})
		return
	}
	c.JSON(http.StatusOK, Response{Success: true, Data: item})
}

// @Summary Create Spend
// @Description Create a new spend under a multisig
// @Tags Spends
// @Param multisig_address path string true "Multisig Address"
// @Param body body Spend true "Spend object"
// @Success 201 {object} Response
// @Router /multisigs/{multisig_address}/spends [post]
func createSpend(c *gin.Context) {
	addr := c.Param("multisig_address")
	var item Spend
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, Response{Success: false, Message: err.Error()})
		return
	}
	item.MultisigAddress = addr
	if err := db.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, Response{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, Response{Success: true, Data: item})
}

// @Summary Update Spend
// @Description Update a spend
// @Tags Spends
// @Param multisig_address path string true "Multisig Address"
// @Param spend_id path int true "Spend ID"
// @Param body body Spend true "Updated spend object"
// @Success 200 {object} Response
// @Router /multisigs/{multisig_address}/spends/{spend_id} [put]
func updateSpend(c *gin.Context) {
	addr := c.Param("multisig_address")
	id, ok := parseID(c, "spend_id")
	if !ok {
		return
	}
	var item Spend
	if err := db.Where("multisig_address = ? AND id = ?", addr, id).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, Response{Success: false, Message: "not found"})
		return
	}
	var input Spend
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, Response{Success: false, Message: err.Error()})
		return
	}
	item.Name = input.Name
	item.Description = input.Description
	if err := db.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, Response{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, Response{Success: true, Data: item})
}

// @Summary Delete Spend
// @Description Delete a spend
// @Tags Spends
// @Param multisig_address path string true "Multisig Address"
// @Param spend_id path int true "Spend ID"
// @Success 200 {object} Response
// @Router /multisigs/{multisig_address}/spends/{spend_id} [delete]
func deleteSpend(c *gin.Context) {
	addr := c.Param("multisig_address")
	id, ok := parseID(c, "spend_id")
	if !ok {
		return
	}
	if err := db.Where("multisig_address = ? AND id = ?", addr, id).Delete(&Spend{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, Response{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, Response{Success: true})
}

// ---------- Router ----------
func main() {
	initDB()
	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.Group("/")
	{
		// Multisigs
		v1.GET("/multisigs", listMultisigs)
		v1.GET("/multisigs/:multisig_address", getMultisig)
		v1.POST("/multisigs", createMultisig)
		v1.PUT("/multisigs/:multisig_address", updateMultisig)
		v1.DELETE("/multisigs/:multisig_address", deleteMultisig)

		// Vaults
		v1.GET("/multisigs/:multisig_address/vaults", listVaults)
		v1.GET("/multisigs/:multisig_address/vaults/:vault_id", getVault)
		v1.POST("/multisigs/:multisig_address/vaults", createVault)
		v1.PUT("/multisigs/:multisig_address/vaults/:vault_id", updateVault)
		v1.DELETE("/multisigs/:multisig_address/vaults/:vault_id", deleteVault)

		// Members
		v1.GET("/multisigs/:multisig_address/members", listMembers)
		v1.GET("/multisigs/:multisig_address/members/:member_id", getMember)
		v1.POST("/multisigs/:multisig_address/members", createMember)
		v1.PUT("/multisigs/:multisig_address/members/:member_id", updateMember)
		v1.DELETE("/multisigs/:multisig_address/members/:member_id", deleteMember)

		// Spends
		v1.GET("/multisigs/:multisig_address/spends", listSpends)
		v1.GET("/multisigs/:multisig_address/spends/:spend_id", getSpend)
		v1.POST("/multisigs/:multisig_address/spends", createSpend)
		v1.PUT("/multisigs/:multisig_address/spends/:spend_id", updateSpend)
		v1.DELETE("/multisigs/:multisig_address/spends/:spend_id", deleteSpend)
	}

	r.Run(":8090")
}
