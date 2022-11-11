package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/i5heu/simple-S3-cache/config"
	"github.com/i5heu/simple-S3-cache/ramCache"
	"github.com/valyala/fasthttp"
)

type Handler struct {
	conf      config.Config
	dataStore *ramCache.DataStore
}

func main() {

	conf := config.GetValues()
	dataStore := ramCache.DataStore{
		Conf: conf,
		Ch:   make(chan ramCache.File, 10000),
	}
	go dataStore.RamFileManager()

	h := Handler{conf: conf, dataStore: &dataStore}
	fasthttp.ListenAndServe(":8084", h.handler)
}

func (h *Handler) handler(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Access-Control-Allow-Origin", h.conf.CORSDomain)
	ctx.Response.Header.Set("Access-Control-Allow-Methods", "GET")

	url := config.GetCompleteURL(h.conf, string(ctx.Path()))
	cachedData := h.dataStore.GetCacheData(url)
	if cachedData != nil {
		ctx.Response.SetBody(cachedData)
		return
	}

	res, err := http.Get(url)
	if err != nil {
		fmt.Printf("error making http request: %s\n", err)
		ctx.Response.SetStatusCode(500)
		return
	}

	defer res.Body.Close()

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	h.dataStore.CacheData(url, bytes)
	ctx.Response.SetBody(bytes)
}
