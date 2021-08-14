package http

import (
	"encoding/json"
	"net/http"

	"liang/internal/model"
	"liang/internal/service"

	"github.com/go-kratos/kratos/pkg/conf/paladin"
	"github.com/go-kratos/kratos/pkg/ecode"
	"github.com/go-kratos/kratos/pkg/log"
	bm "github.com/go-kratos/kratos/pkg/net/http/blademaster"
	"github.com/go-kratos/kratos/pkg/net/http/blademaster/binding"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

var svc *service.Service

// New new a bm server.
func New(s *service.Service) (engine *bm.Engine, err error) {
	var (
		cfg bm.ServerConfig
		ct  paladin.TOML
	)
	if err = paladin.Get("http.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("Server").UnmarshalTOML(&cfg); err != nil {
		return
	}
	svc = s
	engine = bm.DefaultServer(&cfg)
	initRouter(engine)
	err = engine.Start()
	return
}

func initRouter(e *bm.Engine) {
	e.Ping(ping)
	g := e.Group("/v1")
	{
		g.GET("/start", howToStart)
		g.POST("/prioritizeVerb", Prioritize)
		g.GET("/test/default", PromDemo)
		g.GET("/test/prom", RequestPromInfo)
		g.GET("/test/cache", QueryAllCache)
	}
}

func ping(ctx *bm.Context) {
	if _, err := svc.Ping(ctx, nil); err != nil {
		log.Error("ping error(%v)", err)
		ctx.AbortWithStatus(http.StatusServiceUnavailable)
	}
}

// example for http request handler.
func howToStart(c *bm.Context) {
	k := &model.Kratos{
		Hello: "Golang 大法好 !!!",
	}
	c.JSON(k, nil)
}

// Prioritize 根据Pod对Nodes评分
func Prioritize(c *bm.Context) {
	var args extenderv1.ExtenderArgs
	// BindWith will process error
	if err := c.BindWith(&args, binding.JSON); err != nil {
		return
	}

	// print args info
	jres, _ := json.Marshal(args)
	log.V(5).Info("http Prioritize api - args is: \n%s", string(jres))

	// check args nodeNames, it may be nil
	if args.NodeNames == nil {
		nodeNames := make([]string, 0, len(args.Nodes.Items))
		for _, item := range args.Nodes.Items {
			nodeNames = append(nodeNames, item.Name)
		}
		args.NodeNames = &nodeNames
	}

	res, err := svc.Prioritize(&args)
	if err != nil {
		c.JSONMap(map[string]interface{}{
			"error": err.Error(),
		}, ecode.ServerErr)
		return
	}

	if res == nil {
		res := make(extenderv1.HostPriorityList, 0, len(*args.NodeNames))
		for _, name := range *args.NodeNames {
			res = append(res, extenderv1.HostPriority{
				Host:  name,
				Score: 0,
			})
		}
	}

	// 返回评分结果
	bb, _ := json.Marshal(res)
	c.Bytes(http.StatusOK, "application/json; charset=utf-8", bb)
	return
}

func PromDemo(c *bm.Context) {
	svc.PromDemo()
	c.JSON(nil, ecode.OK)
}

func RequestPromInfo(c *bm.Context) {
	// bwType := c.Request.URL.Query().Get("bw_type")
	// if bwType == "" {
	// 	err := fmt.Errorf("bw_type should not be empty")
	// 	c.JSONMap(map[string]interface{}{
	// 		"message": err.Error(),
	// 	}, ecode.RequestErr)
	// 	return
	// }

	res, err := svc.RequestPromInfo()
	if err != nil {
		c.JSONMap(map[string]interface{}{
			"message": err.Error(),
		}, ecode.ServerErr)
		return
	}
	c.JSON(res, ecode.OK)
}

func QueryAllCache(c *bm.Context) {
	res, err := svc.GetAllCache()
	if err != nil {
		c.JSONMap(map[string]interface{}{
			"message": err.Error(),
		}, ecode.ServerErr)
		return
	}

	c.JSON(res, ecode.OK)
}
