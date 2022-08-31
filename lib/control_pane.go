package lib

import (
    "os"
    "strconv"
    "strings"

    "github.com/gin-gonic/gin"
)

type ControlPane struct {
    engine *gin.Engine
    raft   *RaftServer
}

func NewControlPane(raft *RaftServer) *ControlPane {
    r := &ControlPane{
        engine: gin.New(),
        raft:   raft,
    }

    r.engine.Use(gin.Recovery())
    routes := r.engine.Group("/", authRequired)
    routes.GET("/start/cluster/:cluster/mode/:mode", r.startCluster)
    routes.GET("/move/cluster/:cluster/node/:node", r.moveCluster)
    routes.GET("/add/cluster/:cluster/node/:node", r.addNode)
    routes.GET("/info", r.clusterInfo)

    return r
}

func (c *ControlPane) Run(addr string) error {
    return c.engine.Run(addr)
}

func (c *ControlPane) startCluster(g *gin.Context) {
    clusterId, err := strconv.ParseUint(g.Param("cluster"), 10, 64)
    if err != nil {
        g.JSON(500, err.Error())
        return
    }

    members := g.Query("members")
    join := strings.ToLower(g.Param("mode")) == "join"

    err = c.raft.BindCluster(members, join, clusterId)
    if err != nil {
        g.JSON(500, err.Error())
        return
    }

    g.JSON(200, true)
}

func (c *ControlPane) addNode(g *gin.Context) {
    nodeId, err := strconv.ParseUint(g.Param("node"), 10, 64)
    if err != nil {
        g.JSON(500, err.Error())
        return
    }

    clusterId, err := strconv.ParseUint(g.Param("cluster"), 10, 64)
    if err != nil {
        g.JSON(500, err.Error())
        return
    }

    nodeAddrs := g.Query("address")

    err = c.raft.AddNode(nodeId, nodeAddrs, clusterId)
    if err != nil {
        g.JSON(500, err.Error())
        return
    }

    g.JSON(200, true)
}

func (c *ControlPane) moveCluster(g *gin.Context) {
    clusterId, err := strconv.ParseUint(g.Param("cluster"), 10, 64)
    if err != nil {
        g.JSON(500, err.Error())
        return
    }

    nodeId, err := strconv.ParseUint(g.Param("node"), 10, 64)
    if err != nil {
        g.JSON(500, err.Error())
        return
    }

    err = c.raft.TransferClusters(nodeId, clusterId)
    if err != nil {
        g.JSON(500, err.Error())
        return
    }

    g.JSON(200, true)
}

func (c *ControlPane) clusterInfo(g *gin.Context) {
    cmap := c.raft.GetClusterMap()
    g.JSON(200, cmap)
}

func authRequired(c *gin.Context) {
    header := c.Request.Header.Get("Authorization")
    payload := os.Getenv("AUTH_KEY")
    if header == "" || header != payload {
        c.Header("WWW-Authenticate", "Basic")
        c.AbortWithStatus(401)
        return
    }
}