package server

import (
	"github.com/valyala/fasthttp"
)

func requestHandler(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Path()) {
	case "/":
		writeResponse(ctx, health(), JSON)
	case "/file":
		handleFileRoute(ctx)
	default:
		ctx.Error("Unsupported path", fasthttp.StatusNotFound)
	}
}

func Start() {
	err := fasthttp.ListenAndServe(":8080", requestHandler)
	if err != nil {
		panic(err)
		return
	}
}
