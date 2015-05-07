package main

import (
	"github.com/labstack/echo"
	"log"
	"strconv"
	"time"
)

const (
	workerIdBits     = uint(5)
	datacenterIdBits = uint(5)
	sequenceBits     = uint(12)

	workerIdShift      = sequenceBits
	datacenterIdShift  = sequenceBits + workerIdBits
	timestampLeftShift = sequenceBits + workerIdBits + datacenterIdBits
)

type GetIdResponse struct {
	Id    int64  `json:"id"`
	StrId string `json:"str_id"`
}

type GoFlake struct {
	lastTimeStamp int64
}

func New() (g *GoFlake) {
	g = &GoFlake{
		lastTimeStamp: int64(-1),
	}

	return
}

func (g *GoFlake) GetTime() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func (g *GoFlake) GetId() int64 {
	sequence := int64(0)
	dcId := int64(1)
	workerId := int64(1)

	ts := g.GetTime()

	return ((ts << timestampLeftShift) | (dcId << datacenterIdShift) |
		(workerId << workerIdShift) | sequence)
}

func (g *GoFlake) GetIdResponse(c *echo.Context) {
	gid := g.GetId()
	r := &GetIdResponse{
		Id:    gid,
		StrId: strconv.Itoa(int(gid)),
	}

	c.JSON(200, r)
}

func main() {
	gf := New()
	e := echo.New()
	e.Get("/get_id.json", gf.GetIdResponse)
	e.Run(":4444")

	log.Println(gf.GetId())
}
