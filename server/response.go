package server

import (
	"github.com/valyala/fasthttp"
)

type dataTypes string

const (
	JSON dataTypes = "json"
	FILE dataTypes = "file"
)

func writeResponse(ctx *fasthttp.RequestCtx, data []byte, dataType dataTypes) {
	switch dataType {
	case "json":
		ctx.Response.Header.Set("Content-Type", "application/json")
	default:
		ctx.Response.Header.Set("Content-Type", "text/plain")
	}
	ctx.Response.SetBody(data)
}
