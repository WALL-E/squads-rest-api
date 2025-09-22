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

// Build information (set during build time)
var (
	BuildTime = "unknown"
	GitCommit = "unknown"
	Version   = "dev"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	BuildTime string `json:"build_time"`
	GitCommit string `json:"git_commit"`
	Version   string `json:"version"`
}

// ---------- Models ----------
type Multisig struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	MultisigAddress string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"multisig_address"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Logo            string    `json:"logo"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type Vault struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	VaultAddress    string    `gorm:"type:varchar(255);uniqueIndex:idx_multisig_vault;not null" json:"vault_address"`
	MultisigAddress string    `gorm:"type:varchar(255);uniqueIndex:idx_multisig_vault;not null" json:"multisig_address"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type Member struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	MemberAddress   string    `gorm:"type:varchar(255);uniqueIndex:idx_multisig_member;not null" json:"member_address"`
	MultisigAddress string    `gorm:"type:varchar(255);uniqueIndex:idx_multisig_member;not null" json:"multisig_address"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type Spend struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	SpendAddress    string    `gorm:"type:varchar(255);uniqueIndex:idx_multisig_spend;not null" json:"spend_address"`
	MultisigAddress string    `gorm:"type:varchar(255);uniqueIndex:idx_multisig_spend;not null" json:"multisig_address"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type Transaction struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	IndexNumber     uint      `gorm:"not null" json:"indexNumber"`
	Signature       string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"signature"`
	MultisigAddress string    `gorm:"type:varchar(255);not null;index" json:"multisig_address"`
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
	dsn := "root:123456@tcp(127.0.0.1:3306)/rwa?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}

	// Auto migrate the schema
	err = db.AutoMigrate(&Multisig{}, &Vault{}, &Member{}, &Spend{}, &Transaction{})
	if err != nil {
		log.Fatal("failed to migrate database: ", err)
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
	ms.Logo = input.Logo
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
		if err := tx.Where("multisig_address = ?", addr).Delete(&Transaction{}).Error; err != nil {
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
// @Param vault_address path string true "Vault Address"
// @Success 200 {object} Response
// @Router /multisigs/{multisig_address}/vaults/{vault_address} [get]
func getVault(c *gin.Context) {
	addr := c.Param("multisig_address")
	vaultAddr := c.Param("vault_address")
	var item Vault
	if err := db.Where("multisig_address = ? AND vault_address = ?", addr, vaultAddr).First(&item).Error; err != nil {
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
// @Description Update a vault under a multisig
// @Tags Vaults
// @Param multisig_address path string true "Multisig Address"
// @Param vault_address path string true "Vault Address"
// @Param body body Vault true "Vault object"
// @Success 200 {object} Response
// @Router /multisigs/{multisig_address}/vaults/{vault_address} [put]
func updateVault(c *gin.Context) {
	addr := c.Param("multisig_address")
	vaultAddr := c.Param("vault_address")
	var item Vault
	if err := db.Where("multisig_address = ? AND vault_address = ?", addr, vaultAddr).First(&item).Error; err != nil {
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
// @Description Delete a vault under a multisig
// @Tags Vaults
// @Param multisig_address path string true "Multisig Address"
// @Param vault_address path string true "Vault Address"
// @Success 200 {object} Response
// @Router /multisigs/{multisig_address}/vaults/{vault_address} [delete]
func deleteVault(c *gin.Context) {
	addr := c.Param("multisig_address")
	vaultAddr := c.Param("vault_address")
	if err := db.Where("multisig_address = ? AND vault_address = ?", addr, vaultAddr).Delete(&Vault{}).Error; err != nil {
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
// @Param member_address path string true "Member Address"
// @Success 200 {object} Response
// @Router /multisigs/{multisig_address}/members/{member_address} [get]
func getMember(c *gin.Context) {
	addr := c.Param("multisig_address")
	memberAddr := c.Param("member_address")
	var item Member
	if err := db.Where("multisig_address = ? AND member_address = ?", addr, memberAddr).First(&item).Error; err != nil {
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
// @Description Update a member under a multisig
// @Tags Members
// @Param multisig_address path string true "Multisig Address"
// @Param member_address path string true "Member Address"
// @Param body body Member true "Updated member object"
// @Success 200 {object} Response
// @Router /multisigs/{multisig_address}/members/{member_address} [put]
func updateMember(c *gin.Context) {
	addr := c.Param("multisig_address")
	memberAddr := c.Param("member_address")
	var item Member
	if err := db.Where("multisig_address = ? AND member_address = ?", addr, memberAddr).First(&item).Error; err != nil {
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
// @Description Delete a member under a multisig
// @Tags Members
// @Param multisig_address path string true "Multisig Address"
// @Param member_address path string true "Member Address"
// @Success 200 {object} Response
// @Router /multisigs/{multisig_address}/members/{member_address} [delete]
func deleteMember(c *gin.Context) {
	addr := c.Param("multisig_address")
	memberAddr := c.Param("member_address")
	if err := db.Where("multisig_address = ? AND member_address = ?", addr, memberAddr).Delete(&Member{}).Error; err != nil {
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
// @Param spend_address path string true "Spend Address"
// @Success 200 {object} Response
// @Router /multisigs/{multisig_address}/spends/{spend_address} [get]
func getSpend(c *gin.Context) {
	addr := c.Param("multisig_address")
	spendAddr := c.Param("spend_address")
	var item Spend
	if err := db.Where("multisig_address = ? AND spend_address = ?", addr, spendAddr).First(&item).Error; err != nil {
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
// @Description Update a spend under a multisig
// @Tags Spends
// @Param multisig_address path string true "Multisig Address"
// @Param spend_address path string true "Spend Address"
// @Param body body Spend true "Updated spend object"
// @Success 200 {object} Response
// @Router /multisigs/{multisig_address}/spends/{spend_address} [put]
func updateSpend(c *gin.Context) {
	addr := c.Param("multisig_address")
	spendAddr := c.Param("spend_address")
	var item Spend
	if err := db.Where("multisig_address = ? AND spend_address = ?", addr, spendAddr).First(&item).Error; err != nil {
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
// @Description Delete a spend under a multisig
// @Tags Spends
// @Param multisig_address path string true "Multisig Address"
// @Param spend_address path string true "Spend Address"
// @Success 200 {object} Response
// @Router /multisigs/{multisig_address}/spends/{spend_address} [delete]
func deleteSpend(c *gin.Context) {
	addr := c.Param("multisig_address")
	spendAddr := c.Param("spend_address")
	if err := db.Where("multisig_address = ? AND spend_address = ?", addr, spendAddr).Delete(&Spend{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, Response{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, Response{Success: true})
}

// ==================== Transaction Handlers ====================

// @Summary List Transactions
// @Description Get transactions list of a multisig with pagination, search, sort
// @Tags Transactions
// @Param multisig_address path string true "Multisig Address"
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Param search query string false "Search keyword"
// @Param sort query string false "Sort field and order, e.g., 'id desc'"
// @Success 200 {object} Response
// @Router /multisigs/{multisig_address}/transactions [get]
func listTransactions(c *gin.Context) {
	addr := c.Param("multisig_address")
	var items []Transaction
	tx := applyQuery(c, db.Model(&Transaction{}).Where("multisig_address = ?", addr), []string{"signature"}, []string{"id", "index_number", "created_at", "updated_at"})
	result, err := paginate(c, tx, &items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, Response{Success: true, Data: result})
}

// @Summary Get Transaction
// @Description Get a single transaction by index number
// @Tags Transactions
// @Param multisig_address path string true "Multisig Address"
// @Param indexNumber path int true "Transaction Index Number"
// @Success 200 {object} Response
// @Router /multisigs/{multisig_address}/transactions/{indexNumber} [get]
func getTransaction(c *gin.Context) {
	addr := c.Param("multisig_address")
	indexNumber, ok := parseID(c, "indexNumber")
	if !ok {
		return
	}
	var item Transaction
	if err := db.Where("multisig_address = ? AND index_number = ?", addr, indexNumber).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, Response{Success: false, Message: "not found"})
		return
	}
	c.JSON(http.StatusOK, Response{Success: true, Data: item})
}

// @Summary Create Transaction
// @Description Create a new transaction under a multisig
// @Tags Transactions
// @Param multisig_address path string true "Multisig Address"
// @Param body body Transaction true "Transaction object"
// @Success 201 {object} Response
// @Router /multisigs/{multisig_address}/transactions [post]
func createTransaction(c *gin.Context) {
	addr := c.Param("multisig_address")
	var item Transaction
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

// @Summary Update Transaction
// @Description Update an existing transaction by index number
// @Tags Transactions
// @Param multisig_address path string true "Multisig Address"
// @Param indexNumber path int true "Transaction Index Number"
// @Param body body Transaction true "Updated transaction object"
// @Success 200 {object} Response
// @Router /multisigs/{multisig_address}/transactions/{indexNumber} [put]
func updateTransaction(c *gin.Context) {
	addr := c.Param("multisig_address")
	indexNumber, ok := parseID(c, "indexNumber")
	if !ok {
		return
	}
	var tx Transaction
	if err := db.Where("multisig_address = ? AND index_number = ?", addr, indexNumber).First(&tx).Error; err != nil {
		c.JSON(http.StatusNotFound, Response{Success: false, Message: "not found"})
		return
	}
	var input Transaction
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, Response{Success: false, Message: err.Error()})
		return
	}
	tx.IndexNumber = input.IndexNumber
	if err := db.Save(&tx).Error; err != nil {
		c.JSON(http.StatusInternalServerError, Response{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, Response{Success: true, Data: tx})
}

// @Summary Delete Transaction
// @Description Delete a transaction under a multisig
// @Tags Transactions
// @Param multisig_address path string true "Multisig Address"
// @Param indexNumber path int true "Transaction Index Number"
// @Success 200 {object} Response
// @Router /multisigs/{multisig_address}/transactions/{indexNumber} [delete]
func deleteTransaction(c *gin.Context) {
	addr := c.Param("multisig_address")
	indexNumber, ok := parseID(c, "indexNumber")
	if !ok {
		return
	}
	if err := db.Where("multisig_address = ? AND index_number = ?", addr, indexNumber).Delete(&Transaction{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, Response{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, Response{Success: true})
}

// @Summary Health Check
// @Description Get service health status with build information
// @Tags Health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Success:   true,
		Message:   "ok",
		BuildTime: BuildTime,
		GitCommit: GitCommit,
		Version:   Version,
	})
}

// @Summary API Documentation
// @Description Get API documentation homepage with links to Swagger UI
// @Tags Documentation
// @Produce html
// @Success 200 {string} string "HTML content"
// @Router / [get]
func apiDocumentation(c *gin.Context) {
	html := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Squads REST API 文档</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            line-height: 1.6;
            margin: 0;
            padding: 40px;
            background-color: #f5f5f5;
        }
        .container {
            max-width: 800px;
            margin: 0 auto;
            background: white;
            padding: 40px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
            border-bottom: 3px solid #007cba;
            padding-bottom: 10px;
        }
        h2 {
            color: #555;
            margin-top: 30px;
        }
        .api-link {
            display: inline-block;
            background: #007cba;
            color: white;
            padding: 12px 24px;
            text-decoration: none;
            border-radius: 5px;
            margin: 10px 10px 10px 0;
            transition: background-color 0.3s;
        }
        .api-link:hover {
            background: #005a87;
        }
        .description {
            background: #f8f9fa;
            padding: 20px;
            border-left: 4px solid #007cba;
            margin: 20px 0;
        }
        .endpoint {
            background: #f1f3f4;
            padding: 10px;
            border-radius: 4px;
            font-family: monospace;
            margin: 5px 0;
        }
        .version-info {
            background: #e8f5e8;
            padding: 15px;
            border-radius: 5px;
            margin-top: 30px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 Squads REST API</h1>
        
        <div class="description">
            <p><strong>欢迎使用 Squads REST API！</strong></p>
            <p>这是一个用于管理多签钱包、金库、成员、支出和交易的 RESTful API 服务。</p>
        </div>

        <h2>📚 API 文档</h2>
         <a href="/swagger/index.html" class="api-link">📖 Swagger UI 文档</a>
        <a href="/health" class="api-link">💚 健康检查</a>

        <h2>🔗 主要接口</h2>
        <div class="endpoint">GET /multisigs - 获取多签钱包列表</div>
        <div class="endpoint">POST /multisigs - 创建多签钱包</div>
        <div class="endpoint">GET /multisigs/{address}/vaults - 获取金库列表</div>
        <div class="endpoint">GET /multisigs/{address}/members - 获取成员列表</div>
        <div class="endpoint">GET /multisigs/{address}/spends - 获取支出列表</div>
        <div class="endpoint">GET /multisigs/{address}/transactions - 获取交易列表</div>

        <h2>🛠️ 技术栈</h2>
        <ul>
            <li><strong>框架:</strong> Gin (Go)</li>
            <li><strong>数据库:</strong> MySQL + GORM</li>
            <li><strong>文档:</strong> Swagger/OpenAPI 3.0</li>
            <li><strong>部署:</strong> Docker + Nginx</li>
        </ul>

        <div class="version-info">
            <h3>📋 服务信息</h3>
            <p><strong>服务地址:</strong> http://localhost:8090</p>
            <p><strong>API 版本:</strong> v1</p>
            <p><strong>文档更新:</strong> 2025年</p>
        </div>

        <h2>🚀 快速开始</h2>
        <p>1. 查看 <a href="/swagger/index.html">Swagger 文档</a> 了解详细的 API 接口</p>
        <p>2. 使用 <a href="/health">健康检查接口</a> 验证服务状态</p>
        <p>3. 开始调用 API 接口进行开发</p>
    </div>
</body>
</html>`
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// ---------- Router ----------
func main() {
	initDB()
	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.Group("/")
	{
		// API Documentation
		v1.GET("/", apiDocumentation)

		// Health Check
		v1.GET("/health", healthCheck)

		// Multisigs
		v1.GET("/multisigs", listMultisigs)
		v1.GET("/multisigs/:multisig_address", getMultisig)
		v1.POST("/multisigs", createMultisig)
		v1.PUT("/multisigs/:multisig_address", updateMultisig)
		v1.DELETE("/multisigs/:multisig_address", deleteMultisig)

		// Vaults
		v1.GET("/multisigs/:multisig_address/vaults", listVaults)
		v1.GET("/multisigs/:multisig_address/vaults/:vault_address", getVault)
		v1.POST("/multisigs/:multisig_address/vaults", createVault)
		v1.PUT("/multisigs/:multisig_address/vaults/:vault_address", updateVault)
		v1.DELETE("/multisigs/:multisig_address/vaults/:vault_address", deleteVault)

		// Members
		v1.GET("/multisigs/:multisig_address/members", listMembers)
		v1.GET("/multisigs/:multisig_address/members/:member_address", getMember)
		v1.POST("/multisigs/:multisig_address/members", createMember)
		v1.PUT("/multisigs/:multisig_address/members/:member_address", updateMember)
		v1.DELETE("/multisigs/:multisig_address/members/:member_address", deleteMember)

		// Spends
		v1.GET("/multisigs/:multisig_address/spends", listSpends)
		v1.GET("/multisigs/:multisig_address/spends/:spend_address", getSpend)
		v1.POST("/multisigs/:multisig_address/spends", createSpend)
		v1.PUT("/multisigs/:multisig_address/spends/:spend_address", updateSpend)
		v1.DELETE("/multisigs/:multisig_address/spends/:spend_address", deleteSpend)

		// Transactions
		v1.GET("/multisigs/:multisig_address/transactions", listTransactions)
		v1.GET("/multisigs/:multisig_address/transactions/:indexNumber", getTransaction)
		v1.POST("/multisigs/:multisig_address/transactions", createTransaction)
		v1.PUT("/multisigs/:multisig_address/transactions/:indexNumber", updateTransaction)
		v1.DELETE("/multisigs/:multisig_address/transactions/:indexNumber", deleteTransaction)
	}

	r.Run(":8090")
}
