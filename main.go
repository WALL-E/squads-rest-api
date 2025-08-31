package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Multisig struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	MultisigAddress string    `json:"multisig_address"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Vault struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	VaultAddress    string    `json:"vault_address"`
	MultisigAddress string    `json:"multisig_address"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Member struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	MemberAddress   string    `json:"member_address"`
	Name            string    `json:"name"`
	MultisigAddress string    `json:"multisig_address"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

var db *gorm.DB

func main() {
	var err error
	db, err = gorm.Open(sqlite.Open("squads.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database")
	}

	// 自动建表
	err = db.AutoMigrate(&Multisig{}, &Vault{}, &Member{})
	if err != nil {
		log.Fatal("failed to migrate database")
	}

	r := gin.Default()

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Multisig CRUD
	r.POST("/multisigs", createMultisig)
	r.GET("/multisigs", listMultisigs)
	r.GET("/multisigs/:id", getMultisig)
	r.PUT("/multisigs/:id", updateMultisig)
	r.DELETE("/multisigs/:id", deleteMultisig)

	// Vault CRUD
	r.POST("/vaults", createVault)
	r.GET("/vaults", listVaults)
	r.GET("/vaults/:id", getVault)
	r.PUT("/vaults/:id", updateVault)
	r.DELETE("/vaults/:id", deleteVault)

	// Member CRUD
	r.POST("/members", createMember)
	r.GET("/members", listMembers)
	r.GET("/members/:id", getMember)
	r.PUT("/members/:id", updateMember)
	r.DELETE("/members/:id", deleteMember)

	// 子资源查询
	r.GET("/multisigs/:id/vaults", listVaultsByMultisig)
	r.GET("/multisigs/:id/members", listMembersByMultisig)

	r.Run(":8080")
}

//////////////////////////////
// 工具函数
//////////////////////////////

func applyQuery[T any](c *gin.Context, tx *gorm.DB) *gorm.DB {
	// 搜索 q
	if q := c.Query("q"); q != "" {
		tx = tx.Where("name LIKE ?", "%"+q+"%")
	}

	// 排序 sort=field:asc/desc
	if sort := c.Query("sort"); sort != "" {
		parts := strings.Split(sort, ":")
		field := parts[0]
		order := "asc"
		if len(parts) > 1 {
			order = parts[1]
		}
		tx = tx.Order(field + " " + order)
	}

	// 分页
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	return tx.Offset(offset).Limit(pageSize)
}

//////////////////////////////
// Multisig Handlers
//////////////////////////////

func createMultisig(c *gin.Context) {
	var input Multisig
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := db.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, input)
}

func listMultisigs(c *gin.Context) {
	var items []Multisig
	tx := applyQuery[Multisig](c, db.Model(&Multisig{}))
	if err := tx.Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func getMultisig(c *gin.Context) {
	var item Multisig
	if err := db.First(&item, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

func updateMultisig(c *gin.Context) {
	var item Multisig
	if err := db.First(&item, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := db.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

func deleteMultisig(c *gin.Context) {
	if err := db.Delete(&Multisig{}, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

//////////////////////////////
// Vault Handlers
//////////////////////////////

func createVault(c *gin.Context) {
	var input Vault
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := db.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, input)
}

func listVaults(c *gin.Context) {
	var items []Vault
	tx := applyQuery[Vault](c, db.Model(&Vault{}))
	if err := tx.Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func getVault(c *gin.Context) {
	var item Vault
	if err := db.First(&item, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

func updateVault(c *gin.Context) {
	var item Vault
	if err := db.First(&item, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := db.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

func deleteVault(c *gin.Context) {
	if err := db.Delete(&Vault{}, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

//////////////////////////////
// Member Handlers
//////////////////////////////

func createMember(c *gin.Context) {
	var input Member
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := db.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, input)
}

func listMembers(c *gin.Context) {
	var items []Member
	tx := applyQuery[Member](c, db.Model(&Member{}))
	if err := tx.Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func getMember(c *gin.Context) {
	var item Member
	if err := db.First(&item, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

func updateMember(c *gin.Context) {
	var item Member
	if err := db.First(&item, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := db.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

func deleteMember(c *gin.Context) {
	if err := db.Delete(&Member{}, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

//////////////////////////////
// 子资源
//////////////////////////////

func listVaultsByMultisig(c *gin.Context) {
	var items []Vault
	msID := c.Param("id")

	var ms Multisig
	if err := db.First(&ms, msID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "multisig not found"})
		return
	}

	if err := db.Where("multisig_address = ?", ms.MultisigAddress).Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func listMembersByMultisig(c *gin.Context) {
	var items []Member
	msID := c.Param("id")

	var ms Multisig
	if err := db.First(&ms, msID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "multisig not found"})
		return
	}

	if err := db.Where("multisig_address = ?", ms.MultisigAddress).Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}
