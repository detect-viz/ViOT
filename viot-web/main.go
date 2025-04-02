package main

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type Datacenter struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Pending int    `json:"pending"`
}

type Room struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	DCID   string `json:"dc_id"`
	Alerts int    `json:"alerts"`
}

type PDU struct {
	IP        string `json:"ip"`
	Room      string `json:"room"`
	Mfg       string `json:"mfg"`
	Model     string `json:"model"`
	SN        string `json:"sn"`
	Monitored bool   `json:"monitored"`
	Rack      string `json:"rack"`
	Side      string `json:"side"`
	IsPDU     bool   `json:"is_pdu"`
}

func main() {
	r := gin.Default()

	r.Static("/css", "./css")
	r.Static("/js", "./js")
	r.Static("/components", "./components")
	r.Static("/pages", "./pages")

	r.GET("/", func(c *gin.Context) {
		c.File("./index.html")
	})

	r.GET("/api/datacenters", func(c *gin.Context) {
		serveJSON(c, "./demo/datacenters.json")
	})

	r.GET("/api/rooms", func(c *gin.Context) {
		serveJSON(c, "./demo/rooms.json")
	})

	r.GET("/api/pdu-pending-list", func(c *gin.Context) {
		serveJSON(c, "./demo/pdu-pending.json")
	})

	r.GET("/api/pdu-list", func(c *gin.Context) {
		serveJSON(c, "./demo/pdu-grid.json")
	})

	r.GET("/api/pdu/:ip", func(c *gin.Context) {
		ip := c.Param("ip")
		path := "./demo/pdu-" + ip + ".json"
		if _, err := os.Stat(path); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		serveJSON(c, path)
	})

	r.POST("/api/bind", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "bind success"})
	})

	r.POST("/api/unbind", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "unbind success"})
	})

	r.Run(":8080")
}

func serveJSON(c *gin.Context, path string) {
	file, err := os.ReadFile(path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "file read error"})
		return
	}
	var jsonData interface{}
	if err := json.Unmarshal(file, &jsonData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid json"})
		return
	}
	c.JSON(http.StatusOK, jsonData)
}
