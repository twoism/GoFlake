package main

import (
	"flag"
	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
	"github.com/thoas/stats"
	"github.com/twoism/goflake/announcer"
	"strconv"
	"time"
)

const (
	workerIdBits     = uint(5)
	datacenterIdBits = uint(5)
	sequenceBits     = uint(12)
	sequenceMask     = -1 ^ (-1 << sequenceBits)

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

func (g *GoFlake) tilNextMillis(lastTimestamp int64) int64 {
	timestamp := g.GetTime()

	for timestamp <= g.lastTimeStamp {
		timestamp = g.GetTime()
	}
	return timestamp
}

func (g *GoFlake) GetId() int64 {
	sequence := int64(0)
	dcId := int64(1)
	workerId := int64(1)

	ts := g.GetTime()

	if g.lastTimeStamp == ts {
		sequence = (sequence + 1) & sequenceMask

		if sequence == 0 {
			ts = g.tilNextMillis(g.lastTimeStamp)
		}
	} else {
		sequence = 0
	}

	g.lastTimeStamp = ts

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

func (g *GoFlake) RunHttp(address string, port int) {
	e := echo.New()
	s := stats.New()

	e.Use(mw.Logger)
	e.Use(s.Handler)

	e.Get("/get_id.json", g.GetIdResponse)

	addr := address + ":" + strconv.Itoa(port)

	e.Run(addr)
}

func main() {
	dc := flag.String("dc", "dc1", "The service datacenter")
	service := flag.String("srv", "", "The service name")
	serviceID := flag.String("id", "", "The service ID")
	port := flag.Int("port", 4444, "The service port")
	address := flag.String("address", "127.0.0.1", "The service address")

	flag.Parse()

	cfg := announcer.Config{
		Datacenter: *dc,
		Service:    *service,
		ServiceID:  *serviceID,
		Address:    *address,
		Port:       *port,
		Tags:       []string{"goflake"},
	}

	ann := announcer.New(cfg)
	ann.Register()
	gf := New()
	gf.RunHttp(cfg.Address, cfg.Port)
}
