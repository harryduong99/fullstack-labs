package controller

import (
	"cuboid-challenge/app/db"
	"cuboid-challenge/app/models"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ListCuboids(c *gin.Context) {
	var cuboids []models.Cuboid
	if r := db.CONN.Find(&cuboids); r.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": r.Error.Error()})

		return
	}

	c.JSON(http.StatusOK, cuboids)
}

func GetCuboid(c *gin.Context) {
	cuboidId := c.Param("cuboidID")

	var cuboid models.Cuboid
	if r := db.CONN.First(&cuboid, cuboidId); r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Not Found"})
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": r.Error.Error()})
		}

		return
	}

	c.JSON(http.StatusOK, &cuboid)
}

func CreateCuboid(c *gin.Context) {
	var cuboidInput struct {
		Width  uint
		Height uint
		Depth  uint
		BagID  uint `json:"bagId"`
	}

	if err := c.BindJSON(&cuboidInput); err != nil {
		return
	}

	cuboid := models.Cuboid{
		Width:  cuboidInput.Width,
		Height: cuboidInput.Height,
		Depth:  cuboidInput.Depth,
		BagID:  cuboidInput.BagID,
	}

	if IsBagDisabled(cuboid) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Bag is disabled"})
		return
	}

	if !IsAvailableToAddCuboid(cuboid) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Insufficient capacity in bag"})
		return
	}

	if r := db.CONN.Create(&cuboid); r.Error != nil {
		var err models.ValidationErrors
		if ok := errors.As(r.Error, &err); ok {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": r.Error.Error()})
		}

		return
	}

	c.JSON(http.StatusCreated, &cuboid)
}

func UpdateCuboid(c *gin.Context) {
	var cuboidID = c.Param("cuboidID")
	var cuboidInput struct {
		Width  uint
		Height uint
		Depth  uint
		BagID  uint `json:"bagId"`
	}

	if err := c.BindJSON(&cuboidInput); err != nil {
		return
	}

	cuboid := models.Cuboid{
		Width:  cuboidInput.Width,
		Height: cuboidInput.Height,
		Depth:  cuboidInput.Depth,
		BagID:  cuboidInput.BagID,
	}

	// if !IsAvailableToUpdateCuboid(cuboid) {
	// 	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Insufficient capacity in bag"})
	// 	return
	// }

	if r := db.CONN.First(&cuboid, cuboidID); r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Not Found"})
			return
		}
	}

	if r := db.CONN.Where("id", cuboidID).Updates(&cuboid); r.Error != nil {
		var err models.ValidationErrors
		if ok := errors.As(r.Error, &err); ok {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": r.Error.Error()})
		}

	}

	c.JSON(http.StatusOK, &cuboid)
}

func IsAvailableToAddCuboid(cuboid models.Cuboid) bool {
	var bag models.Bag
	if r := db.CONN.Preload("Cuboids").First(&bag, cuboid.BagID); r.Error != nil {
		return false
	}

	var total uint = 0
	for _, v := range bag.Cuboids {
		total += v.PayloadVolume()
	}

	return total+cuboid.PayloadVolume() <= bag.Volume
}

func IsBagDisabled(cuboid models.Cuboid) bool {
	var bag models.Bag

	if r := db.CONN.Preload("Cuboids").First(&bag, cuboid.BagID); r.Error != nil {
		return true
	}

	return bag.Volume == 1
}

func DeleteCuboid(c *gin.Context) {
	var cuboidID = c.Param("cuboidID")

	var cuboid models.Cuboid
	if r := db.CONN.First(&cuboid, cuboidID); r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Not Found"})
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": r.Error.Error()})
		}

		return
	}

	if r := db.CONN.Delete(&cuboid); r.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": r.Error.Error()})

		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}
